package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"iconrepo/internal/app/domain"
	"iconrepo/internal/config"
	"iconrepo/internal/repositories/indexing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/xid"
	"github.com/rs/zerolog"

	aws_dyndb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamodbRepository struct {
	awsClient *aws_dyndb.Client
}

func (repo *DynamodbRepository) GetAwsClient() *aws_dyndb.Client {
	return repo.awsClient
}

func (repo *DynamodbRepository) DescribeAllIcons(ctx context.Context) ([]domain.IconDescriptor, error) {
	iconTable := DyndbIconsTable{awsClient: repo.awsClient}
	items, scanErr := iconTable.GetItems(ctx)
	if scanErr != nil {
		return nil, fmt.Errorf("failed to fetch icon items: %w", scanErr)
	}

	iconDescriptors := []domain.IconDescriptor{}
	for _, item := range items {
		iconDescriptors = append(iconDescriptors, item.toIconDescriptor())
	}
	return iconDescriptors, nil
}

func (repo *DynamodbRepository) DescribeIcon(ctx context.Context, iconName string) (domain.IconDescriptor, error) {
	consistentRead := true
	iconItem, getIconItemErr := repo.getIconItem(ctx, iconName, consistentRead)
	if getIconItemErr != nil {
		return domain.IconDescriptor{}, fmt.Errorf("failed to describe icons table item %s: %w", iconName, getIconItemErr)
	}

	if iconItem.Tags == nil {
		iconItem.Tags = []string{}
	}

	return domain.IconDescriptor{
		IconAttributes: domain.IconAttributes{
			Name:       iconName,
			ModifiedBy: iconItem.ModifiedBy,
			Tags:       iconItem.Tags,
		},
		Iconfiles: toIconfileDescriptorList(iconItem.Iconfiles),
	}, nil
}

func (repo *DynamodbRepository) GetExistingTags(ctx context.Context) ([]string, error) {
	input := &aws_dyndb.ScanInput{
		TableName: &IconTagsTableName,
	}
	result, scanErr := repo.awsClient.Scan(ctx, input)
	if scanErr != nil {
		return nil, fmt.Errorf("failed to scan %s: %w", IconTagsTableName, Unwrap(ctx, scanErr))
	}
	tags := []string{}
	for _, tagItem := range result.Items {
		dynTag := &DyndbTag{}
		if unmarshalErr := dynTag.unmarshal(tagItem); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal tag: %w", unmarshalErr)
		}
		tags = append(tags, dynTag.Tag)
	}
	return tags, nil
}

func (repo *DynamodbRepository) CreateIcon(
	ctx context.Context,
	iconName string,
	iconfile domain.IconfileDescriptor,
	modifiedBy string,
	createSideEffect func() error,
) error {
	logger := zerolog.Ctx(ctx).With().Str("unit", "DynamodbRepository").Str("method", "CreateIcon").Logger()

	dyndbIconfile := DyndbIconfile{}
	dyndbIconfile.fromIconfileDescriptor(iconfile)

	changeId := createChangeId()

	icon := &DyndbIcon{
		IconName:   iconName,
		ModifiedBy: modifiedBy,
		Iconfiles:  []DyndbIconfile{dyndbIconfile},
		ChangeId:   changeId,
	}
	updateErr := repo.updateIcon(ctx, icon, "")
	if updateErr != nil {
		cause := updateErr
		if errors.Is(updateErr, indexing.ErrConditionCheckFailed) {
			cause = fmt.Errorf("change-id required: %s: %w", iconName, domain.ErrIconAlreadyExists)
		}
		return fmt.Errorf("failed to create icon %s: %w", iconName, cause)
	}

	if createSideEffect != nil {
		sideEffectErr := createSideEffect()
		if sideEffectErr != nil {
			rollbackErr := repo.deleteIcon(ctx, icon)
			if rollbackErr != nil {
				logger.Error().Err(rollbackErr).Str("IconName", iconName).Str("ChangeId", icon.ChangeId).Msg("failed to rollback on sideeffect error")
			}
			return sideEffectErr
		}
	}

	return nil
}

