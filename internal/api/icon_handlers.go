package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pdkovacs/igo-repo/internal/domain"
	"github.com/pdkovacs/igo-repo/internal/security/authr"
	"github.com/pdkovacs/igo-repo/internal/services"
	log "github.com/sirupsen/logrus"
)

const iconRootPath = "/icon"

type IconPath struct {
	domain.IconfileDescriptor
	Path string `json:"path"`
}

type ResponseIcon struct {
	Name       string     `json:"name"`
	ModifiedBy string     `json:"modifiedBy"`
	Paths      []IconPath `json:"paths"`
	Tags       []string   `json:"tags"`
}

func createIconfilePath(baseUrl string, iconName string, iconfileDescriptor domain.IconfileDescriptor) string {
	return fmt.Sprintf("%s/%s/format/%s/size/%s", baseUrl, iconName, iconfileDescriptor.Format, iconfileDescriptor.Size)
}

func CreateIconPath(baseUrl string, iconName string, iconfileDescriptor domain.IconfileDescriptor) IconPath {
	return IconPath{
		IconfileDescriptor: domain.IconfileDescriptor{
			Format: iconfileDescriptor.Format,
			Size:   iconfileDescriptor.Size,
		},
		Path: createIconfilePath(baseUrl, iconName, iconfileDescriptor),
	}
}

func CreateIconfilePaths(baseUrl string, iconDesc domain.IconDescriptor) []IconPath {
	iconPaths := []IconPath{}
	for _, iconfileDescriptor := range iconDesc.Iconfiles {
		iconPaths = append(iconPaths, CreateIconPath(baseUrl, iconDesc.Name, iconfileDescriptor))
	}
	return iconPaths
}

func CreateResponseIcon(iconPathRoot string, iconDesc domain.IconDescriptor) ResponseIcon {
	return ResponseIcon{
		Name:       iconDesc.Name,
		ModifiedBy: iconDesc.ModifiedBy,
		Paths:      CreateIconfilePaths(iconPathRoot, iconDesc),
		Tags:       iconDesc.Tags,
	}
}

func IconfilesToIconfileDescriptors(iconfiles []domain.Iconfile) []domain.IconfileDescriptor {
	iconfileDescriptors := []domain.IconfileDescriptor{}
	for _, iconfile := range iconfiles {
		iconfileDescriptors = append(iconfileDescriptors, iconfile.IconfileDescriptor)
	}
	return iconfileDescriptors
}

func iconToResponseIcon(icon domain.Icon) ResponseIcon {
	return CreateResponseIcon(
		iconRootPath,
		domain.IconDescriptor{
			IconAttributes: icon.IconAttributes,
			Iconfiles:      IconfilesToIconfileDescriptors(icon.Iconfiles),
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
			responseIcon = append(responseIcon, CreateResponseIcon(iconRootPath, icon))
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
		responseIcon := CreateResponseIcon(iconRootPath, icon)
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
				c.AbortWithStatus(403)
				return
			} else {
				c.AbortWithStatus(500)
				return
			}
		}
		c.JSON(201, iconToResponseIcon(icon))
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

		session := MustGetUserSession(c)
		authrErr := authr.HasRequiredPermissions(
			session.UserInfo.UserId,
			session.UserInfo.Permissions,
			[]authr.PermissionID{
				authr.UPDATE_ICON,
				authr.ADD_ICONFILE,
			})
		if authrErr != nil {
			c.AbortWithStatus(403)
			return
		}

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
			} else if errors.Is(errCreate, domain.ErrIconfileAlreadyExists) {
				c.AbortWithStatus(409)
				return
			} else {
				c.AbortWithStatus(500)
			}
		}
		c.JSON(200, CreateIconPath(iconRootPath, iconName, iconfileDescriptor))

		buf.Reset()
	}
}

func deleteIconHandler(iconService *services.IconService) func(c *gin.Context) {
	logger := log.WithField("prefix", "deleteIconHandler")
	return func(c *gin.Context) {
		session := MustGetUserSession(c)
		iconName := c.Param("name")
		deleteError := iconService.DeleteIcon(iconName, session.UserInfo)
		if deleteError != nil {
			if errors.Is(deleteError, authr.ErrPermission) {
				c.AbortWithStatus(403)
				return
			}
			logger.Errorf("failed to delete icon \"%s\": %v", iconName, deleteError)
			c.AbortWithStatus(500)
			return
		}
		c.Status(204)
	}
}

