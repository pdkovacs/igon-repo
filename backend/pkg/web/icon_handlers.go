package web

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authr"
	"github.com/pdkovacs/igo-repo/backend/pkg/services"
	log "github.com/sirupsen/logrus"
)

func describeAllIconsHanler(iconService *services.IconService) func(c *gin.Context) {
	return func(c *gin.Context) {
		logger := log.WithField("prefix", "createIconHandler")
		icons, err := iconService.DescribeAllIcons()
		if err != nil {
			logger.Errorf("%v", err)
			c.AbortWithStatus(500)
		}
		c.JSON(200, icons)
	}
}

func createIconHandler(iconService *services.IconService) func(c *gin.Context) {
	return func(c *gin.Context) {
		logger := log.WithField("prefix", "createIconHandler")

		r := c.Request
		r.ParseMultipartForm(32 << 20) // limit your max input length to 32MB

		iconName := r.FormValue("iconName")
		if len(iconName) == 0 {
			logger.Infof("invalid icon name: <empty-string>")
			c.AbortWithStatus(400)
			return
		}
		logger.Debugf("icon name: %s", iconName)

		var buf bytes.Buffer

		// in your case file would be fileupload
		file, header, err := r.FormFile("iconfile")
		if err != nil {
			logger.Infof("failed to retrieve iconfile for icon %s: %v", iconName, err)
			c.AbortWithStatus(400)
			return
		}
		defer file.Close()

		name := strings.Split(header.Filename, ".")
		logger.Infof("File name %s\n", name[0])

		// Copy the file data to my buffer
		io.Copy(&buf, file)
		logger.Infof("received %d bytes for icon %s", buf.Len(), iconName)

		// do something with the contents...
		icon, errCreate := iconService.CreateIcon(iconName, buf.Bytes(), MustGetUserSession(c).UserInfo)
		if errCreate != nil {
			logger.Errorf("failed to create icon %v", errCreate)
			if errors.Is(errCreate, authr.ErrPermission) {
				c.AbortWithStatusJSON(403, icon)
				return
			} else {
				c.AbortWithStatusJSON(500, icon)
			}
		}
		c.JSON(201, icon)

		buf.Reset()
		// do something else
		// etc write header
		return
	}
}