func (repo *DynamodbRepository) AddIconfileToIcon(
	ctx context.Context,
	iconName string,
	iconfile domain.IconfileDescriptor,
	modifiedBy string,
	createSideEffect func() error,
) error {
	logger := zerolog.Ctx(ctx).With().Str("unit", "DynamodbRepository").Str("method", "AddIconfileToIcon").Logger()

	original, getOriginalErr := repo.getIconItem(ctx, iconName, true)
	if getOriginalErr != nil {
		return fmt.Errorf("failed to get original of %s for adding iconfile to it: %w", iconName, getOriginalErr)
	}

	originalDescriptors := toIconfileDescriptorList(original.Iconfiles)
	for _, originalDescriptor := range originalDescriptors {
		if iconfile.Equals(originalDescriptor) {
			return domain.ErrIconfileAlreadyExists
		}
	}

	newChangeId := createChangeId()
	iconfileToAdd := DyndbIconfile{}
	iconfileToAdd.fromIconfileDescriptor(iconfile)

	updatedIcon := &DyndbIcon{IconName: iconName, ChangeId: newChangeId, ModifiedBy: modifiedBy}
	if original != nil {
		updatedIcon.Iconfiles = original.Iconfiles
	}
	updatedIcon.Iconfiles = append(updatedIcon.Iconfiles, iconfileToAdd)

	updateIconErr := repo.updateIcon(ctx, updatedIcon, original.ChangeId)
	if updateIconErr != nil {
		// TODO: back out and propagate meaningful error to client on failed condition validation
		return fmt.Errorf("failed to update icon %s: %w", iconName, updateIconErr)
	}

	if createSideEffect != nil {
		sideEffectErr := createSideEffect()
		if sideEffectErr != nil {
			rollbackErr := repo.updateIcon(ctx, original, newChangeId)
			if rollbackErr != nil {
				logger.Error().Err(rollbackErr).Str("IconName", iconName).Msg("failed to rollback on sideeffect error")
			}
			return sideEffectErr
		}
	}

	return nil
}

func (repo *DynamodbRepository) AddTag(ctx context.Context, iconName string, tag string, modifiedBy string) error {
	oldIconItem, getIconItemErr := repo.getIconItem(ctx, iconName, false)
	if getIconItemErr != nil {
		return fmt.Errorf("failed to get icon item %s to add tag %s to: %w", iconName, tag, getIconItemErr)
	}

	oldTags := oldIconItem.Tags
	for _, oldTag := range oldTags {
		if tag == oldTag {
			return nil
		}
	}

	newTags := append(oldTags, tag)
	changeId := createChangeId()

	newIconItem := &DyndbIcon{IconName: iconName, Iconfiles: oldIconItem.Iconfiles, ChangeId: changeId, ModifiedBy: modifiedBy, Tags: newTags}

	updateIconErr := repo.updateIcon(ctx, newIconItem, oldIconItem.ChangeId)
	if updateIconErr != nil {
		if errors.Is(updateIconErr, indexing.ErrConditionCheckFailed) {
			return indexing.ErrModifyingStaleItem
		} else {
			return fmt.Errorf("failed to update icon item %s to add tag %s to: %w", iconName, tag, updateIconErr)
		}
	}

	updateTagsErr := repo.updateTag(ctx, tag, true, changeId)
	if updateTagsErr != nil {
		// Roll back
		deleteErr := repo.deleteIcon(ctx, newIconItem)
		if deleteErr != nil {
			if errors.Is(deleteErr, indexing.ErrConditionCheckFailed) {
				// the IconItem has been concurrently overwritten, not much to do
				return nil
			}
			return fmt.Errorf("failed to rollback updating icon %s with tag %s: %w", iconName, tag, deleteErr)
		}

		if errors.Is(updateTagsErr, indexing.ErrConditionCheckFailed) {
			// TODO: expect condition validation failure reported; perhaps the same tag was added to, or removed from, another icon?
			fmt.Printf(">>>>>>>>>>>>>> repeat all updates before giving up and rolling back\n")
		}
		return fmt.Errorf("failed to update tags table with tag %s to be added to icon item %s: %w", tag, iconName, updateTagsErr)
	}

	return nil
}

