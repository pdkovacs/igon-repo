package dynamodb

import (
	"errors"
	"fmt"
	"iconrepo/internal/app/domain"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var emptyIconDescriptor = domain.IconDescriptor{}

func NewDynDBRepository() *DynDBRepository {
	return &DynDBRepository{}
}

type DynDBRepository struct{}

func (dynDb *DynDBRepository) DescribeAllIcons() ([]domain.IconDescriptor, error) {
	return nil, errors.New("not yet implemented")
}

func (dynDb *DynDBRepository) DescribeIcon(iconName string) (domain.IconDescriptor, error) {
	return emptyIconDescriptor, errors.New("not yet implemented")
}

func (dynDb *DynDBRepository) GetExistingTags() ([]string, error) {
	return nil, errors.New("not yet implemented")
}

func (dynDb *DynDBRepository) CreateIcon(iconName string, iconfile domain.IconfileDescriptor, modifiedBy string, createSideEffect func() error) error {
	dynDbSession, sessionCreationErr := session.NewSession()
	if sessionCreationErr != nil {
		return sessionCreationErr
	}
	svc := dynamodb.New(dynDbSession)
	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"IconName": {
				S: aws.String(fmt.Sprintf("%d#%s", 0, iconName)), // TODO: fix id
			},
			"Tag": {
				S: aws.String(fmt.Sprintf("%s#%s", iconfile.Format, iconfile.Size)),
			},
			"ModifiedBy": {
				S: aws.String(modifiedBy),
			},
			"ModifiedAt": {
				N: aws.String(fmt.Sprintf("%d", time.Now().UnixMilli())),
			},
		},
		ReturnConsumedCapacity: aws.String("TOTAL"),
		TableName:              aws.String("icons"),
	}

	_, err := svc.PutItem(input) // TODO: check result?
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				fmt.Println(dynamodb.ErrCodeConditionalCheckFailedException, aerr.Error())
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
			case dynamodb.ErrCodeItemCollectionSizeLimitExceededException:
				fmt.Println(dynamodb.ErrCodeItemCollectionSizeLimitExceededException, aerr.Error())
			case dynamodb.ErrCodeTransactionConflictException:
				fmt.Println(dynamodb.ErrCodeTransactionConflictException, aerr.Error())
			case dynamodb.ErrCodeRequestLimitExceeded:
				fmt.Println(dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
			case dynamodb.ErrCodeInternalServerError:
				fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return err
	}
	return nil
}

func (dynDb *DynDBRepository) AddIconfileToIcon(iconName string, iconfile domain.IconfileDescriptor, modifiedBy string, createSideEffect func() error) error {
	return errors.New("not yet implemented")
}

func (dynDb *DynDBRepository) AddTag(iconName string, tag string, modifiedBy string) error {
	return errors.New("not yet implemented")
}

func (dynDb *DynDBRepository) RemoveTag(iconName string, tag string, modifiedBy string) error {
	return errors.New("not yet implemented")
}

func (dynDb *DynDBRepository) DeleteIcon(iconName string, modifiedBy string, createSideEffect func() error) error {
	return errors.New("not yet implemented")
}

func (dynDb *DynDBRepository) DeleteIconfile(iconName string, iconfile domain.IconfileDescriptor, modifiedBy string, createSideEffect func() error) error {
	return errors.New("not yet implemented")
}
