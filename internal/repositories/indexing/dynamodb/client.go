package dynamodb

import (
	"context"
	"errors"
	"iconrepo/internal/config"
	"iconrepo/internal/repositories/indexing"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	aws_dyndb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog"
)

func createAwsConfig(conf *config.Options) (aws.Config, error) {

	loadOptions := []func(*aws_config.LoadOptions) error{}

	if len(conf.DynamodbURL) > 1 {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           conf.DynamodbURL,
				SigningRegion: region,
			}, nil
		})
		loadOptions = append(loadOptions, aws_config.WithEndpointResolverWithOptions(customResolver))
	}

	if profile := os.Getenv("AWS_PROFILE"); len(profile) > 0 {
		loadOptions = append(loadOptions, aws_config.WithSharedConfigProfile(profile))
	}

	return aws_config.LoadDefaultConfig(
		context.TODO(),
		loadOptions...,
	)
}

func NewDynamodbClient(conf *config.Options) (*aws_dyndb.Client, error) {
	awsConf, err := createAwsConfig(conf)
	if err != nil {
		return nil, err
	}
	svc := aws_dyndb.NewFromConfig(awsConf)
	return svc, nil
}

func Unwrap(ctx context.Context, err error) error {
	logger := zerolog.Ctx(ctx)
	var oe *smithy.OperationError
	if errors.As(err, &oe) {
		logger.Error().Str("service", oe.Service()).Str("operation", oe.Operation()).Err(oe.Unwrap()).Msgf("failed to call service: %s", oe.Service())

		tmpErr := &types.ResourceNotFoundException{}
		if errors.As(oe.Err, &tmpErr) {
			return indexing.ErrTableNotFound
		}
	}
	logger.Error().Err(err).Msg("some error occured")
	return err
}