func (repo *DynamodbRepository) RemoveTag(ctx context.Context, iconName string, tag string, modifiedBy string) error {
	oldIconItem, getIconItemErr := repo.getIconItem(ctx, iconName, false)
	if getIconItemErr != nil {
		return fmt.Errorf("failed to get icon item %s to add tag %s to: %w", iconName, tag, getIconItemErr)
	}

	newTags := []string{}
	oldTags := oldIconItem.Tags
	foundTag := false
	for _, oldTag := range oldTags {
		if tag == oldTag {
			foundTag = true
			continue
		}
		newTags = append(newTags, oldTag)
	}
	if !foundTag {
		return nil
	}

	changeId := createChangeId()

	newIconItem := &DyndbIcon{IconName: iconName, Iconfiles: oldIconItem.Iconfiles, ChangeId: changeId, ModifiedBy: modifiedBy, Tags: newTags}

	updateIconErr := repo.updateIcon(ctx, newIconItem, oldIconItem.ChangeId)
	if updateIconErr != nil {
		if errors.Is(updateIconErr, indexing.ErrConditionCheckFailed) {
			return indexing.ErrModifyingStaleItem
		} else {
			return fmt.Errorf("failed to update icon item %s to add tag %s to: %w", iconName, tag, updateIconErr)
		}
	}

	updateTagsErr := repo.updateTag(ctx, tag, true, changeId)
	if updateTagsErr != nil {
		// Roll back
		deleteErr := repo.deleteIcon(ctx, newIconItem)
		if deleteErr != nil {
			if errors.Is(deleteErr, indexing.ErrConditionCheckFailed) {
				// the IconItem has been concurrently overwritten, not much to do
				return nil
			}
			return fmt.Errorf("failed to rollback updating icon %s with tag %s: %w", iconName, tag, deleteErr)
		}

		if errors.Is(updateTagsErr, indexing.ErrConditionCheckFailed) {
			// TODO: expect condition validation failure reported; perhaps the same tag was added to, or removed from, another icon?
			fmt.Printf(">>>>>>>>>>>>>> repeat all updates before giving up and rolling back\n")
		}
		return fmt.Errorf("failed to update tags table with tag %s to be added to icon item %s: %w", tag, iconName, updateTagsErr)
	}

	return nil
}

func (repo *DynamodbRepository) DeleteIcon(ctx context.Context, iconName string, modifiedBy string, createSideEffect func() error) error {
	logger := zerolog.Ctx(ctx).With().Str("method", "DynamodbRepository.DeleteIcon").Logger()

	iconItem, getIconItemErr := repo.getIconItem(ctx, iconName, false)
	if getIconItemErr != nil {
		return fmt.Errorf("failed to fetch %s for deletion (to delete associated tags): %w", iconName, getIconItemErr)
	}

	changeId := createChangeId()

	for _, tag := range iconItem.Tags {
		updateTagErr := repo.updateTag(ctx, tag, false, changeId)
		if updateTagErr != nil {
			// TODO: restore old state of the tags updated so far
			return fmt.Errorf("failed to update tag %s: %w", tag, updateTagErr)
		}
	}

	deleteErr := repo.deleteIcon(ctx, iconItem)
	if deleteErr != nil {
		return fmt.Errorf("failed to delete icon %s: %w", iconName, deleteErr)
	}

	if createSideEffect != nil {
		sideEffectErr := createSideEffect()
		if sideEffectErr != nil {
			// TODO: rollback tags updates
			rollbackErr := repo.updateIcon(ctx, iconItem, "")
			if rollbackErr != nil {
				logger.Error().Err(rollbackErr).Str("IconName", iconName).Msg("failed to rollback on side-effect error")
			}
			return fmt.Errorf("failed to delete icon %s due to side-effect failure: %w", iconName, sideEffectErr)
		}
	}

	return nil
}