func deleteIconfileHandler(iconService *services.IconService) func(c *gin.Context) {
	logger := log.WithField("prefix", "delteIconfileHandler")
	return func(c *gin.Context) {
		session := MustGetUserSession(c)
		iconName := c.Param("name")
		format := c.Param("format")
		size := c.Param("size")
		iconfileDescriptor := domain.IconfileDescriptor{Format: format, Size: size}
		deleteError := iconService.DeleteIconfile(iconName, iconfileDescriptor, session.UserInfo)
		if deleteError != nil {
			if errors.Is(deleteError, authr.ErrPermission) {
				c.AbortWithStatus(403)
				return
			}
			if errors.Is(deleteError, domain.ErrIconNotFound) {
				logger.Infof("Icon %s not found", iconName)
				c.AbortWithStatus(404)
				return
			}
			if errors.Is(deleteError, domain.ErrIconfileNotFound) {
				logger.Infof("Iconfile %v of %s not found", iconfileDescriptor, iconName)
				c.AbortWithStatus(404)
				return
			}
			logger.Errorf("failed to delete iconfile %v of \"%s\": %v", iconfileDescriptor, iconName, deleteError)
			c.AbortWithStatus(500)
			return
		}
		c.Status(204)
	}
}

func getTagsHandler(iconService *services.IconService) func(c *gin.Context) {
	logger := log.WithField("prefix", "getTagsHandler")
	return func(c *gin.Context) {
		tags, serviceError := iconService.GetTags()
		if serviceError != nil {
			logger.Errorf("Failed to retrieve tags: %v", serviceError)
			c.AbortWithStatus(500)
			return
		}
		c.JSON(200, tags)
	}
}

type AddServiceRequestData struct {
	Tag string `json:"tag"`
}

func addTagHandler(iconService *services.IconService) func(c *gin.Context) {
	logger := log.WithField("prefix", "addTagHandler")
	return func(c *gin.Context) {
		session := MustGetUserSession(c)
		iconName := c.Param("name")

		jsonData, readBodyErr := io.ReadAll(c.Request.Body)
		if readBodyErr != nil {
			logger.Errorf("failed to read body: %v", readBodyErr)
			c.AbortWithStatus(400)
			return
		}
		tagRequestData := AddServiceRequestData{}
		json.Unmarshal(jsonData, &tagRequestData)
		tag := tagRequestData.Tag

		serviceError := iconService.AddTag(iconName, tag, session.UserInfo)
		if serviceError != nil {
			if errors.Is(serviceError, authr.ErrPermission) {
				logger.Infof("Icon %s not found to add/remove tag %s to/from: %v", iconName, tag, serviceError)
				c.AbortWithStatus(403)
				return
			}
			if errors.Is(serviceError, domain.ErrIconNotFound) {
				logger.Infof("Icon %s not found to add/remove tag %s to/from: %v", iconName, tag, serviceError)
				c.AbortWithStatus(404)
				return
			}
			logger.Errorf("Failed to add/remove tag %s to/from %s: %v", tag, iconName, serviceError)
			c.AbortWithStatus(500)
			return
		}
		c.Status(201)
	}
}

func removeTagHandler(iconService *services.IconService) func(c *gin.Context) {
	logger := log.WithField("prefix", "removeTagHandler")
	return func(c *gin.Context) {
		session := MustGetUserSession(c)
		iconName := c.Param("name")
		tag := c.Param("tag")
		serviceError := iconService.RemoveTag(iconName, tag, session.UserInfo)
		if serviceError != nil {
			if errors.Is(serviceError, authr.ErrPermission) {
				logger.Infof("Icon %s not found to add/remove tag %s to/from: %v", iconName, tag, serviceError)
				c.AbortWithStatus(403)
				return
			}
			if errors.Is(serviceError, domain.ErrIconNotFound) {
				logger.Infof("Icon %s not found to add/remove tag %s to/from: %v", iconName, tag, serviceError)
				c.AbortWithStatus(404)
				return
			}
			logger.Errorf("Failed to add/remove tag %s to/from %s: %v", tag, iconName, serviceError)
			c.AbortWithStatus(500)
			return
		}
		c.Status(204)
	}
}
