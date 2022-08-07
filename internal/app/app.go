package app

import (
	"igo-repo/internal/app/domain"
	"igo-repo/internal/app/security/authr"
	"igo-repo/internal/app/services"

	"github.com/rs/zerolog"
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

type App struct {
	Repository Repository
}

func (app *App) GetAPI(logger zerolog.Logger) *api {
	return &api{
		IconService: services.NewIconService(app.Repository, logger),
	}
}