func (repo *DynamodbRepository) DeleteIconfile(
	ctx context.Context,
	iconName string,
	iconfile domain.IconfileDescriptor,
	modifiedBy string,
	createSideEffect func() error,
) error {
	logger := zerolog.Ctx(ctx).With().Str("method", "DynamodbRepository.DeleteIconfile").Logger()

	oldIconItem, getIconItemErr := repo.getIconItem(ctx, iconName, false)
	if getIconItemErr != nil {
		return fmt.Errorf("failed to get DyndbIcon to remove %s: %w", iconName, getIconItemErr)
	}

	newChangeId := createChangeId()
	newIconItem := DyndbIcon{
		IconName:   iconName,
		ChangeId:   newChangeId,
		ModifiedBy: modifiedBy,
		Tags:       oldIconItem.Tags,
	}

	newIconfiles := []DyndbIconfile{}
	found := false
	for _, oldIconfile := range oldIconItem.Iconfiles {
		if oldIconfile.toIconfileDescriptor().Equals(iconfile) {
			found = true
			continue
		}
		newIconfiles = append(newIconfiles, oldIconfile)
	}
	if !found {
		return domain.ErrIconfileNotFound
	}
	if len(newIconfiles) == 0 {
		deleteIconErr := repo.DeleteIcon(ctx, iconName, modifiedBy, createSideEffect)
		if deleteIconErr != nil {
			return fmt.Errorf("failed to delete icon (with no more iconfiles left) %s: %w", iconName, deleteIconErr)
		}
		return nil
	}
	newIconItem.Iconfiles = newIconfiles

	updateErr := repo.updateIcon(ctx, &newIconItem, oldIconItem.ChangeId)
	if updateErr != nil {
		cause := updateErr
		if errors.Is(updateErr, indexing.ErrConditionCheckFailed) {
			cause = fmt.Errorf("change-id required: %s: %w", oldIconItem.ChangeId, indexing.ErrModifyingStaleItem)
		}
		return fmt.Errorf("failed to update icon %s with %v removed: %w", iconName, iconfile, cause)
	}

	if createSideEffect != nil {
		sideEffectErr := createSideEffect()
		if sideEffectErr != nil {
			rollbackErr := repo.updateIcon(ctx, oldIconItem, newIconItem.ChangeId)
			if rollbackErr != nil {
				logger.Error().Err(rollbackErr).Str("IconName", iconName).Msg("failed to rollback on side-effect error")
			}
			return fmt.Errorf("failed to delete icon file %v from %s due to side-effect failure: %w", iconfile, iconName, sideEffectErr)
		}
	}

	return nil
}

func (repo *DynamodbRepository) getIconItem(ctx context.Context, iconName string, consistentRead bool) (*DyndbIcon, error) {
	logger := zerolog.Ctx(ctx).With().Str("unit", "DynamodbRepository").Str("method", "getIconItem").Logger()

	icon := &DyndbIcon{IconName: iconName}
	key, keyErr := icon.GetKey(ctx)
	if keyErr != nil {
		return icon, keyErr
	}

	input := &aws_dyndb.GetItemInput{
		TableName:      &IconsTableName,
		Key:            key,
		ConsistentRead: &consistentRead,
	}
	output, getItemErr := repo.awsClient.GetItem(ctx, input)
	if getItemErr != nil {
		return icon, fmt.Errorf("failed to get icon item for %s: %w", iconName, Unwrap(ctx, getItemErr))
	}

	if logger.GetLevel() == zerolog.DebugLevel {
		logger.Debug().Interface("output", output).Str("api", "dynamodb.GetItem").Send()
	}
	if output.Item == nil {
		return icon, domain.ErrIconNotFound
	}

	if unmarshalErr := icon.unmarshal(output.Item); unmarshalErr != nil {
		return icon, fmt.Errorf("failed to unmarshal icon item for %s: %w", iconName, Unwrap(ctx, unmarshalErr))
	}

	return icon, nil
}

