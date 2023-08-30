package indexing

import (
	"context"
	"errors"
	"fmt"
	"iconrepo/internal/repositories/indexing/dynamodb"

	aws_dyndb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
)

var errSideEffectTest = errors.New("some error occurred in side-effect")

type DynamodbTestRepository struct {
	*dynamodb.DynamodbRepository
}

func (testRepo *DynamodbTestRepository) Close() error {
	return testRepo.DynamodbRepository.Close()
}

func (testRepo *DynamodbTestRepository) GetIconCount(ctx context.Context) (int, error) {
	ionsTable := dynamodb.NewDyndbIconsTable(testRepo.GetAwsClient())
	icons, scanErr := ionsTable.GetItems(ctx)
	if scanErr != nil {
		return 0, scanErr
	}
	return len(icons), nil
}

func (testRepo *DynamodbTestRepository) GetIconFileCount(ctx context.Context) (int, error) {
	iconfileCount := 0
	iconsTable := dynamodb.NewDyndbIconsTable(testRepo.GetAwsClient())
	icons, scanErr := iconsTable.GetItems(ctx)
	if scanErr != nil {
		return 0, scanErr
	}

	for _, icon := range icons {
		iconfileCount = iconfileCount + len(icon.Iconfiles)
	}

	return iconfileCount, nil
}

func (testRepo *DynamodbTestRepository) GetTagRelationCount(ctx context.Context) (int, error) {
	tagRelationCount := 0
	iconTagsTable := dynamodb.NewDyndbIconTagsTable(testRepo.GetAwsClient())
	iconTagItems, scanErr := iconTagsTable.GetItems(ctx)
	if scanErr != nil {
		return 0, scanErr
	}
	for _, tagItem := range iconTagItems {
		tagRelationCount += int(tagItem.ReferenceCount)
	}

	return tagRelationCount, nil
}

func (testRepo *DynamodbTestRepository) ResetData(ctx context.Context) error {
	iconsTable := dynamodb.NewDyndbIconsTable(testRepo.GetAwsClient())

	icons, getIconsErr := iconsTable.GetItems(ctx)
	if getIconsErr != nil {
		return getIconsErr
	}

	for _, icon := range icons {
		deletErr := testRepo.DeleteAll(ctx, dynamodb.IconsTableName, icon)
		if deletErr != nil {
			return deletErr
		}
	}

	iconTagsTable := dynamodb.NewDyndbIconTagsTable(testRepo.GetAwsClient())
	iconTags, getIconTagsErr := iconTagsTable.GetItems(ctx)
	if getIconTagsErr != nil {
		return getIconTagsErr
	}

	for _, icon := range iconTags {
		deletErr := testRepo.DeleteAll(ctx, dynamodb.IconTagsTableName, icon)
		if deletErr != nil {
			return deletErr
		}
	}

	return dynamodb.DeleteLockItems(ctx, testRepo.GetAwsClient())
}

type keyGetter interface {
	GetKey(ctx context.Context) (map[string]types.AttributeValue, error)
}

func (testRepo *DynamodbTestRepository) DeleteAll(ctx context.Context, tableName string, item keyGetter) error {
	logger := zerolog.Ctx(ctx).With().Str("method", "DynamodbRepository.DeleteAll").Logger()

	logger.Debug().Msgf("deleting item %v...", item)
	key, getKeyErr := (item).GetKey(ctx)
	if getKeyErr != nil {
		return getKeyErr
	}
	input := &aws_dyndb.DeleteItemInput{
		TableName: &tableName,
		Key:       key,
	}
	_, deleteErr := testRepo.GetAwsClient().DeleteItem(ctx, input)
	if deleteErr != nil {
		return fmt.Errorf("failed to delete item %#v: %w", item, dynamodb.Unwrap(ctx, deleteErr))
	}
	return nil
}
