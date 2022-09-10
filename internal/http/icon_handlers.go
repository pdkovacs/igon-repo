package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"igo-repo/internal/app/domain"
	"igo-repo/internal/app/security/authn"
	"igo-repo/internal/app/security/authr"
	"igo-repo/internal/app/services"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

const iconRootPath = "/icon"

type IconPath struct {
	domain.IconfileDescriptor
	Path string `json:"path"`
}

type IconDTO struct {
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

func CreateResponseIcon(iconPathRoot string, iconDesc domain.IconDescriptor) IconDTO {
	return IconDTO{
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

func iconToResponseIcon(icon domain.Icon) IconDTO {
	return CreateResponseIcon(
		iconRootPath,
		domain.IconDescriptor{
			IconAttributes: icon.IconAttributes,
			Iconfiles:      IconfilesToIconfileDescriptors(icon.Iconfiles),
		},
	)
}

func describeAllIconsHanler(describeAllIcons func() ([]domain.IconDescriptor, error), logger zerolog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		icons, err := describeAllIcons()
		if err != nil {
			logger.Error().Msgf("%v", err)
			c.AbortWithStatus(500)
		}
		responseIcon := []IconDTO{}
		for _, icon := range icons {
			responseIcon = append(responseIcon, CreateResponseIcon(iconRootPath, icon))
		}
		c.JSON(200, responseIcon)
	}
}

func describeIconHandler(describeIcon func(iconName string) (domain.IconDescriptor, error), logger zerolog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		iconName := c.Param("name")
		icon, err := describeIcon(iconName)
		if err != nil {
			logger.Error().Msgf("%v", err)
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

func createIconHandler(
	getUserInfo func(c *gin.Context) authr.UserInfo,
	createIcon func(iconName string, initialIconfileContent []byte, modifiedBy authr.UserInfo) (domain.Icon, error),
	publish func(msg services.NotificationMessage, initiator authn.UserID),
	logger zerolog.Logger,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		r := c.Request
		r.ParseMultipartForm(32 << 20) // limit your max input length to 32MB

		iconName := r.FormValue("iconName")
		if len(iconName) == 0 {
			logger.Info().Msgf("invalid icon name: <empty-string>")
			c.AbortWithStatus(400)
			return
		}
		logger.Debug().Msgf("icon name: %s", iconName)

		var buf bytes.Buffer

		// in your case file would be fileupload
		file, header, err := r.FormFile("iconfile")
		if err != nil {
			logger.Info().Msgf("failed to retrieve iconfile for icon %s: %v", iconName, err)
			c.AbortWithStatus(400)
			return
		}
		defer file.Close()

		name := strings.Split(header.Filename, ".")
		logger.Info().Msgf("File name %s\n", name[0])

		// Copy the file data to my buffer
		io.Copy(&buf, file)
		defer buf.Reset()
		logger.Info().Msgf("received %d bytes for icon %s", buf.Len(), iconName)

		authorInfo := getUserInfo(c)
		// do something with the contents...
		icon, errCreate := createIcon(iconName, buf.Bytes(), authorInfo)
		if errCreate != nil {
			logger.Error().Msgf("failed to create icon %v", errCreate)
			if errors.Is(errCreate, authr.ErrPermission) {
				c.AbortWithStatus(403)
				return
			} else if errors.Is(errCreate, domain.ErrIconAlreadyExists) {
				c.AbortWithStatus(409)
				return
			} else {
				c.AbortWithStatus(500)
				return
			}
		}
		c.JSON(201, iconToResponseIcon(icon))
		publish(services.NotifMsgIconCreated, authorInfo.UserId)
	}
}

func getIconfileHandler(getIconfile func(iconName string, iconfile domain.IconfileDescriptor) ([]byte, error), logger zerolog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		iconName := c.Param("name")
		format := c.Param("format")
		size := c.Param("size")
		iconfile, err := getIconfile(iconName, domain.IconfileDescriptor{
			Format: format,
			Size:   size,
		})
		if err != nil {
			if errors.Is(err, domain.ErrIconfileNotFound) {
				c.AbortWithStatus(404)
				return
			}
			logger.Error().Msgf("failed to retrieve %s:%scontents for icon %s: %v", iconName, size, format, err)
			c.AbortWithStatus(500)
		}
		if format == "svg" {
			c.Data(200, "image/svg+xml", iconfile)
			return
		}
		c.Data(200, "application/octet-stream", iconfile)
	}
}

func addIconfileHandler(
	getUserInfo func(c *gin.Context) authr.UserInfo,
	addIconfile func(iconName string, initialIconfileContent []byte, modifiedBy authr.UserInfo) (domain.IconfileDescriptor, error),
	publish func(msg services.NotificationMessage, initiator authn.UserID),
	logger zerolog.Logger,
) func(c *gin.Context) {

	return func(c *gin.Context) {
		authorInfo := getUserInfo(c)

		authrErr := authr.HasRequiredPermissions(
			authorInfo,
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
			logger.Info().Msgf("invalid icon name: <empty-string>")
			c.AbortWithStatus(400)
			return
		}
		logger.Debug().Msgf("icon name: %s", iconName)

		var buf bytes.Buffer

		// in your case file would be fileupload
		file, header, err := r.FormFile("iconfile")
		if err != nil {
			logger.Info().Msgf("failed to retrieve iconfile for icon %s: %v", iconName, err)
			c.AbortWithStatus(400)
			return
		}
		defer file.Close()

		name := strings.Split(header.Filename, ".")
		logger.Info().Msgf("File name %s\n", name[0])

		// Copy the file data to my buffer
		io.Copy(&buf, file)
		logger.Info().Msgf("received %d bytes as iconfile content for icon %s", buf.Len(), iconName)

		// do something with the contents...
		iconfileDescriptor, errCreate := addIconfile(iconName, buf.Bytes(), authorInfo)
		if errCreate != nil {
			logger.Error().Msgf("failed to add iconfile %v", errCreate)
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
		publish(services.NotifMsgIconfileAdded, authorInfo.UserId)
		buf.Reset()
	}
}

func deleteIconHandler(
	getUserInfo func(c *gin.Context) authr.UserInfo,
	deleteIcon func(iconName string, modifiedBy authr.UserInfo) error,
	publish func(msg services.NotificationMessage, initiator authn.UserID),
	logger zerolog.Logger,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		authorInfo := getUserInfo(c)
		iconName := c.Param("name")
		deleteError := deleteIcon(iconName, authorInfo)
		if deleteError != nil {
			if errors.Is(deleteError, authr.ErrPermission) {
				c.AbortWithStatus(403)
				return
			}
			logger.Error().Msgf("failed to delete icon \"%s\": %v", iconName, deleteError)
			c.AbortWithStatus(500)
			return
		}
		c.Status(204)
		publish(services.NotifMsgIconDeleted, authorInfo.UserId)
	}
}

func deleteIconfileHandler(
	getUserInfo func(c *gin.Context) authr.UserInfo,
	deleteIconfile func(iconName string, iconfile domain.IconfileDescriptor, modifiedBy authr.UserInfo) error,
	publish func(msg services.NotificationMessage, initiator authn.UserID),
	logger zerolog.Logger,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		authorInfo := getUserInfo(c)
		iconName := c.Param("name")
		format := c.Param("format")
		size := c.Param("size")
		iconfileDescriptor := domain.IconfileDescriptor{Format: format, Size: size}
		deleteError := deleteIconfile(iconName, iconfileDescriptor, authorInfo)
		if deleteError != nil {
			if errors.Is(deleteError, authr.ErrPermission) {
				c.AbortWithStatus(403)
				return
			}
			if errors.Is(deleteError, domain.ErrIconNotFound) {
				logger.Info().Msgf("Icon %s not found", iconName)
				c.AbortWithStatus(404)
				return
			}
			if errors.Is(deleteError, domain.ErrIconfileNotFound) {
				logger.Info().Msgf("Iconfile %v of %s not found", iconfileDescriptor, iconName)
				c.AbortWithStatus(404)
				return
			}
			logger.Error().Msgf("failed to delete iconfile %v of \"%s\": %v", iconfileDescriptor, iconName, deleteError)
			c.AbortWithStatus(500)
			return
		}
		c.Status(204)
		publish(services.NotifMsgIconfileDeleted, authorInfo.UserId)
	}
}

func getTagsHandler(getTags func() ([]string, error), logger zerolog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		tags, serviceError := getTags()
		if serviceError != nil {
			logger.Error().Msgf("Failed to retrieve tags: %v", serviceError)
			c.AbortWithStatus(500)
			return
		}
		c.JSON(200, tags)
	}
}

