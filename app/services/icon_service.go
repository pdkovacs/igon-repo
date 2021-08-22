package services

import (
	"bytes"
	"fmt"
	"image"

	"github.com/pdkovacs/igo-repo/app/domain"
	"github.com/pdkovacs/igo-repo/app/security/authr"
	log "github.com/sirupsen/logrus"
)

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

type IconService struct {
	Repository Repository
}

func (service *IconService) DescribeAllIcons() ([]domain.IconDescriptor, error) {
	icons, err := service.Repository.DescribeAllIcons()
	if err != nil {
		return []domain.IconDescriptor{}, fmt.Errorf("failed to describe all icons: %w", err)
	}
	return icons, err
}

func (service *IconService) DescribeIcon(iconName string) (domain.IconDescriptor, error) {
	icon, err := service.Repository.DescribeIcon(iconName)
	if err != nil {
		return domain.IconDescriptor{}, fmt.Errorf("failed to describe icon \"%s\": %w", iconName, err)
	}
	return icon, err
}

func (service *IconService) CreateIcon(iconName string, initialIconfileContent []byte, modifiedBy authr.UserInfo) (domain.Icon, error) {
	logger := log.WithField("prefix", "CreateIcon")
	err := authr.HasRequiredPermissions(modifiedBy.UserId, modifiedBy.Permissions, []authr.PermissionID{
		authr.CREATE_ICON,
	})
	if err != nil {
		return domain.Icon{}, fmt.Errorf("failed to create icon %v: %w", iconName, err)
	}
	logger.Infof("iconName: %s, initialIconfileContent: %v encoded bytes, modifiedBy: %s", iconName, len(initialIconfileContent), modifiedBy)
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
	logger.Infof(
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

func (service *IconService) GetIconfile(iconName string, iconfile domain.IconfileDescriptor) (domain.Iconfile, error) {
	content, err := service.Repository.GetIconFile(iconName, iconfile)
	if err != nil {
		return domain.Iconfile{}, fmt.Errorf("failed to retrieve iconfile %v: %w", iconfile, err)
	}
	return domain.Iconfile{
		IconfileDescriptor: iconfile,
		Content:            content,
	}, nil
}

func (service *IconService) AddIconfile(iconName string, initialIconfileContent []byte, modifiedBy authr.UserInfo) (domain.IconfileDescriptor, error) {
	logger := log.WithField("prefix", "AddIconfile")
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
		logger.Errorf("failed to decode image configuration of iconfile for %s: %v", iconName, err)
		return domain.IconfileDescriptor{}, fmt.Errorf("failed to decode image configuration of iconfile for %s: %w", iconName, err)
	}
	iconfile := domain.Iconfile{
		IconfileDescriptor: domain.IconfileDescriptor{
			Format: format,
			Size:   fmt.Sprintf("%dpx", config.Height),
		},
		Content: initialIconfileContent,
	}
	logger.Infof(
		"iconName: %s, iconfile: %v, content of iconfile to add size: %d, modifiedBy: %s",
		iconName, iconfile, len(initialIconfileContent), modifiedBy,
	)
	errAddIconfile := service.Repository.AddIconfile(iconName, iconfile, modifiedBy)
	if errAddIconfile != nil {
		return domain.IconfileDescriptor{}, errAddIconfile
	}

	return iconfile.IconfileDescriptor, nil
}

func (service *IconService) DeleteIcon(iconName string, modifiedBy authr.UserInfo) error {
	err := authr.HasRequiredPermissions(modifiedBy.UserId, modifiedBy.Permissions, []authr.PermissionID{
		authr.REMOVE_ICON,
	})
	if err != nil {
		return fmt.Errorf("not enough permissions to delete icon \"%v\" to : %w", iconName, err)
	}
	return service.Repository.DeleteIcon(iconName, modifiedBy)
}

func (service *IconService) DeleteIconfile(iconName string, iconfileDescriptor domain.IconfileDescriptor, modifiedBy authr.UserInfo) error {
	err := authr.HasRequiredPermissions(modifiedBy.UserId, modifiedBy.Permissions, []authr.PermissionID{
		authr.REMOVE_ICONFILE,
	})
	if err != nil {
		return fmt.Errorf("not enough permissions to delete icon \"%v\" to : %w", iconName, err)
	}
	return service.Repository.DeleteIconfile(iconName, iconfileDescriptor, modifiedBy)
}

func (service *IconService) GetTags() ([]string, error) {
	return service.Repository.GetTags()
}

func (service *IconService) AddTag(iconName string, tag string, userInfo authr.UserInfo) error {
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

func (service *IconService) RemoveTag(iconName string, tag string, userInfo authr.UserInfo) error {
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

func init() {
	RegisterSVGDecoder()
}
