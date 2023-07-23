package app

import (
	"context"
	"iconrepo/internal/app/domain"
	"iconrepo/internal/app/security/authr"
	"iconrepo/internal/app/services"
)

type iconService interface {
	DescribeAllIcons(ctx context.Context) ([]domain.IconDescriptor, error)
	DescribeIcon(ctx context.Context, iconName string) (domain.IconDescriptor, error)
	CreateIcon(ctx context.Context, iconName string, initialIconfileContent []byte, modifiedBy authr.UserInfo) (domain.Icon, error)
	GetIconfile(iconName string, iconfile domain.IconfileDescriptor) ([]byte, error)
	AddIconfile(ctx context.Context, iconName string, initialIconfileContent []byte, modifiedBy authr.UserInfo) (domain.IconfileDescriptor, error)
	DeleteIcon(ctx context.Context, iconName string, modifiedBy authr.UserInfo) error
	DeleteIconfile(ctx context.Context, iconName string, iconfileDescriptor domain.IconfileDescriptor, modifiedBy authr.UserInfo) error
	GetTags(ctx context.Context) ([]string, error)
	AddTag(ctx context.Context, iconName string, tag string, userInfo authr.UserInfo) error
	RemoveTag(ctx context.Context, iconName string, tag string, userInfo authr.UserInfo) error
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
