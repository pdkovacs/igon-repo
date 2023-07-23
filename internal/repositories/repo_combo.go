package repositories

import (
	"context"
	"fmt"

	"iconrepo/internal/app/domain"
	"iconrepo/internal/app/security/authn"
	"iconrepo/internal/app/security/authr"
)

type IndexRepository interface {
	DescribeAllIcons(ctx context.Context) ([]domain.IconDescriptor, error)
	DescribeIcon(ctx context.Context, iconName string) (domain.IconDescriptor, error)
	GetExistingTags(tx context.Context) ([]string, error)
	CreateIcon(ctx context.Context, iconName string, iconfile domain.IconfileDescriptor, modifiedBy string, createSideEffect func() error) error
	AddIconfileToIcon(ctx context.Context, iconName string, iconfile domain.IconfileDescriptor, modifiedBy string, createSideEffect func() error) error
	AddTag(ctx context.Context, iconName string, tag string, modifiedBy string) error
	RemoveTag(ctx context.Context, iconName string, tag string, modifiedBy string) error
	DeleteIcon(ctx context.Context, iconName string, modifiedBy string, createSideEffect func() error) error
	DeleteIconfile(ctx context.Context, iconName string, iconfile domain.IconfileDescriptor, modifiedBy string, createSideEffect func() error) error
}

type BlobstoreRepository interface {
	fmt.Stringer
	CreateRepository() error
	AddIconfile(iconName string, iconfile domain.Iconfile, modifiedBy string) error
	GetIconfile(iconName string, iconfile domain.IconfileDescriptor) ([]byte, error)
	DeleteIcon(ctx context.Context, iconDesc domain.IconDescriptor, modifiedBy authn.UserID) error
	DeleteIconfile(ctx context.Context, iconName string, iconfileDesc domain.IconfileDescriptor, modifiedBy authn.UserID) error
}

type RepoCombo struct {
	Index     IndexRepository
	Blobstore BlobstoreRepository
}

func (combo *RepoCombo) DescribeAllIcons(ctx context.Context) ([]domain.IconDescriptor, error) {
	return combo.Index.DescribeAllIcons(ctx)
}

func (combo *RepoCombo) DescribeIcon(ctx context.Context, iconName string) (domain.IconDescriptor, error) {
	return combo.Index.DescribeIcon(ctx, iconName)
}

func (combo *RepoCombo) CreateIcon(ctx context.Context, iconName string, iconfile domain.Iconfile, modifiedBy authr.UserInfo) error {
	return combo.Index.CreateIcon(ctx, iconName, iconfile.IconfileDescriptor, modifiedBy.UserId.String(), func() error {
		return combo.Blobstore.AddIconfile(iconName, iconfile, modifiedBy.UserId.String())
	})
}

func (combo *RepoCombo) DeleteIcon(ctx context.Context, iconName string, modifiedBy authr.UserInfo) error {
	iconDesc, describeErr := combo.Index.DescribeIcon(ctx, iconName)
	if describeErr != nil {
		return fmt.Errorf("failed to have to-be-deleted icon \"%s\" described: %w", iconName, describeErr)
	}

	return combo.Index.DeleteIcon(ctx, iconName, modifiedBy.UserId.String(), func() error {
		return combo.Blobstore.DeleteIcon(ctx, iconDesc, modifiedBy.UserId)
	})
}

func (combo *RepoCombo) AddIconfile(ctx context.Context, iconName string, iconfile domain.Iconfile, modifiedBy authr.UserInfo) error {
	return combo.Index.AddIconfileToIcon(ctx, iconName, iconfile.IconfileDescriptor, modifiedBy.UserId.String(), func() error {
		return combo.Blobstore.AddIconfile(iconName, iconfile, modifiedBy.UserId.String())
	})
}

func (combo *RepoCombo) GetIconfile(iconName string, iconfile domain.IconfileDescriptor) ([]byte, error) {
	return combo.Blobstore.GetIconfile(iconName, iconfile)
}

func (combo *RepoCombo) DeleteIconfile(ctx context.Context, iconName string, iconfile domain.IconfileDescriptor, modifiedBy authr.UserInfo) error {
	return combo.Index.DeleteIconfile(ctx, iconName, iconfile, modifiedBy.UserId.String(), func() error {
		return combo.Blobstore.DeleteIconfile(ctx, iconName, iconfile, modifiedBy.UserId)
	})
}

func (combo *RepoCombo) GetTags(ctx context.Context) ([]string, error) {
	return combo.Index.GetExistingTags(ctx)
}

func (combo *RepoCombo) AddTag(ctx context.Context, iconName string, tag string, modifiedBy authr.UserInfo) error {
	return combo.Index.AddTag(ctx, iconName, tag, modifiedBy.UserId.String())
}

func (combo *RepoCombo) RemoveTag(ctx context.Context, iconName string, tag string, modifiedBy authr.UserInfo) error {
	return combo.Index.RemoveTag(ctx, iconName, tag, modifiedBy.UserId.String())
}