type AddServiceRequestData struct {
	Tag string `json:"tag"`
}

func addTagHandler(
	getUserInfo func(c *gin.Context) authr.UserInfo,
	addTag func(iconName string, tag string, modifiedBy authr.UserInfo) error,
	logger zerolog.Logger,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		userInfo := getUserInfo(c)
		iconName := c.Param("name")

		jsonData, readBodyErr := io.ReadAll(c.Request.Body)
		if readBodyErr != nil {
			logger.Error().Msgf("failed to read body: %v", readBodyErr)
			c.AbortWithStatus(400)
			return
		}
		tagRequestData := AddServiceRequestData{}
		json.Unmarshal(jsonData, &tagRequestData)
		tag := tagRequestData.Tag
		if len(tag) == 0 {
			logger.Error().Msg("Tags must be at last one character long")
			c.AbortWithStatus(400)
			return
		}

		serviceError := addTag(iconName, tag, userInfo)
		if serviceError != nil {
			if errors.Is(serviceError, authr.ErrPermission) {
				logger.Info().Msgf("Icon %s not found to add/remove tag %s to/from: %v", iconName, tag, serviceError)
				c.AbortWithStatus(403)
				return
			}
			if errors.Is(serviceError, domain.ErrIconNotFound) {
				logger.Info().Msgf("Icon %s not found to add/remove tag %s to/from: %v", iconName, tag, serviceError)
				c.AbortWithStatus(404)
				return
			}
			logger.Error().Msgf("Failed to add/remove tag %s to/from %s: %v", tag, iconName, serviceError)
			c.AbortWithStatus(500)
			return
		}
		c.Status(201)
	}
}

func removeTagHandler(
	getUserInfo func(c *gin.Context) authr.UserInfo,
	removeTag func(iconName string, tag string, modifiedBy authr.UserInfo) error,
	logger zerolog.Logger,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		userInfo := getUserInfo(c)
		iconName := c.Param("name")
		tag := c.Param("tag")
		serviceError := removeTag(iconName, tag, userInfo)
		if serviceError != nil {
			if errors.Is(serviceError, authr.ErrPermission) {
				logger.Info().Msgf("Icon %s not found to add/remove tag %s to/from: %v", iconName, tag, serviceError)
				c.AbortWithStatus(403)
				return
			}
			if errors.Is(serviceError, domain.ErrIconNotFound) {
				logger.Info().Msgf("Icon %s not found to add/remove tag %s to/from: %v", iconName, tag, serviceError)
				c.AbortWithStatus(404)
				return
			}
			logger.Error().Msgf("Failed to add/remove tag %s to/from %s: %v", tag, iconName, serviceError)
			c.AbortWithStatus(500)
			return
		}
		c.Status(204)
	}
}
