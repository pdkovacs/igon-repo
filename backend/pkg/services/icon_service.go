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
	service.Repositories.DB.CreateIcon(iconName, iconfile, modifiedBy.UserId.String(), func() error {
		return service.Repositories.Git.AddIconfile(iconName, iconfile, modifiedBy.UserId.String())
	})
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

func (service *IconService) AddIconfile(iconName string, initialIconfileContent []byte, modifiedBy UserInfo) (domain.Iconfile, error) {
	logger := log.WithField("prefix", "AddIconfile")
	err := authr.HasRequiredPermissions(modifiedBy.UserId, modifiedBy.Permissions, []authr.PermissionID{
		authr.CREATE_ICON,
	})
	if err != nil {
		return domain.Iconfile{}, fmt.Errorf("failed to create icon %v: %w", iconName, err)
	}
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
		"iconName: %s, iconfile: %v, content of iconfile to add size: %d, modifiedBy: %s",
		iconName, iconfile, len(initialIconfileContent), modifiedBy,
	)
	service.Repositories.DB.AddIconfileToIcon(iconName, iconfile, modifiedBy.UserId.String(), func() error {
		return service.Repositories.Git.AddIconfile(iconName, iconfile, modifiedBy.UserId.String())
	})
	return iconfile, nil
}
