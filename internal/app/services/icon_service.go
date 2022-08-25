package services

import (
	"bytes"
	"fmt"
	"image"

	"igo-repo/internal/app/domain"
	"igo-repo/internal/app/security/authr"
	"igo-repo/internal/logging"

	"github.com/rs/zerolog"
)

type Repository interface {
	DescribeAllIcons() ([]domain.IconDescriptor, error)
	DescribeIcon(iconName string) (domain.IconDescriptor, error)
	CreateIcon(iconName string, iconfile domain.Iconfile, modifiedBy authr.UserInfo) error
	DeleteIcon(iconName string, modifiedBy authr.UserInfo) error

	GetIconfile(iconName string, iconfile domain.IconfileDescriptor) ([]byte, error)
	AddIconfile(iconName string, iconfile domain.Iconfile, modifiedBy authr.UserInfo) error
	DeleteIconfile(iconName string, iconfile domain.IconfileDescriptor, modifiedBy authr.UserInfo) error

	GetTags() ([]string, error)
	AddTag(iconName string, tag string, modifiedBy authr.UserInfo) error
	RemoveTag(iconName string, tag string, modifiedBy authr.UserInfo) error
}

type iconService struct {
	Repository Repository
	logger     zerolog.Logger
}

func NewIconService(repo Repository, logger zerolog.Logger) *iconService {
	RegisterSVGDecoder(logging.CreateUnitLogger(logger, "svg-decoder"))
	return &iconService{
		Repository: repo,
		logger:     logger,
	}
}

func (service *iconService) DescribeAllIcons() ([]domain.IconDescriptor, error) {
	icons, err := service.Repository.DescribeAllIcons()
	if err != nil {
		return []domain.IconDescriptor{}, fmt.Errorf("failed to describe all icons: %w", err)
	}
	return icons, err
}

func (service *iconService) DescribeIcon(iconName string) (domain.IconDescriptor, error) {
	icon, err := service.Repository.DescribeIcon(iconName)
	if err != nil {
		return domain.IconDescriptor{}, fmt.Errorf("failed to describe icon \"%s\": %w", iconName, err)
	}
	return icon, err
}

func (service *iconService) CreateIcon(iconName string, initialIconfileContent []byte, modifiedBy authr.UserInfo) (domain.Icon, error) {
	logger := logging.CreateMethodLogger(service.logger, "CreateIcon")
	err := authr.HasRequiredPermissions(modifiedBy.UserId, modifiedBy.Permissions, []authr.PermissionID{
		authr.CREATE_ICON,
	})
	if err != nil {
		return domain.Icon{}, fmt.Errorf("failed to create icon %v: %w", iconName, err)
	}
	logger.Info().Msgf("iconName: %s, initialIconfileContent: %v encoded bytes, modifiedBy: %s", iconName, len(initialIconfileContent), modifiedBy)
	config, format, err := image.DecodeConfig(bytes.NewReader(initialIconfileContent))
	if err != nil {
		return domain.Icon{}, fmt.Errorf("failed to decode iconfile: %w", err)
	}
	iconfile := domain.Iconfile{
		IconfileDescriptor: domain.IconfileDescriptor{
			Format: format,
			Size:   fmt.Sprintf("%dpx", config.Height),
		},
		Content: initialIconfileContent,
	}
	logger.Info().Msgf(
		"iconName: %s, iconfile: %v, initialIconfileContent size: %d, modifiedBy: %s",
		iconName, iconfile, len(initialIconfileContent), modifiedBy,
	)

	errCreate := service.Repository.CreateIcon(iconName, iconfile, modifiedBy)
	if errCreate != nil {
		return domain.Icon{}, errCreate
	}

	return domain.Icon{
		IconAttributes: domain.IconAttributes{
			Name:       iconName,
			ModifiedBy: modifiedBy.UserId.String(),
			Tags:       []string{},
		},
		Iconfiles: []domain.Iconfile{
			iconfile,
		},
	}, nil
}

