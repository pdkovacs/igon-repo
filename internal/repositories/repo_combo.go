package repositories

import (
	"fmt"

	"iconrepo/internal/app/domain"
	"iconrepo/internal/app/security/authn"
	"iconrepo/internal/app/security/authr"
)

type DBRepository interface {
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

type GitRepository interface {
	fmt.Stringer
	Create() error
	AddIconfile(iconName string, iconfile domain.Iconfile, modifiedBy string) error
	GetIconfile(iconName string, iconfile domain.IconfileDescriptor) ([]byte, error)
	DeleteIcon(iconDesc domain.IconDescriptor, modifiedBy authn.UserID) error
	DeleteIconfile(iconName string, iconfileDesc domain.IconfileDescriptor, modifiedBy authn.UserID) error
}

type RepoCombo struct {
	Db  DBRepository
	Git GitRepository
}

func (combo *RepoCombo) DescribeAllIcons() ([]domain.IconDescriptor, error) {
	return combo.Db.DescribeAllIcons()
}

func (combo *RepoCombo) DescribeIcon(iconName string) (domain.IconDescriptor, error) {
	return combo.Db.DescribeIcon(iconName)
}

func (combo *RepoCombo) CreateIcon(iconName string, iconfile domain.Iconfile, modifiedBy authr.UserInfo) error {
	return combo.Db.CreateIcon(iconName, iconfile.IconfileDescriptor, modifiedBy.UserId.String(), func() error {
		return combo.Git.AddIconfile(iconName, iconfile, modifiedBy.UserId.String())
	})
}

func (combo *RepoCombo) DeleteIcon(iconName string, modifiedBy authr.UserInfo) error {
	iconDesc, describeErr := combo.Db.DescribeIcon(iconName)
	if describeErr != nil {
		return fmt.Errorf("failed to have to-be-deleted icon \"%s\" described: %w", iconName, describeErr)
	}

	return combo.Db.DeleteIcon(iconName, modifiedBy.UserId.String(), func() error {
		return combo.Git.DeleteIcon(iconDesc, modifiedBy.UserId)
	})
}

func (combo *RepoCombo) AddIconfile(iconName string, iconfile domain.Iconfile, modifiedBy authr.UserInfo) error {
	return combo.Db.AddIconfileToIcon(iconName, iconfile.IconfileDescriptor, modifiedBy.UserId.String(), func() error {
		return combo.Git.AddIconfile(iconName, iconfile, modifiedBy.UserId.String())
	})
}

func (combo *RepoCombo) GetIconfile(iconName string, iconfile domain.IconfileDescriptor) ([]byte, error) {
	return combo.Git.GetIconfile(iconName, iconfile)
}

func (combo *RepoCombo) DeleteIconfile(iconName string, iconfile domain.IconfileDescriptor, modifiedBy authr.UserInfo) error {
	return combo.Db.DeleteIconfile(iconName, iconfile, modifiedBy.UserId.String(), func() error {
		return combo.Git.DeleteIconfile(iconName, iconfile, modifiedBy.UserId)
	})
}

func (combo *RepoCombo) GetTags() ([]string, error) {
	return combo.Db.GetExistingTags()
}

func (combo *RepoCombo) AddTag(iconName string, tag string, modifiedBy authr.UserInfo) error {
	return combo.Db.AddTag(iconName, tag, modifiedBy.UserId.String())
}

func (combo *RepoCombo) RemoveTag(iconName string, tag string, modifiedBy authr.UserInfo) error {
	return combo.Db.RemoveTag(iconName, tag, modifiedBy.UserId.String())
}
