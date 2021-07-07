package services

import (
	"encoding/base64"
	"fmt"
	"image"
	_ "image/png"
	"strings"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/repositories"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authr"
	log "github.com/sirupsen/logrus"
)

type IconService struct {
	Repositories *repositories.Repositories
}

func (server *IconService) DescribeAllIcons() ([]domain.IconDescriptor, error) {
	icons, err := server.Repositories.DB.DescribeAllIcons()
	if err != nil {
		return []domain.IconDescriptor{}, fmt.Errorf("failed to describe all icons: %w", err)
	}
	return icons, err
}

func (server *IconService) DescribeIcon(iconName string) (domain.IconDescriptor, error) {
	icon, err := server.Repositories.DB.DescribeIcon(iconName)
	if err != nil {
		return domain.IconDescriptor{}, fmt.Errorf("failed to describe icon \"%s\": %w", iconName, err)
	}
	return icon, err
}

func (service *IconService) CreateIcon(iconName string, initialIconfileContent []byte, modifiedBy UserInfo) (domain.Icon, error) {
	logger := log.WithField("prefix", "CreateIcon")
	err := authr.HasRequiredPermissions(modifiedBy.UserId, modifiedBy.Permissions, []authr.PermissionID{
		authr.CREATE_ICON,
	})
	if err != nil {
		return domain.Icon{}, fmt.Errorf("failed to create icon %v: %w", iconName, err)
	}
	logger.Infof("iconName: %s, initialIconfileContent: %v, modifiedBy: %s", iconName, string(initialIconfileContent), modifiedBy)
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(string(initialIconfileContent)))
	config, format, err := image.DecodeConfig(reader)
	if err != nil {
		log.Fatal(err)
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

	errCreate := service.Repositories.DB.CreateIcon(iconName, iconfile, modifiedBy.UserId.String(), func() error {
		return service.Repositories.Git.AddIconfile(iconName, iconfile, modifiedBy.UserId.String())
	})
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
	content, err := service.Repositories.DB.GetIconFile(iconName, iconfile.Format, iconfile.Size)
	if err != nil {
		return domain.Iconfile{}, fmt.Errorf("failed to retrieve iconfile %v: %w", iconfile, err)
	}
	return domain.Iconfile{
		IconfileDescriptor: iconfile,
		Content:            content,
	}, nil
}

func (service *IconService) AddIconfile(iconName string, initialIconfileContent []byte, modifiedBy UserInfo) (domain.IconfileDescriptor, error) {
	logger := log.WithField("prefix", "AddIconfile")
	err := authr.HasRequiredPermissions(modifiedBy.UserId, modifiedBy.Permissions, []authr.PermissionID{
		authr.UPDATE_ICON,
		authr.ADD_ICONFILE,
	})
	if err != nil {
		return domain.IconfileDescriptor{}, fmt.Errorf("failed to add iconfile %v: %w", iconName, err)
	}
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(string(initialIconfileContent)))
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
	errAddIconfile := service.Repositories.DB.AddIconfileToIcon(iconName, iconfile, modifiedBy.UserId.String(), func() error {
		return service.Repositories.Git.AddIconfile(iconName, iconfile, modifiedBy.UserId.String())
	})
	if errAddIconfile != nil {
		return domain.IconfileDescriptor{}, errAddIconfile
	}

	return iconfile.IconfileDescriptor, nil
}

func (service *IconService) DeleteIcon(iconName string, modifiedBy UserInfo) error {
	err := authr.HasRequiredPermissions(modifiedBy.UserId, modifiedBy.Permissions, []authr.PermissionID{
		authr.REMOVE_ICON,
	})
	if err != nil {
		return fmt.Errorf("not enough permissions to delete icon \"%v\" to : %w", iconName, err)
	}
	iconDesc, describeErr := service.Repositories.DB.DescribeIcon(iconName)
	if describeErr != nil {
		return fmt.Errorf("failed to have to-be-deleted icon \"%s\" described: %w", iconName, describeErr)
	}
	errDeleteIcon := service.Repositories.DB.DeleteIcon(iconName, modifiedBy.UserId.String(), func() error {
		return service.Repositories.Git.DeleteIcon(iconDesc, modifiedBy.UserId)
	})
	return errDeleteIcon
}

func (service *IconService) DeleteIconfile(iconName string, iconfileDescriptor domain.IconfileDescriptor, modifiedBy UserInfo) error {
	err := authr.HasRequiredPermissions(modifiedBy.UserId, modifiedBy.Permissions, []authr.PermissionID{
		authr.REMOVE_ICONFILE,
	})
	if err != nil {
		return fmt.Errorf("not enough permissions to delete icon \"%v\" to : %w", iconName, err)
	}
	errDeleteIcon := service.Repositories.DB.DeleteIconfile(iconName, iconfileDescriptor, modifiedBy.UserId.String(), func() error {
		return service.Repositories.Git.DeleteIconfile(iconName, iconfileDescriptor, modifiedBy.UserId)
	})
	return errDeleteIcon
}

func (service *IconService) GetTags() ([]string, error) {
	return service.Repositories.DB.GetExistingTags()
}

func (service *IconService) AddTag(iconName string, tag string, userInfo UserInfo) error {
	permErr := authr.HasRequiredPermissions(userInfo.UserId, userInfo.Permissions, []authr.PermissionID{authr.ADD_TAG})
	if permErr != nil {
		return authr.ErrPermission
	}
	dbErr := service.Repositories.DB.AddTag(iconName, tag, userInfo.UserId.String())
	if dbErr != nil {
		return fmt.Errorf("failed to add tag %s to \"%s\": %w", tag, iconName, dbErr)
	}
	return nil
}

func (service *IconService) RemoveTag(iconName string, tag string, userInfo UserInfo) error {
	permErr := authr.HasRequiredPermissions(userInfo.UserId, userInfo.Permissions, []authr.PermissionID{authr.REMOVE_TAG})
	if permErr != nil {
		return authr.ErrPermission
	}
	dbErr := service.Repositories.DB.RemoveTag(iconName, tag, userInfo.UserId.String())
	if dbErr != nil {
		return fmt.Errorf("failed to remove tag %s from \"%s\": %w", tag, iconName, dbErr)
	}
	return nil
}

func init() {
	registerSVGDecoder()
}
