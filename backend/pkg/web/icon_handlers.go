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

const iconRootPath = "/icon"

type IconPath struct {
	domain.IconfileDescriptor
	Path string
}

type ResponseIcon struct {
	Name       string
	ModifiedBy string
	Paths      []IconPath
	Tags       []string
}

func createIconfilePath(baseUrl string, iconName string, iconfileDescriptor domain.IconfileDescriptor) string {
	return fmt.Sprintf("%s/%s/format/%s/size/%s", baseUrl, iconName, iconfileDescriptor.Format, iconfileDescriptor.Size)
}

func createIconPath(baseUrl string, iconName string, iconfileDescriptor domain.IconfileDescriptor) IconPath {
	return IconPath{
		IconfileDescriptor: domain.IconfileDescriptor{
			Format: iconfileDescriptor.Format,
			Size:   iconfileDescriptor.Size,
		},
		Path: createIconfilePath(baseUrl, iconName, iconfileDescriptor),
	}
}

func createIconfilePaths(baseUrl string, iconDesc domain.IconDescriptor) []IconPath {
	iconPaths := []IconPath{}
	for _, iconfileDescriptor := range iconDesc.Iconfiles {
		iconPaths = append(iconPaths, createIconPath(baseUrl, iconDesc.Name, iconfileDescriptor))
	}
	return iconPaths
}

func createResponseIcon(iconPathRoot string, iconDesc domain.IconDescriptor) ResponseIcon {
	return ResponseIcon{
		Name:       iconDesc.Name,
		ModifiedBy: iconDesc.ModifiedBy,
		Paths:      createIconfilePaths(iconPathRoot, iconDesc),
		Tags:       iconDesc.Tags,
	}
}

func iconfilesToIconfileDescriptors(iconfiles []domain.Iconfile) []domain.IconfileDescriptor {
	iconfileDescriptors := []domain.IconfileDescriptor{}
	for _, iconfile := range iconfiles {
		iconfileDescriptors = append(iconfileDescriptors, iconfile.IconfileDescriptor)
	}
	return iconfileDescriptors
}

func iconToResponseIcon(icon domain.Icon) ResponseIcon {
	return createResponseIcon(
		iconRootPath,
		domain.IconDescriptor{
			IconAttributes: icon.IconAttributes,
			Iconfiles:      iconfilesToIconfileDescriptors(icon.Iconfiles),
		},
	)
}

func describeAllIconsHanler(iconService *services.IconService) func(c *gin.Context) {
	return func(c *gin.Context) {
		logger := log.WithField("prefix", "createIconHandler")
		icons, err := iconService.DescribeAllIcons()
		if err != nil {
			logger.Errorf("%v", err)
			c.AbortWithStatus(500)
		}
		responseIcon := []ResponseIcon{}
		for _, icon := range icons {
			responseIcon = append(responseIcon, createResponseIcon(iconRootPath, icon))
		}
		c.JSON(200, responseIcon)
	}
}

func describeIconHandler(iconService *services.IconService) func(c *gin.Context) {
	return func(c *gin.Context) {
		logger := log.WithField("prefix", "createIconHandler")
		iconName := c.Param("name")
		icon, err := iconService.DescribeIcon(iconName)
		if err != nil {
			logger.Errorf("%v", err)
			if errors.Is(err, domain.ErrIconNotFound) {
				c.AbortWithStatus(404)
				return
			}
			c.AbortWithStatus(500)
			return
		}
		responseIcon := createResponseIcon(iconRootPath, icon)
		c.JSON(200, responseIcon)
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
		defer buf.Reset()
		logger.Infof("received %d bytes for icon %s", buf.Len(), iconName)

		// do something with the contents...
		icon, errCreate := iconService.CreateIcon(iconName, buf.Bytes(), MustGetUserSession(c).UserInfo)
		if errCreate != nil {
			logger.Errorf("failed to create icon %v", errCreate)
			if errors.Is(errCreate, authr.ErrPermission) {
				c.AbortWithStatusJSON(403, iconToResponseIcon(icon))
				return
			} else {
				c.AbortWithStatusJSON(500, iconToResponseIcon(icon))
				return
			}
		}
		c.JSON(201, iconToResponseIcon(icon))

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
		iconfileDescriptor, errCreate := iconService.AddIconfile(iconName, buf.Bytes(), MustGetUserSession(c).UserInfo)
		if errCreate != nil {
			logger.Errorf("failed to add iconfile %v", errCreate)
			if errors.Is(errCreate, authr.ErrPermission) {
				c.AbortWithStatus(403)
				return
			} else {
				c.AbortWithStatus(500)
			}
		}
		c.JSON(201, createIconPath(iconRootPath, iconName, iconfileDescriptor))

		buf.Reset()
		// do something else
		// etc write header
		return
	}
}

func deleteIconHandler(iconService *services.IconService) func(c *gin.Context) {
	logger := log.WithField("prefix", "deleteIconHandler")
	return func(c *gin.Context) {
		session := MustGetUserSession(c)
		errPerm := authr.HasRequiredPermissions(session.UserInfo.UserId, session.UserInfo.Permissions, []authr.PermissionID{authr.REMOVE_ICON})
		if errPerm != nil {
			c.AbortWithStatus(403)
		}
		iconName := c.Param("name")
		deletionErr := iconService.DeleteIcon(iconName, session.UserInfo)
		if deletionErr != nil {
			logger.Errorf("failed to delete icon \"%s\": %v", iconName, deletionErr)
			c.AbortWithStatus(500)
			return
		}
		c.Status(204)
	}
}