func (repo *DynamodbRepository) updateIcon(ctx context.Context, icon *DyndbIcon, requiredChangeId string) error {
	logger := zerolog.Ctx(ctx).With().Str("method", "DynamodbRepository.updateIcon").Logger()

	newItem, marshalErr := attributevalue.MarshalMap(icon)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal new icon item for updating %s: %w", icon.IconName, marshalErr)
	}

	conditionExpression, conditionNames, conditionValues, condBuildErr := buildTimestampCondition(ctx, requiredChangeId)
	if condBuildErr != nil {
		return fmt.Errorf("failed to build 'must be equal' condition expression for %s: %w", requiredChangeId, condBuildErr)
	}

	if len(requiredChangeId) == 0 {
		conditionExpression, conditionNames, conditionValues, condBuildErr = buildPkNotExistsCondition(ctx, iconNameAttribute)
		if condBuildErr != nil {
			return fmt.Errorf("failed to build 'attribute not exists' condition expression for %s: %w", iconNameAttribute, condBuildErr)
		}
	}

	input := &aws_dyndb.PutItemInput{
		TableName:                 aws.String(IconsTableName),
		Item:                      newItem,
		ConditionExpression:       conditionExpression,
		ExpressionAttributeNames:  conditionNames,
		ExpressionAttributeValues: conditionValues,
		ReturnConsumedCapacity:    "TOTAL",
	}
	output, err := repo.awsClient.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update icon %s: %w", icon.IconName, Unwrap(ctx, err))
	}
	if logger.GetLevel() == zerolog.DebugLevel {
		logger.Debug().Str("IconName", icon.IconName).Interface("output", output).Send()
	}

	return nil
}

func (repo *DynamodbRepository) deleteIcon(ctx context.Context, iconItem *DyndbIcon) error {
	logger := zerolog.Ctx(ctx).With().Str("unit", "DynamodbRepository").Str("method", "deleteIcon").Logger()

	pk, getKeyErr := iconItem.GetKey(ctx)
	if getKeyErr != nil {
		return fmt.Errorf("failed to get key for deleting %s: %w", iconItem.IconName, getKeyErr)
	}

	conditionExpression, conditionNames, conditionValues, condBuildErr := buildTimestampCondition(ctx, iconItem.ChangeId)
	if condBuildErr != nil {
		return condBuildErr
	}

	input := &aws_dyndb.DeleteItemInput{
		TableName:                 aws.String(IconsTableName),
		Key:                       pk,
		ConditionExpression:       conditionExpression,
		ExpressionAttributeNames:  conditionNames,
		ExpressionAttributeValues: conditionValues,
		ReturnConsumedCapacity:    "TOTAL",
	}
	output, err := repo.awsClient.DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete icon %s: %w", iconItem.IconName, Unwrap(ctx, err))
	}

	if logger.GetLevel() == zerolog.DebugLevel {
		logger.Debug().Str("IconName", iconItem.IconName).Interface("delete-result", output).Send()
	}

	return nil
}

func (repo *DynamodbRepository) getTagItem(ctx context.Context, tag string, consistentRead bool) (*DyndbTag, error) {
	logger := zerolog.Ctx(ctx).With().Str("unit", "DynamodbRepository").Str("method", "getTagItem").Logger()

	tagItem := &DyndbTag{Tag: tag}
	key, keyErr := tagItem.GetKey(ctx)
	if keyErr != nil {
		return tagItem, keyErr
	}

	input := &aws_dyndb.GetItemInput{
		TableName:      &IconTagsTableName,
		Key:            key,
		ConsistentRead: &consistentRead,
	}
	output, getItemErr := repo.awsClient.GetItem(ctx, input)
	if getItemErr != nil {
		return tagItem, fmt.Errorf("failed to get tag item for %s: %w", tag, Unwrap(ctx, getItemErr))
	}

	if logger.GetLevel() == zerolog.DebugLevel {
		logger.Debug().Interface("output", output).Str("api", "dynamodb.GetItem").Send()
	}
	if output.Item == nil {
		return nil, nil
	}

	unmarshalErr := tagItem.unmarshal(output.Item)
	if unmarshalErr != nil {
		return tagItem, fmt.Errorf("failed to unmarshal tag item for %s: %w", tag, Unwrap(ctx, unmarshalErr))
	}

	return tagItem, nil
}

