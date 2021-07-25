package app

import (
	"github.com/pdkovacs/igo-repo/app/domain"
	"github.com/pdkovacs/igo-repo/app/security/authr"
)

type API struct {
	DescribeAllIcons func() ([]domain.IconDescriptor, error)
	DescribeIcon     func(iconName string) (domain.IconDescriptor, error)
	CreateIcon       func(iconName string, initialIconfileContent []byte, modifiedBy authr.UserInfo) (domain.Icon, error)
	DeleteIcon       func(iconName string, modifiedBy authr.UserInfo) error

	GetIconfile    func(iconName string, iconfile domain.IconfileDescriptor) (domain.Iconfile, error)
	AddIconfile    func(iconName string, iconfileContent []byte, modifiedBy authr.UserInfo) (domain.IconfileDescriptor, error)
	DeleteIconfile func(iconName string, iconfileDescriptor domain.IconfileDescriptor, modifiedBy authr.UserInfo) error

	GetTags   func() ([]string, error)
	AddTag    func(iconName string, tag string, userInfo authr.UserInfo) error
	RemoveTag func(iconName string, tag string, userInfo authr.UserInfo) error
}

type Repository interface {
	DescribeAllIcons() ([]domain.IconDescriptor, error)
	DescribeIcon(iconName string) (domain.IconDescriptor, error)
	CreateIcon(iconName string, iconfile domain.Iconfile, modifiedBy authr.UserInfo) error
	DeleteIcon(iconName string, modifiedBy authr.UserInfo) error

	GetIconFile(iconName string, iconfile domain.IconfileDescriptor) ([]byte, error)
	AddIconfile(iconName string, iconfile domain.Iconfile, modifiedBy authr.UserInfo) error
	DeleteIconfile(iconName string, iconfile domain.IconfileDescriptor, modifiedBy authr.UserInfo) error

	GetTags() ([]string, error)
	AddTag(iconName string, tag string, modifiedBy authr.UserInfo) error
	RemoveTag(iconName string, tag string, modifiedBy authr.UserInfo) error
}

type App struct {
	API                API
	RegisterRepository func(repository Repository)
}
