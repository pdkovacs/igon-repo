package repositories

import (
	"fmt"

	"github.com/pdkovacs/igo-repo/app/domain"
	"github.com/pdkovacs/igo-repo/app/security/authr"
)

type RepoCombo struct {
	DescribeAllIcons func() ([]domain.IconDescriptor, error)
	DescribeIcon     func(iconName string) (domain.IconDescriptor, error)
	CreateIcon       func(iconName string, iconfile domain.Iconfile, modifiedBy authr.UserInfo) error
	DeleteIcon       func(iconName string, modifiedBy authr.UserInfo) error

	GetIconFile    func(iconName string, iconfile domain.IconfileDescriptor) ([]byte, error)
	AddIconfile    func(iconName string, iconfile domain.Iconfile, modifiedBy authr.UserInfo) error
	DeleteIconfile func(iconName string, iconfile domain.IconfileDescriptor, modifiedBy authr.UserInfo) error

	GetTags   func() ([]string, error)
	AddTag    func(iconName string, tag string, modifiedBy authr.UserInfo) error
	RemoveTag func(iconName string, tag string, modifiedBy authr.UserInfo) error
}

func NewRepoCombo(db DatabaseRepository, git GitRepository) RepoCombo {
	createIcon := func(iconName string, iconfile domain.Iconfile, modifiedBy authr.UserInfo) error {
		return db.CreateIcon(iconName, iconfile, modifiedBy.UserId.String(), func() error {
			return git.AddIconfile(iconName, iconfile, modifiedBy.UserId.String())
		})
	}

	deleteIcon := func(iconName string, modifiedBy authr.UserInfo) error {
		iconDesc, describeErr := db.DescribeIcon(iconName)
		if describeErr != nil {
			return fmt.Errorf("failed to have to-be-deleted icon \"%s\" described: %w", iconName, describeErr)
		}

		return db.DeleteIcon(iconName, modifiedBy.UserId.String(), func() error {
			return git.DeleteIcon(iconDesc, modifiedBy.UserId)
		})
	}

	addIconfile := func(iconName string, iconfile domain.Iconfile, modifiedBy authr.UserInfo) error {
		return db.AddIconfileToIcon(iconName, iconfile, modifiedBy.UserId.String(), func() error {
			return git.AddIconfile(iconName, iconfile, modifiedBy.UserId.String())
		})
	}

	deleteIconfile := func(iconName string, iconfile domain.IconfileDescriptor, modifiedBy authr.UserInfo) error {
		return db.DeleteIconfile(iconName, iconfile, modifiedBy.UserId.String(), func() error {
			return git.DeleteIconfile(iconName, iconfile, modifiedBy.UserId)
		})
	}

	addTag := func(iconName string, tag string, modifiedBy authr.UserInfo) error {
		return db.AddTag(iconName, tag, modifiedBy.UserId.String())
	}

	removeTag := func(iconName string, tag string, modifiedBy authr.UserInfo) error {
		return db.RemoveTag(iconName, tag, modifiedBy.UserId.String())
	}

	return RepoCombo{
		DescribeAllIcons: db.DescribeAllIcons,
		DescribeIcon:     db.DescribeIcon,
		CreateIcon:       createIcon,
		DeleteIcon:       deleteIcon,
		GetIconFile:      db.GetIconFile,
		AddIconfile:      addIconfile,
		DeleteIconfile:   deleteIconfile,
		GetTags:          db.GetExistingTags,
		AddTag:           addTag,
		RemoveTag:        removeTag,
	}
}
