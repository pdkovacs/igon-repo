package web

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authr"
	"github.com/pdkovacs/igo-repo/backend/pkg/services"
	log "github.com/sirupsen/logrus"
)

type IconPath struct {
	domain.IconfileDescriptor
	path string
}

type IconDTO struct {
	name       string
	modifiedBy string
	paths      []IconPath
	tags       []string
}

func createIconfilePath(baseUrl string, iconName string, iconfileDescriptor domain.IconfileDescriptor) string {
	return fmt.Sprintf("%s/%s/format/%s/size/%s", baseUrl, iconName, iconfileDescriptor.Format, iconfileDescriptor.Size)
}

func createIconfilePaths(baseUrl string, iconDesc domain.IconDescriptor) []IconPath {
	iconPaths := []IconPath{}
	for _, iconfileDescriptor := range iconDesc.Iconfiles {
		iconPath := IconPath{
			IconfileDescriptor: domain.IconfileDescriptor{
				Format: iconfileDescriptor.Format,
				Size:   iconfileDescriptor.Size,
			},
			path: createIconfilePath(baseUrl, iconDesc.Name, iconfileDescriptor),
		}
		iconPaths = append(iconPaths, iconPath)
	}
	return iconPaths
}

func createIconDTO(iconPathRoot string, iconDesc domain.IconDescriptor) IconDTO {
	return IconDTO{
		name:       iconDesc.Name,
		modifiedBy: iconDesc.ModifiedBy,
		paths:      createIconfilePaths(iconPathRoot, iconDesc),
		tags:       iconDesc.Tags,
	}
}

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

func getIconfileHandler(iconService *services.IconService) func(c *gin.Context) {
	return func(c *gin.Context) {
		logger := log.WithField("prefix", "getIconfileHandler")
		iconName := c.Param("name")
		format := c.Param("format")
		size := c.Param("size")
		iconFile, err := iconService.GetIconfile(iconName, domain.IconfileDescriptor{
			Format: format,
			Size:   size,
		})
		if err != nil {
			logger.Errorf("failed to retrieve %s:%scontents for icon %s: %v", iconName, size, format, err)
			c.AbortWithStatus(500)
		}
		c.JSON(200, iconFile)
	}
}

func addIconfileHandler(iconService *services.IconService) func(c *gin.Context) {
	return func(c *gin.Context) {
		logger := log.WithField("prefix", "addIconfileHandler")

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
		logger.Infof("received %d bytes as iconfile content for icon %s", buf.Len(), iconName)

		// do something with the contents...
		icon, errCreate := iconService.AddIconfile(iconName, buf.Bytes(), MustGetUserSession(c).UserInfo)
		if errCreate != nil {
			logger.Errorf("failed to add iconfile %v", errCreate)
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
