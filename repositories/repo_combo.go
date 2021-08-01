package repositories

import (
	"fmt"

	"github.com/pdkovacs/igo-repo/app/domain"
	"github.com/pdkovacs/igo-repo/app/security/authr"
)

type RepoCombo struct {
	DB  *DatabaseRepository
	Git *GitRepository
}

func (combo *RepoCombo) DescribeAllIcons() ([]domain.IconDescriptor, error) {
	return combo.DB.DescribeAllIcons()
}

func (combo *RepoCombo) DescribeIcon(iconName string) (domain.IconDescriptor, error) {
	return combo.DB.DescribeIcon(iconName)
}

func (combo *RepoCombo) CreateIcon(iconName string, iconfile domain.Iconfile, modifiedBy authr.UserInfo) error {
	return combo.DB.CreateIcon(iconName, iconfile, modifiedBy.UserId.String(), func() error {
		return combo.Git.AddIconfile(iconName, iconfile, modifiedBy.UserId.String())
	})
}

func (combo *RepoCombo) DeleteIcon(iconName string, modifiedBy authr.UserInfo) error {
	iconDesc, describeErr := combo.DB.DescribeIcon(iconName)
	if describeErr != nil {
		return fmt.Errorf("failed to have to-be-deleted icon \"%s\" described: %w", iconName, describeErr)
	}

	return combo.DB.DeleteIcon(iconName, modifiedBy.UserId.String(), func() error {
		return combo.Git.DeleteIcon(iconDesc, modifiedBy.UserId)
	})
}

func (combo *RepoCombo) AddIconfile(iconName string, iconfile domain.Iconfile, modifiedBy authr.UserInfo) error {
	return combo.DB.AddIconfileToIcon(iconName, iconfile, modifiedBy.UserId.String(), func() error {
		return combo.Git.AddIconfile(iconName, iconfile, modifiedBy.UserId.String())
	})
}

func (combo *RepoCombo) GetIconFile(iconName string, iconfile domain.IconfileDescriptor) ([]byte, error) {
	return combo.DB.GetIconFile(iconName, iconfile)
}

func (combo *RepoCombo) DeleteIconfile(iconName string, iconfile domain.IconfileDescriptor, modifiedBy authr.UserInfo) error {
	return combo.DB.DeleteIconfile(iconName, iconfile, modifiedBy.UserId.String(), func() error {
		return combo.Git.DeleteIconfile(iconName, iconfile, modifiedBy.UserId)
	})
}

func (combo *RepoCombo) GetTags() ([]string, error) {
	return combo.DB.GetExistingTags()
}

func (combo *RepoCombo) AddTag(iconName string, tag string, modifiedBy authr.UserInfo) error {
	return combo.DB.AddTag(iconName, tag, modifiedBy.UserId.String())
}

func (combo *RepoCombo) RemoveTag(iconName string, tag string, modifiedBy authr.UserInfo) error {
	return combo.DB.RemoveTag(iconName, tag, modifiedBy.UserId.String())
}
