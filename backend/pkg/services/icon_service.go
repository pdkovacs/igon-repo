package services

import (
	"encoding/base64"
	"fmt"
	"image"
	_ "image/png"
	"strings"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	log "github.com/sirupsen/logrus"
)

func CreateIcon(iconName string, initialIconfileContent []byte, modifiedBy string) domain.Iconfile {
	logger := log.WithField("prefix", "CreateIcon")
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
	return iconfile
}
