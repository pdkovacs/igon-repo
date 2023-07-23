package indexing

import (
	"context"
	"errors"
	"fmt"
	"iconrepo/internal/repositories/indexing/dynamodb"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

var errSideEffectTest = errors.New("some error occurred in side-effect")

type DynamodbTestRepository struct {
	*dynamodb.DynamodbRepository
}

func (testRepo *DynamodbTestRepository) Close() error {
	return errors.New("DynamodbTestRepository.Close not yet implemented")
}

func (testRepo *DynamodbTestRepository) GetIconCount(ctx context.Context) (int, error) {
	scanResult, scanErr := testRepo.ScanTable(ctx, dynamodb.IconsTableName)
	if scanErr != nil {
		return 0, scanErr
	}
	return len(scanResult), nil
}

func (testRepo *DynamodbTestRepository) GetIconFileCount(ctx context.Context) (int, error) {
	iconfileCount := 0
	scanResult, scanErr := testRepo.ScanTable(ctx, dynamodb.IconsTableName)
	if scanErr != nil {
		return 0, scanErr
	}
	for _, item := range scanResult {
		var iconItem dynamodb.DyndbIcon
		unmarshalErr := attributevalue.UnmarshalMap(item, &iconItem)
		if unmarshalErr != nil {
			return 0, fmt.Errorf("failed to unmarshal %T: %w", iconItem, dynamodb.Unwrap(ctx, unmarshalErr))
		}
		iconfileCount = iconfileCount + len(iconItem.Iconfiles)
	}

	return iconfileCount, nil
}

func (testRepo *DynamodbTestRepository) GetTagRelationCount(ctx context.Context) (int, error) {
	tagRelationCount := 0
	scanResult, scanErr := testRepo.ScanTable(ctx, dynamodb.IconTagsTableName)
	if scanErr != nil {
		return 0, scanErr
	}
	for _, item := range scanResult {
		var tagItem dynamodb.DyndbTag
		unmarshalErr := attributevalue.UnmarshalMap(item, &tagItem)
		if unmarshalErr != nil {
			return 0, fmt.Errorf("failed to unmarshal %T: %w", tagItem, dynamodb.Unwrap(ctx, unmarshalErr))
		}
		tagRelationCount += int(tagItem.ReferenceCount)
	}

	return tagRelationCount, nil
}

func (testRepo *DynamodbTestRepository) ResetData(ctx context.Context) error {
	for _, tableName := range dynamodb.GetAllTableNames() {
		scanResult, scanErr := testRepo.ScanTable(ctx, tableName)
		if scanErr != nil {
			return scanErr
		}
		deleteErr := testRepo.DeleteAll(ctx, tableName, scanResult)
		if deleteErr != nil {
			return deleteErr
		}
	}
	return nil
}
