package services

import (
	"encoding/base64"
	"fmt"
	"image"
	_ "image/png"
	"strings"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authr"
	log "github.com/sirupsen/logrus"
)

func CreateIcon(iconName string, initialIconfileContent []byte, modifiedBy UserInfo) (domain.Iconfile, error) {
	logger := log.WithField("prefix", "CreateIcon")
	err := authr.HasRequiredPermissions(modifiedBy.UserId, modifiedBy.Permissions, []authr.PermissionID{
		authr.CREATE_ICON,
	})
	if err != nil {
		return domain.Iconfile{}, fmt.Errorf("failed to create icon %v: %w", iconName, err)
	}
	logger.Infof("iconName: %s, initialIconfileContent: %v, modifiedBy: %s", iconName, string(initialIconfileContent), modifiedBy)
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(string(initialIconfileContent)))
	config, format, err := image.DecodeConfig(reader)
	if err != nil {
		log.Fatal(err)
	}
	iconfile := domain.Iconfile{
		Format: format,
		Size:   fmt.Sprintf("%dpx", config.Height),
	}
	logger.Infof(
		"iconName: %s, iconfile: %v, initialIconfileContent size: %d, modifiedBy: %s",
		iconName, iconfile, len(initialIconfileContent), modifiedBy,
	)
	return iconfile, nil
}
