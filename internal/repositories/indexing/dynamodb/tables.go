package dynamodb

import (
	"context"
	"fmt"

	aws_dyndb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
)

type DyndbIconsTable struct {
	awsClient *aws_dyndb.Client
}

func (iconsTable *DyndbIconsTable) GetItems(ctx context.Context) ([]*DyndbIcon, error) {
	logger := zerolog.Ctx(ctx).With().Str("method", "DyndbIconsTable.GetItems").Logger()

	items, err := GetItems(ctx, iconsTable.awsClient, IconsTableName, func() *DyndbIcon {
		return &DyndbIcon{}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get %T items: %w", DyndbIcon{}, err)
	}

	logger.Debug().Int("icon-tag-item-count", len(items)).Msg("returning")

	return items, nil
}

func NewDyndbIconsTable(awsClient *aws_dyndb.Client) *DyndbIconsTable {
	return &DyndbIconsTable{awsClient: awsClient}
}

type DyndbIconTagsTable struct {
	awsClient *aws_dyndb.Client
}

func (iconTagsTable *DyndbIconTagsTable) GetItems(ctx context.Context) ([]*DyndbTag, error) {
	logger := zerolog.Ctx(ctx).With().Str("method", "DyndbIconTagsTable.GetItems").Logger()

	items, err := GetItems(ctx, iconTagsTable.awsClient, IconTagsTableName, func() *DyndbTag {
		return &DyndbTag{}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get %T items: %w", DyndbIcon{}, err)
	}

	logger.Debug().Int("icon-item-count", len(items)).Msg("returning")

	return items, nil
}

func NewDyndbIconTagsTable(awsClient *aws_dyndb.Client) *DyndbIconTagsTable {
	return &DyndbIconTagsTable{awsClient: awsClient}
}

func DeleteLockItems(ctx context.Context, awsClient *aws_dyndb.Client) error {
	err := deleteLockItemsFromTable(ctx, awsClient, iconsLockTableName)
	if err != nil {
		return err
	}
	return deleteLockItemsFromTable(ctx, awsClient, iconTagsLockTableName)
}

func deleteLockItemsFromTable(ctx context.Context, awsClient *aws_dyndb.Client, tableName string) error {
	items, iconsScanErr := scanTable(ctx, awsClient, tableName)
	if iconsScanErr != nil {
		return fmt.Errorf("failed to scan %s: %w", tableName, iconsScanErr)
	}
	for _, item := range items {
		deleteItemInput := aws_dyndb.DeleteItemInput{
			TableName: &tableName,
			Key:       map[string]types.AttributeValue{"key": item["key"]},
		}
		_, deleteErr := awsClient.DeleteItem(ctx, &deleteItemInput)
		if deleteErr != nil {
			return fmt.Errorf("failed to delete item %v from %s: %w", item["key"], tableName, deleteErr)
		}
	}
	return nil
}

func scanTable(ctx context.Context, awsClient *aws_dyndb.Client, tableName string) ([]map[string]types.AttributeValue, error) {
	input := &aws_dyndb.ScanInput{
		TableName: &tableName,
	}
	result, scanErr := awsClient.Scan(ctx, input)
	if scanErr != nil {
		return nil, fmt.Errorf("failed to scan %s: %w", tableName, Unwrap(ctx, scanErr))
	}
	return result.Items, nil
}

// The need for the interface and the explicitly added `unmarshal` method is a work-around
// for this go issue: https://stackoverflow.com/a/71378366/1194266
func GetItems[T interface {
	*DyndbIcon | *DyndbTag
	unmarshal(attribs map[string]types.AttributeValue) error
}](
	ctx context.Context,
	awsClient *aws_dyndb.Client,
	tableName string,
	alloc func() T,
) ([]T, error) {
	logger := zerolog.Ctx(ctx).With().Str("method", "DyndbIconTagsTable.GetItems").Logger()

	scanResult, scanErr := scanTable(ctx, awsClient, tableName)
	if scanErr != nil {
		return nil, fmt.Errorf("failed to scan %s: %w", tableName, scanErr)
	}
	logger.Debug().Int("scan-result-count", len(scanResult)).Msg("found")

	items := []T{}

	for _, scanItem := range scanResult {
		item := alloc()
		unmarshalErr := item.unmarshal(scanItem)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}
		items = append(items, item)
	}

	logger.Debug().Int("icon-item-count", len(scanResult)).Msg("returning")

	return items, nil
}