func (repo *DynamodbRepository) updateTag(ctx context.Context, tag string, add bool, parentChangeId string) error {
	oldTagItem, getTagItemErr := repo.getTagItem(ctx, tag, true)
	if getTagItemErr != nil {
		return fmt.Errorf("failed to get old tag item %s: %w", tag, getTagItemErr)
	}

	oldChangeId := ""
	if oldTagItem != nil {
		oldChangeId = oldTagItem.ChangeId
	}
	conditionExpression, conditionNames, conditionValues, condBuildErr := buildTimestampCondition(ctx, oldChangeId)
	if condBuildErr != nil {
		return condBuildErr
	}

	var newTagItem = &DyndbTag{Tag: tag, ChangeId: parentChangeId}
	if oldTagItem == nil {
		oldTagItem = &DyndbTag{ChangeId: "", ReferenceCount: 0} // Don't require time stamp during update
	}
	if add {
		newTagItem.ReferenceCount = oldTagItem.ReferenceCount + 1
	} else {
		newTagItem.ReferenceCount = oldTagItem.ReferenceCount - 1
	}

	if newTagItem.ReferenceCount > 0 {
		newItem, marshalErr := attributevalue.MarshalMap(newTagItem)
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal new icon item for updating %s: %w", tag, marshalErr)
		}

		input := &aws_dyndb.PutItemInput{
			TableName:                 aws.String(IconTagsTableName),
			Item:                      newItem,
			ConditionExpression:       conditionExpression,
			ExpressionAttributeNames:  conditionNames,
			ExpressionAttributeValues: conditionValues,
			ReturnConsumedCapacity:    "TOTAL",
		}
		_, err := repo.awsClient.PutItem(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to update icon %s: %w", tag, Unwrap(ctx, err))
		}

		return nil
	}

	pk, getKeyErr := oldTagItem.GetKey(ctx)
	if getKeyErr != nil {
		return fmt.Errorf("failed to generate PK of %v for removal: %w", oldTagItem, getKeyErr)
	}

	input := &aws_dyndb.DeleteItemInput{
		TableName:                 &IconTagsTableName,
		Key:                       pk,
		ConditionExpression:       conditionExpression,
		ExpressionAttributeNames:  conditionNames,
		ExpressionAttributeValues: conditionValues,
		ReturnConsumedCapacity:    "TOTAL",
	}
	_, deleteErr := repo.awsClient.DeleteItem(ctx, input)
	if deleteErr != nil {
		return fmt.Errorf("failed to delete tag %v: %w", oldTagItem, deleteErr)
	}

	return nil
}

func NewDynamodbRepository(conf *config.Options) (*DynamodbRepository, error) {
	awsConf, err := createConfig(conf)
	if err != nil {
		return nil, err
	}
	svc := aws_dyndb.NewFromConfig(awsConf)
	return &DynamodbRepository{svc}, nil
}

func createChangeId() string {
	return xid.New().String()
}

func buildTimestampCondition(ctx context.Context, requiredChangeId string) (*string, map[string]string, map[string]types.AttributeValue, error) {
	var conditionExpression *string
	var conditionNames map[string]string
	var conditionValues map[string]types.AttributeValue
	if len(requiredChangeId) > 0 {
		condBuilder := expression.Name(changeIdAttribute).Equal(expression.Value(requiredChangeId))
		condEx, buildErr := expression.NewBuilder().WithCondition(condBuilder).Build()
		if buildErr != nil {
			return conditionExpression, conditionNames, conditionValues, fmt.Errorf("failed to build conditional expression: %w", Unwrap(ctx, buildErr))
		}
		conditionExpression = condEx.Condition()
		conditionNames = condEx.Names()
		conditionValues = condEx.Values()
	}
	return conditionExpression, conditionNames, conditionValues, nil
}

func buildPkNotExistsCondition(ctx context.Context, pk string) (*string, map[string]string, map[string]types.AttributeValue, error) {
	var conditionExpression *string
	var conditionNames map[string]string
	var conditionValues map[string]types.AttributeValue
	condBuilder := expression.AttributeNotExists(expression.Name(pk))
	condEx, buildErr := expression.NewBuilder().WithCondition(condBuilder).Build()
	if buildErr != nil {
		return conditionExpression, conditionNames, conditionValues, fmt.Errorf("failed to build conditional expression: %w", Unwrap(ctx, buildErr))
	}
	conditionExpression = condEx.Condition()
	conditionNames = condEx.Names()
	conditionValues = condEx.Values()
	return conditionExpression, conditionNames, conditionValues, nil
}