func (service *iconService) GetIconfile(iconName string, iconfile domain.IconfileDescriptor) ([]byte, error) {
	content, err := service.Repository.GetIconfile(iconName, iconfile)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve iconfile %v: %w", iconfile, err)
	}
	return content, nil
}

func (service *iconService) AddIconfile(iconName string, initialIconfileContent []byte, modifiedBy authr.UserInfo) (domain.IconfileDescriptor, error) {
	logger := logging.CreateMethodLogger(service.logger, "AddIconfile")
	err := authr.HasRequiredPermissions(modifiedBy.UserId, modifiedBy.Permissions, []authr.PermissionID{
		authr.UPDATE_ICON,
		authr.ADD_ICONFILE,
	})
	if err != nil {
		return domain.IconfileDescriptor{}, fmt.Errorf("failed to add iconfile %v: %w", iconName, err)
	}
	reader := bytes.NewReader(initialIconfileContent)
	config, format, err := image.DecodeConfig(reader)
	if err != nil {
		logger.Error().Msgf("failed to decode image configuration of iconfile for %s: %v", iconName, err)
		return domain.IconfileDescriptor{}, fmt.Errorf("failed to decode image configuration of iconfile for %s: %w", iconName, err)
	}
	iconfile := domain.Iconfile{
		IconfileDescriptor: domain.IconfileDescriptor{
			Format: format,
			Size:   fmt.Sprintf("%dpx", config.Height),
		},
		Content: initialIconfileContent,
	}
	logger.Info().Msgf(
		"iconName: %s, iconfile: %v, content of iconfile to add size: %d, modifiedBy: %s",
		iconName, iconfile, len(initialIconfileContent), modifiedBy,
	)
	errAddIconfile := service.Repository.AddIconfile(iconName, iconfile, modifiedBy)
	if errAddIconfile != nil {
		return domain.IconfileDescriptor{}, errAddIconfile
	}

	return iconfile.IconfileDescriptor, nil
}

func (service *iconService) DeleteIcon(iconName string, modifiedBy authr.UserInfo) error {
	err := authr.HasRequiredPermissions(modifiedBy.UserId, modifiedBy.Permissions, []authr.PermissionID{
		authr.REMOVE_ICON,
	})
	if err != nil {
		return fmt.Errorf("not enough permissions to delete icon \"%v\" to : %w", iconName, err)
	}
	return service.Repository.DeleteIcon(iconName, modifiedBy)
}

func (service *iconService) DeleteIconfile(iconName string, iconfileDescriptor domain.IconfileDescriptor, modifiedBy authr.UserInfo) error {
	err := authr.HasRequiredPermissions(modifiedBy.UserId, modifiedBy.Permissions, []authr.PermissionID{
		authr.REMOVE_ICONFILE,
	})
	if err != nil {
		return fmt.Errorf("not enough permissions to delete icon \"%v\" to : %w", iconName, err)
	}
	return service.Repository.DeleteIconfile(iconName, iconfileDescriptor, modifiedBy)
}

func (service *iconService) GetTags() ([]string, error) {
	return service.Repository.GetTags()
}

func (service *iconService) AddTag(iconName string, tag string, userInfo authr.UserInfo) error {
	permErr := authr.HasRequiredPermissions(userInfo.UserId, userInfo.Permissions, []authr.PermissionID{authr.ADD_TAG})
	if permErr != nil {
		return authr.ErrPermission
	}
	dbErr := service.Repository.AddTag(iconName, tag, userInfo)
	if dbErr != nil {
		return fmt.Errorf("failed to add tag %s to \"%s\": %w", tag, iconName, dbErr)
	}
	return nil
}

func (service *iconService) RemoveTag(iconName string, tag string, userInfo authr.UserInfo) error {
	permErr := authr.HasRequiredPermissions(userInfo.UserId, userInfo.Permissions, []authr.PermissionID{authr.REMOVE_TAG})
	if permErr != nil {
		return authr.ErrPermission
	}
	dbErr := service.Repository.RemoveTag(iconName, tag, userInfo)
	if dbErr != nil {
		return fmt.Errorf("failed to remove tag %s from \"%s\": %w", tag, iconName, dbErr)
	}
	return nil
}
