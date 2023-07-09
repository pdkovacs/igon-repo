package app

import (
	"iconrepo/internal/app/domain"
	"iconrepo/internal/app/security/authr"
	"iconrepo/internal/app/services"
)

type iconService interface {
	DescribeAllIcons() ([]domain.IconDescriptor, error)
	DescribeIcon(iconName string) (domain.IconDescriptor, error)
	CreateIcon(iconName string, initialIconfileContent []byte, modifiedBy authr.UserInfo) (domain.Icon, error)
	GetIconfile(iconName string, iconfile domain.IconfileDescriptor) ([]byte, error)
	AddIconfile(iconName string, initialIconfileContent []byte, modifiedBy authr.UserInfo) (domain.IconfileDescriptor, error)
	DeleteIcon(iconName string, modifiedBy authr.UserInfo) error
	DeleteIconfile(iconName string, iconfileDescriptor domain.IconfileDescriptor, modifiedBy authr.UserInfo) error
	GetTags() ([]string, error)
	AddTag(iconName string, tag string, userInfo authr.UserInfo) error
	RemoveTag(iconName string, tag string, userInfo authr.UserInfo) error
}

// Primary port
type api struct {
	IconService iconService
}

// Secondary port
type Repository = services.Repository

type AppCore struct {
	Repository Repository
}

func (app *AppCore) GetAPI() *api {
	return &api{
		IconService: services.NewIconService(app.Repository),
	}
}
