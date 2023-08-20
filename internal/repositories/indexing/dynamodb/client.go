package dynamodb

import (
	"context"
	"errors"
	"iconrepo/internal/config"
	"iconrepo/internal/repositories/indexing"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	aws_dyndb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog"
)

func createConfig(conf *config.Options) (aws.Config, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           conf.DynamodbURL,
			SigningRegion: region,
		}, nil
	})

	return aws_config.LoadDefaultConfig(context.TODO(), aws_config.WithEndpointResolverWithOptions(customResolver))
}

func NewDynamodbClient(conf *config.Options) (*aws_dyndb.Client, error) {
	awsConf, err := createConfig(conf)
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

		tmpErr1 := &types.ConditionalCheckFailedException{}
		if errors.As(oe.Err, &tmpErr1) {
			return indexing.ErrConditionCheckFailed
		}
	}
	logger.Error().Err(err).Msg("some error occured")
	return err
}
