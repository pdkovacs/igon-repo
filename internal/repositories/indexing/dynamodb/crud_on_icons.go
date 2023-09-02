package dynamodb

import (
	"context"
	"fmt"
	"iconrepo/internal/app/domain"
	"iconrepo/internal/config"
	"iconrepo/internal/logging"
	"time"

	"cirello.io/dynamolock/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/rs/zerolog"

	aws_dyndb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type dlockLogger struct {
	logger zerolog.Logger
}

func (l *dlockLogger) Println(arg ...any) {
	if l.logger.GetLevel() == zerolog.DebugLevel {
		l.logger.Print(arg...)
	}
}

type DynamodbRepository struct {
	awsClient          *aws_dyndb.Client
	iconsLockClient    *dynamolock.Client
	iconTagsLockClient *dynamolock.Client
}

func NewDynamodbRepository(conf *config.Options) (*DynamodbRepository, error) {
	svc, clientErr := NewDynamodbClient(conf)
	if clientErr != nil {
		return nil, fmt.Errorf("failed to create dynamodb client: %w", clientErr)
	}

	iconsLockClient, iconsLockClientErr := dynamolock.New(
		svc,
		iconsLockTableName,
		dynamolock.WithLeaseDuration(3*time.Second),
		dynamolock.WithHeartbeatPeriod(1*time.Second),
		dynamolock.WithLogger(&dlockLogger{logging.Get().With().Str("unit", "cirello.io/dynamolock/v2").Logger()}),
	)
	if iconsLockClientErr != nil {
		return nil, fmt.Errorf("failed to create iconsLockClient: %w", iconsLockClientErr)
	}

	iconTagsLockClient, iconTagsLockClientErr := dynamolock.New(
		svc,
		iconTagsLockTableName,
		dynamolock.WithLeaseDuration(3*time.Second),
		dynamolock.WithHeartbeatPeriod(1*time.Second),
		dynamolock.WithLogger(&dlockLogger{logging.Get().With().Str("unit", "cirello.io/dynamolock/v2").Logger()}),
	)
	if iconsLockClientErr != nil {
		return nil, fmt.Errorf("failed to create iconTagsLockClient: %w", iconTagsLockClientErr)
	}

	return &DynamodbRepository{svc, iconsLockClient, iconTagsLockClient}, nil
}

func (repo *DynamodbRepository) Close() error {
	logger := logging.Get().With().Str("unit", "DynamodbRepository").Str("method", "Close").Logger()

	var myErr error

	iconsLockCloseErr := repo.iconsLockClient.Close()
	if iconsLockCloseErr != nil {
		logger.Err(iconsLockCloseErr).Msg("error while closing iconsLockClose")
		myErr = iconsLockCloseErr
	}

	iconTagsLockCloseErr := repo.iconTagsLockClient.Close()
	if iconTagsLockCloseErr != nil {
		logger.Err(iconTagsLockCloseErr).Msg("error while closing iconTagsLockClose")
		if myErr == nil { // we only return the first one
			myErr = iconTagsLockCloseErr
		}
	}

	if myErr != nil {
		return fmt.Errorf("errors while closing DynamodbRepository: %w", myErr)
	}

	return nil
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
		TableName: aws.String(IconTagsTableName),
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

	icon := &DyndbIcon{
		IconName:   iconName,
		ModifiedBy: modifiedBy,
		Iconfiles:  []DyndbIconfile{dyndbIconfile},
	}

	lock, lockErr := repo.iconsLockClient.AcquireLockWithContext(ctx, iconName, dynamolock.WithData([]byte("CreateIcon")), dynamolock.ReplaceData())
	if lockErr != nil {
		return fmt.Errorf("failed to acquire lock on icons_table#%s: %w", iconName, lockErr)
	}
	defer func() {
		success, releaseLockErr := repo.iconsLockClient.ReleaseLock(lock)
		if releaseLockErr != nil {
			logger.Error().Err(releaseLockErr).Msg("error while releasing lock on icons table")
		}
		if success {
			logger.Debug().Str("lockTable", "icons").Str("lockKey", iconName).Bool("lockReleaseResult", success).Send()
		} else {
			logger.Info().Str("lockTable", "icons").Str("lockKey", iconName).Bool("lockReleaseResult", success).Send()
		}
	}()

	updateErr := repo.updateIcon(ctx, icon)
	if updateErr != nil {
		return fmt.Errorf("failed to create icon %s: %w", iconName, updateErr)
	}

	if createSideEffect != nil {
		sideEffectErr := createSideEffect()
		if sideEffectErr != nil {
			rollbackErr := repo.deleteIcon(ctx, icon)
			if rollbackErr != nil {
				logger.Error().Err(rollbackErr).Str("IconName", iconName).Msg("failed to rollback on sideeffect error")
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

	lock, lockErr := repo.iconsLockClient.AcquireLockWithContext(ctx, iconName, dynamolock.WithData([]byte("AddIconfileToIcon")), dynamolock.ReplaceData())
	if lockErr != nil {
		return fmt.Errorf("failed to acquire lock on icons_table#%s: %w", iconName, lockErr)
	}
	defer func() {
		success, releaseLockErr := repo.iconsLockClient.ReleaseLock(lock)
		if releaseLockErr != nil {
			logger.Error().Err(releaseLockErr).Msg("error while releasing lock on icons table")
		}
		if success {
			logger.Debug().Str("lockTable", "icons").Str("lockKey", iconName).Bool("lockReleaseResult", success).Send()
		} else {
			logger.Info().Str("lockTable", "icons").Str("lockKey", iconName).Bool("lockReleaseResult", success).Send()
		}
	}()

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

	iconfileToAdd := DyndbIconfile{}
	iconfileToAdd.fromIconfileDescriptor(iconfile)

	updatedIcon := &DyndbIcon{IconName: iconName, ModifiedBy: modifiedBy}
	if original != nil {
		updatedIcon.Iconfiles = original.Iconfiles
	}
	updatedIcon.Iconfiles = append(updatedIcon.Iconfiles, iconfileToAdd)

	updateIconErr := repo.updateIcon(ctx, updatedIcon)
	if updateIconErr != nil {
		// TODO: back out and propagate meaningful error to client on failed condition validation
		return fmt.Errorf("failed to update icon %s: %w", iconName, updateIconErr)
	}

	if createSideEffect != nil {
		sideEffectErr := createSideEffect()
		if sideEffectErr != nil {
			rollbackErr := repo.updateIcon(ctx, original)
			if rollbackErr != nil {
				logger.Error().Err(rollbackErr).Str("IconName", iconName).Msg("failed to rollback on sideeffect error")
			}
			return sideEffectErr
		}
	}

	return nil
}

func (repo *DynamodbRepository) AddTag(ctx context.Context, iconName string, tag string, modifiedBy string) error {
	logger := zerolog.Ctx(ctx).With().Str("unit", "DynamodbRepository").Str("method", "AddTag").Logger()

	lock, lockErr := repo.iconsLockClient.AcquireLockWithContext(ctx, iconName, dynamolock.WithData([]byte("AddTag")), dynamolock.ReplaceData())
	if lockErr != nil {
		return fmt.Errorf("failed to acquire lock on icons_table#%s: %w", iconName, lockErr)
	}
	defer func() {
		success, releaseLockErr := repo.iconsLockClient.ReleaseLock(lock)
		if releaseLockErr != nil {
			logger.Error().Err(releaseLockErr).Msg("error while releasing lock on icons table")
		}
		if success {
			logger.Debug().Str("lockTable", "icons").Str("lockKey", iconName).Bool("lockReleaseResult", success).Send()
		} else {
			logger.Info().Str("lockTable", "icons").Str("lockKey", iconName).Bool("lockReleaseResult", success).Send()
		}
	}()

	tagsLock, lockErr := repo.iconTagsLockClient.AcquireLockWithContext(ctx, tag, dynamolock.WithData([]byte("AddTag")), dynamolock.ReplaceData())
	if lockErr != nil {
		return fmt.Errorf("failed to acquire lock on icon_tags_table#%s: %w", iconName, lockErr)
	}
	defer repo.iconTagsLockClient.ReleaseLock(tagsLock)

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

	newIconItem := &DyndbIcon{IconName: iconName, Iconfiles: oldIconItem.Iconfiles, ModifiedBy: modifiedBy, Tags: newTags}

	updateIconErr := repo.updateIcon(ctx, newIconItem)
	if updateIconErr != nil {
		return fmt.Errorf("failed to update icon item %s to add tag %s to: %w", iconName, tag, updateIconErr)
	}

	oldTagItem, err := repo.getTagItem(ctx, tag, true)
	if err != nil {
		return fmt.Errorf("failed to get old tag-items for icon %s about to be deleted: %w", iconName, err)
	}
	if oldTagItem == nil {
		oldTagItem = &DyndbTag{tag, 0}
	}

	updateTagsErr := repo.updateTag(ctx, *oldTagItem, true)
	if updateTagsErr != nil {
		// Roll back
		updateErr := repo.updateIcon(ctx, oldIconItem)
		if updateErr != nil {
			return fmt.Errorf("failed to rollback updating icon %s with tag %s: %w", iconName, tag, updateErr)
		}
		return fmt.Errorf("failed to update tags table with tag %s to be added to icon item %s: %w", tag, iconName, updateTagsErr)
	}

	return nil
}

func (repo *DynamodbRepository) RemoveTag(ctx context.Context, iconName string, tag string, modifiedBy string) error {
	logger := zerolog.Ctx(ctx).With().Str("method", "DynamodbRepository.RemoveTag").Logger()

	lock, lockErr := repo.iconsLockClient.AcquireLockWithContext(ctx, iconName, dynamolock.WithData([]byte("RemoveTag")), dynamolock.ReplaceData())
	if lockErr != nil {
		lock, readDataErr := repo.iconsLockClient.Get(iconName)
		if readDataErr != nil {
			logger.Error().Err(readDataErr).Str("lockTable", "icons").Str("lockKey", iconName).Msg("failed to read lock data after failing to acquire lock")
		}
		return fmt.Errorf("failed to acquire lock on icons_table#%s (<%s): %w", iconName, string(lock.Data()), lockErr)
	}
	defer func() {
		success, releaseLockErr := repo.iconsLockClient.ReleaseLock(lock)
		if releaseLockErr != nil {
			logger.Error().Err(releaseLockErr).Msg("error while releasing lock on icons table")
		}
		if success {
			logger.Debug().Str("lockTable", "icons").Str("lockKey", iconName).Bool("lockReleaseResult", success).Send()
		} else {
			logger.Info().Str("lockTable", "icons").Str("lockKey", iconName).Bool("lockReleaseResult", success).Send()
		}
	}()

	tagsLock, lockErr := repo.iconTagsLockClient.AcquireLockWithContext(ctx, tag, dynamolock.WithData([]byte("RemoveTag")), dynamolock.ReplaceData())
	if lockErr != nil {
		return fmt.Errorf("failed to acquire lock on icon_tags_table#%s: %w", tag, lockErr)
	}
	defer repo.iconTagsLockClient.ReleaseLock(tagsLock)

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

	newIconItem := &DyndbIcon{IconName: iconName, Iconfiles: oldIconItem.Iconfiles, ModifiedBy: modifiedBy, Tags: newTags}

	updateIconErr := repo.updateIcon(ctx, newIconItem)
	if updateIconErr != nil {
		return fmt.Errorf("failed to update icon item %s to add tag %s to: %w", iconName, tag, updateIconErr)
	}

	oldTagItem, err := repo.getTagItem(ctx, tag, true)
	if err != nil {
		return fmt.Errorf("failed to get old tag-items for icon %s about to be deleted: %w", iconName, err)
	}

	updateTagsErr := repo.updateTag(ctx, *oldTagItem, false)
	if updateTagsErr != nil {
		// Roll back
		updateErr := repo.updateIcon(ctx, oldIconItem)
		if updateErr != nil {
			return fmt.Errorf("failed to rollback updating icon %s with tag %s: %w", iconName, tag, updateErr)
		}
		return fmt.Errorf("failed to update tags table with tag %s to be added to icon item %s: %w", tag, iconName, updateTagsErr)
	}

	return nil
}

func (repo *DynamodbRepository) DeleteIcon(ctx context.Context, iconName string, modifiedBy string, createSideEffect func() error) error {
	logger := zerolog.Ctx(ctx).With().Str("method", "DynamodbRepository.DeleteIcon").Str("iconName", iconName).Logger()

	lock, lockErr := repo.iconsLockClient.AcquireLockWithContext(ctx, iconName, dynamolock.WithData([]byte("DeleteIcon")), dynamolock.ReplaceData())
	if lockErr != nil {
		lock, readDataErr := repo.iconsLockClient.Get(iconName)
		if readDataErr != nil {
			logger.Error().Err(readDataErr).Str("lockTable", "icons").Str("lockKey", iconName).Msg("failed to read lock data after failing to acquire lock")
		}
		return fmt.Errorf("failed to acquire lock on icons_table#%s (<%s): %w", iconName, string(lock.Data()), lockErr)
	}
	defer func() {
		success, releaseLockErr := repo.iconsLockClient.ReleaseLock(lock)
		if releaseLockErr != nil {
			logger.Error().Err(releaseLockErr).Msg("error while releasing lock on icons table")
		}
		if success {
			logger.Debug().Str("lockTable", "icons").Str("lockKey", iconName).Bool("lockReleaseResult", success).Send()
		} else {
			logger.Info().Str("lockTable", "icons").Str("lockKey", iconName).Bool("lockReleaseResult", success).Send()
		}
	}()

	return repo.deleteIconNoLock(ctx, iconName, modifiedBy, createSideEffect)
}

func (repo *DynamodbRepository) deleteIconNoLock(ctx context.Context, iconName string, modifiedBy string, createSideEffect func() error) error {
	logger := zerolog.Ctx(ctx).With().Str("method", "DynamodbRepository.deleteIcon0").Str("iconName", iconName).Logger()

	iconItem, getIconItemErr := repo.getIconItem(ctx, iconName, false)
	if getIconItemErr != nil {
		return fmt.Errorf("failed to fetch %s for deletion (to delete associated tags): %w", iconName, getIconItemErr)
	}

	tagsUpdatedWithDecrRefCount := []*DyndbTag{}

	rollbackTagsUpdatedSoFar := func() {
		for _, item := range tagsUpdatedWithDecrRefCount {
			rollbackErr := repo.updateTag(ctx, *item, true)
			if rollbackErr != nil {
				logger.Error().Err(rollbackErr).Str("icon", iconName).Str("tag", item.Tag).Msg("failed to rollback tag ref-count decrement")
			}
		}
	}

	for _, oldTag := range iconItem.Tags {
		oldItem, err := repo.getTagItem(ctx, oldTag, true)
		if err != nil {
			return fmt.Errorf("failed to get old tag-items for icon %s about to be deleted: %w", iconName, err)
		}

		tagsLock, lockErr := repo.iconTagsLockClient.AcquireLockWithContext(ctx, oldTag, dynamolock.WithData([]byte("DeleteIcon")), dynamolock.ReplaceData())
		if lockErr != nil {
			rollbackTagsUpdatedSoFar()
			return fmt.Errorf("failed to acquire lock on icon_tags_table#%s: %w", iconName, lockErr)
		}
		defer repo.iconTagsLockClient.ReleaseLock(tagsLock)

		updateErr := repo.updateTag(ctx, *oldItem, false)
		if updateErr != nil {
			rollbackTagsUpdatedSoFar()
			return fmt.Errorf("failed to update tags for icon %s about to be deleted: %w", iconName, updateErr)
		}
		tagsUpdatedWithDecrRefCount = append(tagsUpdatedWithDecrRefCount, oldItem)
	}

	deleteErr := repo.deleteIcon(ctx, iconItem)
	if deleteErr != nil {
		return fmt.Errorf("failed to delete icon %s: %w", iconName, deleteErr)
	}

	if createSideEffect != nil {
		sideEffectErr := createSideEffect()
		if sideEffectErr != nil {
			rollbackTagsUpdatedSoFar()
			rollbackErr := repo.updateIcon(ctx, iconItem)
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
	logger := zerolog.Ctx(ctx).With().Str("method", "DynamodbRepository.DeleteIconfile").Str("iconName", iconName).Str("iconfile", iconfile.String()).Logger()

	readDataLock, _ := repo.iconsLockClient.Get(iconName)
	logger.Debug().Str("currentLockData", string(readDataLock.Data())).Msg("about to acquire lock")

	lock, lockErr := repo.iconsLockClient.AcquireLockWithContext(ctx, iconName, dynamolock.WithData([]byte("DeleteIconfile")), dynamolock.ReplaceData())
	if lockErr != nil {
		lock, readDataErr := repo.iconsLockClient.Get(iconName)
		if readDataErr != nil {
			logger.Error().Err(readDataErr).Str("lockTable", "icons").Str("lockKey", iconName).Msg("failed to read lock data after failing to acquire lock")
		}
		return fmt.Errorf("failed to acquire lock on icons_table#%s (<%s): %w", iconName, string(lock.Data()), lockErr)
	}
	defer func() {
		success, releaseLockErr := repo.iconsLockClient.ReleaseLock(lock)
		if releaseLockErr != nil {
			logger.Error().Err(releaseLockErr).Msg("error while releasing lock on icons table")
		}
		if success {
			logger.Debug().Str("lockTable", "icons").Str("lockKey", iconName).Bool("lockReleaseResult", success).Send()
		} else {
			logger.Info().Str("lockTable", "icons").Str("lockKey", iconName).Bool("lockReleaseResult", success).Send()
		}
	}()

	logger.Debug().Msg("about to call repo.getIconItem...")
	oldIconItem, getIconItemErr := repo.getIconItem(ctx, iconName, false)
	if getIconItemErr != nil {
		return fmt.Errorf("failed to get DyndbIcon to remove %s: %w", iconName, getIconItemErr)
	}

	newIconItem := DyndbIcon{
		IconName:   iconName,
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
		deleteIconErr := repo.deleteIconNoLock(ctx, iconName, modifiedBy, createSideEffect)
		if deleteIconErr != nil {
			return fmt.Errorf("failed to delete icon (with no more iconfiles left) %s: %w", iconName, deleteIconErr)
		}
		return nil
	}
	newIconItem.Iconfiles = newIconfiles

	updateErr := repo.updateIcon(ctx, &newIconItem)
	if updateErr != nil {
		return fmt.Errorf("failed to update icon %s with %v removed: %w", iconName, iconfile, updateErr)
	}

	if createSideEffect != nil {
		sideEffectErr := createSideEffect()
		if sideEffectErr != nil {
			rollbackErr := repo.updateIcon(ctx, oldIconItem)
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
	logger.Debug().Str("iconName", iconName).Msg("BEGIN")

	icon := &DyndbIcon{IconName: iconName}
	key, keyErr := icon.GetKey(ctx)
	if keyErr != nil {
		return icon, keyErr
	}

	input := &aws_dyndb.GetItemInput{
		TableName:      aws.String(IconsTableName),
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

func (repo *DynamodbRepository) updateIcon(ctx context.Context, icon *DyndbIcon) error {
	logger := zerolog.Ctx(ctx).With().Str("method", "DynamodbRepository.updateIcon").Logger()

	newItem, marshalErr := attributevalue.MarshalMap(icon)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal new icon item for updating %s: %w", icon.IconName, marshalErr)
	}

	input := &aws_dyndb.PutItemInput{
		TableName:              aws.String(IconsTableName),
		Item:                   newItem,
		ReturnConsumedCapacity: "TOTAL",
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

	input := &aws_dyndb.DeleteItemInput{
		TableName:              aws.String(IconsTableName),
		Key:                    pk,
		ReturnConsumedCapacity: "TOTAL",
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
		TableName:      aws.String(IconTagsTableName),
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

func (repo *DynamodbRepository) updateTag(ctx context.Context, tagItem DyndbTag, add bool) error {
	var newRefCount int64
	if add {
		newRefCount = tagItem.ReferenceCount + 1
	} else {
		newRefCount = tagItem.ReferenceCount - 1
	}

	if newRefCount > 0 {
		newItem, marshalErr := attributevalue.MarshalMap(&DyndbTag{tagItem.Tag, newRefCount})
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal new icon item for updating %s: %w", tagItem.Tag, marshalErr)
		}

		input := &aws_dyndb.PutItemInput{
			TableName:              aws.String(IconTagsTableName),
			Item:                   newItem,
			ReturnConsumedCapacity: "TOTAL",
		}
		_, err := repo.awsClient.PutItem(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to update icon %s: %w", tagItem.Tag, Unwrap(ctx, err))
		}

		return nil
	}

	pk, getKeyErr := tagItem.GetKey(ctx)
	if getKeyErr != nil {
		return fmt.Errorf("failed to generate PK of %v for removal: %w", tagItem, getKeyErr)
	}

	input := &aws_dyndb.DeleteItemInput{
		TableName:              aws.String(IconTagsTableName),
		Key:                    pk,
		ReturnConsumedCapacity: "TOTAL",
	}
	_, deleteErr := repo.awsClient.DeleteItem(ctx, input)
	if deleteErr != nil {
		return fmt.Errorf("failed to delete tag %v: %w", tagItem, deleteErr)
	}

	return nil
}
