package repositories

import (
	"fmt"

	"iconrepo/internal/app/domain"
	"iconrepo/internal/app/security/authn"
	"iconrepo/internal/app/security/authr"
)

type IndexRepository interface {
	DescribeAllIcons() ([]domain.IconDescriptor, error)
	DescribeIcon(iconName string) (domain.IconDescriptor, error)
	GetExistingTags() ([]string, error)
	CreateIcon(iconName string, iconfile domain.IconfileDescriptor, modifiedBy string, createSideEffect func() error) error
	AddIconfileToIcon(iconName string, iconfile domain.IconfileDescriptor, modifiedBy string, createSideEffect func() error) error
	AddTag(iconName string, tag string, modifiedBy string) error
	RemoveTag(iconName string, tag string, modifiedBy string) error
	DeleteIcon(iconName string, modifiedBy string, createSideEffect func() error) error
	DeleteIconfile(iconName string, iconfile domain.IconfileDescriptor, modifiedBy string, createSideEffect func() error) error
}

type BlobstoreRepository interface {
	fmt.Stringer
	Create() error
	AddIconfile(iconName string, iconfile domain.Iconfile, modifiedBy string) error
	GetIconfile(iconName string, iconfile domain.IconfileDescriptor) ([]byte, error)
	DeleteIcon(iconDesc domain.IconDescriptor, modifiedBy authn.UserID) error
	DeleteIconfile(iconName string, iconfileDesc domain.IconfileDescriptor, modifiedBy authn.UserID) error
}

type RepoCombo struct {
	Index     IndexRepository
	Blobstore BlobstoreRepository
}

func (combo *RepoCombo) DescribeAllIcons() ([]domain.IconDescriptor, error) {
	return combo.Index.DescribeAllIcons()
}

func (combo *RepoCombo) DescribeIcon(iconName string) (domain.IconDescriptor, error) {
	return combo.Index.DescribeIcon(iconName)
}

func (combo *RepoCombo) CreateIcon(iconName string, iconfile domain.Iconfile, modifiedBy authr.UserInfo) error {
	return combo.Index.CreateIcon(iconName, iconfile.IconfileDescriptor, modifiedBy.UserId.String(), func() error {
		return combo.Blobstore.AddIconfile(iconName, iconfile, modifiedBy.UserId.String())
	})
}

func (combo *RepoCombo) DeleteIcon(iconName string, modifiedBy authr.UserInfo) error {
	iconDesc, describeErr := combo.Index.DescribeIcon(iconName)
	if describeErr != nil {
		return fmt.Errorf("failed to have to-be-deleted icon \"%s\" described: %w", iconName, describeErr)
	}

	return combo.Index.DeleteIcon(iconName, modifiedBy.UserId.String(), func() error {
		return combo.Blobstore.DeleteIcon(iconDesc, modifiedBy.UserId)
	})
}

func (combo *RepoCombo) AddIconfile(iconName string, iconfile domain.Iconfile, modifiedBy authr.UserInfo) error {
	return combo.Index.AddIconfileToIcon(iconName, iconfile.IconfileDescriptor, modifiedBy.UserId.String(), func() error {
		return combo.Blobstore.AddIconfile(iconName, iconfile, modifiedBy.UserId.String())
	})
}

func (combo *RepoCombo) GetIconfile(iconName string, iconfile domain.IconfileDescriptor) ([]byte, error) {
	return combo.Blobstore.GetIconfile(iconName, iconfile)
}

func (combo *RepoCombo) DeleteIconfile(iconName string, iconfile domain.IconfileDescriptor, modifiedBy authr.UserInfo) error {
	return combo.Index.DeleteIconfile(iconName, iconfile, modifiedBy.UserId.String(), func() error {
		return combo.Blobstore.DeleteIconfile(iconName, iconfile, modifiedBy.UserId)
	})
}

func (combo *RepoCombo) GetTags() ([]string, error) {
	return combo.Index.GetExistingTags()
}

func (combo *RepoCombo) AddTag(iconName string, tag string, modifiedBy authr.UserInfo) error {
	return combo.Index.AddTag(iconName, tag, modifiedBy.UserId.String())
}

func (combo *RepoCombo) RemoveTag(iconName string, tag string, modifiedBy authr.UserInfo) error {
	return combo.Index.RemoveTag(iconName, tag, modifiedBy.UserId.String())
}
