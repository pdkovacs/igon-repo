package httpadapter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"iconrepo/internal/app/domain"
	"iconrepo/internal/app/security/authn"
	"iconrepo/internal/app/security/authr"
	"iconrepo/internal/app/services"

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

func describeAllIcons(describeAllIcons func() ([]domain.IconDescriptor, error)) func(g *gin.Context) {
	return func(g *gin.Context) {
		logger := zerolog.Ctx(g.Request.Context())

		icons, err := describeAllIcons()
		if err != nil {
			logger.Error().Err(err).Send()
			g.AbortWithStatus(500)
		}
		responseIcon := []IconDTO{}
		for _, icon := range icons {
			responseIcon = append(responseIcon, CreateResponseIcon(iconRootPath, icon))
		}
		g.JSON(200, responseIcon)
	}
}

func describeIcon(describeIcon func(iconName string) (domain.IconDescriptor, error)) func(g *gin.Context) {
	return func(g *gin.Context) {
		logger := zerolog.Ctx(g.Request.Context())

		iconName := g.Param("name")
		icon, err := describeIcon(iconName)
		if err != nil {
			logger.Error().Err(err).Send()
			if errors.Is(err, domain.ErrIconNotFound) {
				g.AbortWithStatus(404)
				return
			}
			g.AbortWithStatus(500)
			return
		}
		responseIcon := CreateResponseIcon(iconRootPath, icon)
		g.JSON(200, responseIcon)
	}
}

func createIcon(
	getUserInfo func(c *gin.Context) authr.UserInfo,
	createIcon func(iconName string, initialIconfileContent []byte, modifiedBy authr.UserInfo) (domain.Icon, error),
	publish func(msg services.NotificationMessage, initiator authn.UserID),
) func(g *gin.Context) {
	return func(g *gin.Context) {
		logger := zerolog.Ctx(g.Request.Context())

		r := g.Request
		r.ParseMultipartForm(32 << 20) // limit your max input length to 32MB

		iconName := r.FormValue("iconName")
		if len(iconName) == 0 {
			logger.Info().Msg("invalid icon name: <empty-string>")
			g.AbortWithStatus(400)
			return
		}

		var buf bytes.Buffer

		file, _, err := r.FormFile("iconfile")
		if err != nil {
			logger.Info().Str("icon-name", iconName).Err(err).Msg("failed to retrieve iconfile for icon")
			g.AbortWithStatus(400)
			return
		}
		defer file.Close()

		io.Copy(&buf, file)
		defer buf.Reset()
		logger.Info().Str("icon-name", iconName).Int("byte-count", buf.Len()).Msg("received icon")

		authorInfo := getUserInfo(g)
		icon, errCreate := createIcon(iconName, buf.Bytes(), authorInfo)
		if errCreate != nil {
			logger.Error().Str("icon-name", iconName).Err(errCreate).Msg("failed to create icon")
			if errors.Is(errCreate, authr.ErrPermission) {
				g.AbortWithStatus(403)
				return
			} else if errors.Is(errCreate, domain.ErrIconAlreadyExists) {
				g.AbortWithStatus(409)
				return
			} else {
				g.AbortWithStatus(500)
				return
			}
		}
		g.JSON(201, iconToResponseIcon(icon))
		publish(services.NotifMsgIconCreated, authorInfo.UserId)
	}
}

func getIconfile(getIconfile func(iconName string, iconfile domain.IconfileDescriptor) ([]byte, error)) func(g *gin.Context) {
	return func(g *gin.Context) {
		logger := zerolog.Ctx(g.Request.Context())

		iconName := g.Param("name")
		format := g.Param("format")
		size := g.Param("size")
		iconfile, err := getIconfile(iconName, domain.IconfileDescriptor{
			Format: format,
			Size:   size,
		})
		if err != nil {
			if errors.Is(err, domain.ErrIconfileNotFound) {
				g.AbortWithStatus(404)
				return
			}
			logger.Error().Err(err).Str("icon-name", iconName).Str("format", format).Str("size", size).Msg("failed to retrieve icon content")
			g.AbortWithStatus(500)
		}
		if format == "svg" {
			g.Data(200, "image/svg+xml", iconfile)
			return
		}
		g.Data(200, "application/octet-stream", iconfile)
	}
}

func addIconfile(
	getUserInfo func(g *gin.Context) authr.UserInfo,
	addIconfile func(iconName string, initialIconfileContent []byte, modifiedBy authr.UserInfo) (domain.IconfileDescriptor, error),
	publish func(msg services.NotificationMessage, initiator authn.UserID),
) func(g *gin.Context) {

	return func(g *gin.Context) {
		logger := zerolog.Ctx(g.Request.Context())

		authorInfo := getUserInfo(g)

		authrErr := authr.HasRequiredPermissions(
			authorInfo,
			[]authr.PermissionID{
				authr.UPDATE_ICON,
				authr.ADD_ICONFILE,
			})
		if authrErr != nil {
			g.AbortWithStatus(403)
			return
		}

		r := g.Request
		r.ParseMultipartForm(32 << 20) // limit your max input length to 32MB

		iconName := r.FormValue("iconName")
		if len(iconName) == 0 {
			logger.Info().Msg("invalid icon name: <empty-string>")
			g.AbortWithStatus(400)
			return
		}

		var buf bytes.Buffer

		file, _, err := r.FormFile("iconfile")
		if err != nil {
			logger.Info().Err(err).Str("icon-name", iconName).Msg("failed to retrieve iconfile")
			g.AbortWithStatus(400)
			return
		}
		defer file.Close()

		io.Copy(&buf, file)
		logger.Info().Str("icon-name", iconName).Msg("received iconfile content")

		iconfileDescriptor, errAdd := addIconfile(iconName, buf.Bytes(), authorInfo)
		if errAdd != nil {
			logger.Error().Err(errAdd).Str("icon-name", iconName).Msg("failed to add iconfile")
			if errors.Is(errAdd, authr.ErrPermission) {
				g.AbortWithStatus(403)
				return
			} else if errors.Is(errAdd, domain.ErrIconfileAlreadyExists) {
				g.AbortWithStatus(409)
				return
			} else {
				g.AbortWithStatus(500)
			}
		}
		g.JSON(200, CreateIconPath(iconRootPath, iconName, iconfileDescriptor))
		publish(services.NotifMsgIconfileAdded, authorInfo.UserId)
		buf.Reset()
	}
}

func deleteIcon(
	getUserInfo func(g *gin.Context) authr.UserInfo,
	deleteIcon func(iconName string, modifiedBy authr.UserInfo) error,
	publish func(msg services.NotificationMessage, initiator authn.UserID),
) func(g *gin.Context) {
	return func(g *gin.Context) {
		logger := zerolog.Ctx(g.Request.Context())

		authorInfo := getUserInfo(g)
		iconName := g.Param("name")
		deleteError := deleteIcon(iconName, authorInfo)
		if deleteError != nil {
			if errors.Is(deleteError, authr.ErrPermission) {
				g.AbortWithStatus(403)
				return
			}
			logger.Error().Err(deleteError).Str("icon-name", iconName).Msg("failed to delete icon")
			g.AbortWithStatus(500)
			return
		}
		g.Status(204)
		publish(services.NotifMsgIconDeleted, authorInfo.UserId)
	}
}

func deleteIconfile(
	getUserInfo func(c *gin.Context) authr.UserInfo,
	deleteIconfile func(iconName string, iconfile domain.IconfileDescriptor, modifiedBy authr.UserInfo) error,
	publish func(msg services.NotificationMessage, initiator authn.UserID),
) func(g *gin.Context) {
	return func(g *gin.Context) {
		logger := zerolog.Ctx(g.Request.Context())

		authorInfo := getUserInfo(g)
		iconName := g.Param("name")
		format := g.Param("format")
		size := g.Param("size")
		iconfileDescriptor := domain.IconfileDescriptor{Format: format, Size: size}
		deleteError := deleteIconfile(iconName, iconfileDescriptor, authorInfo)
		if deleteError != nil {
			if errors.Is(deleteError, authr.ErrPermission) {
				g.AbortWithStatus(403)
				return
			}
			if errors.Is(deleteError, domain.ErrIconNotFound) {
				logger.Info().Str("icon-name", iconName).Str("format", format).Str("size", size).Msg("Icon not found")
				g.AbortWithStatus(404)
				return
			}
			if errors.Is(deleteError, domain.ErrIconfileNotFound) {
				logger.Info().Str("icon-name", iconName).Str("format", format).Str("size", size).Msg("iconfile not found")
				g.AbortWithStatus(404)
				return
			}
			logger.Error().Str("icon-name", iconName).Str("format", format).Str("size", size).Msg("failed to delete iconfile")
			g.AbortWithStatus(500)
			return
		}
		g.Status(204)
		publish(services.NotifMsgIconfileDeleted, authorInfo.UserId)
	}
}

func getTags(getTags func() ([]string, error)) func(g *gin.Context) {
	return func(g *gin.Context) {
		logger := zerolog.Ctx(g.Request.Context())

		tags, serviceError := getTags()
		if serviceError != nil {
			logger.Error().Err(serviceError).Msg("failed to retrieve tags")
			g.AbortWithStatus(500)
			return
		}
		g.JSON(200, tags)
	}
}

type AddServiceRequestData struct {
	Tag string `json:"tag"`
}

func addTag(
	getUserInfo func(c *gin.Context) authr.UserInfo,
	addTag func(iconName string, tag string, modifiedBy authr.UserInfo) error,
) func(g *gin.Context) {
	return func(g *gin.Context) {
		logger := zerolog.Ctx(g.Request.Context())

		userInfo := getUserInfo(g)
		iconName := g.Param("name")

		jsonData, readBodyErr := io.ReadAll(g.Request.Body)
		if readBodyErr != nil {
			logger.Error().Err(readBodyErr).Msg("failed to read body")
			g.AbortWithStatus(400)
			return
		}
		tagRequestData := AddServiceRequestData{}
		json.Unmarshal(jsonData, &tagRequestData)
		tag := tagRequestData.Tag
		if len(tag) == 0 {
			logger.Error().Msg("Tags must be at last one character long")
			g.AbortWithStatus(400)
			return
		}

		serviceError := addTag(iconName, tag, userInfo)
		if serviceError != nil {
			if errors.Is(serviceError, authr.ErrPermission) {
				logger.Info().Err(serviceError).Str("icon-name", iconName).Str("tag", tag).Msg("icon not found to add/remove tag")
				g.AbortWithStatus(403)
				return
			}
			if errors.Is(serviceError, domain.ErrIconNotFound) {
				logger.Info().Err(serviceError).Str("icon-name", iconName).Str("tag", tag).Msg("icon not found to add/remove tag")
				g.AbortWithStatus(404)
				return
			}
			logger.Error().Err(serviceError).Str("icon-name", iconName).Str("tag", tag).Msg("failed to add/remove tag")
			g.AbortWithStatus(500)
			return
		}
		g.Status(201)
	}
}

func removeTag(
	getUserInfo func(c *gin.Context) authr.UserInfo,
	removeTag func(iconName string, tag string, modifiedBy authr.UserInfo) error,
) func(g *gin.Context) {
	return func(g *gin.Context) {
		logger := zerolog.Ctx(g.Request.Context())

		userInfo := getUserInfo(g)
		iconName := g.Param("name")
		tag := g.Param("tag")
		serviceError := removeTag(iconName, tag, userInfo)
		if serviceError != nil {
			if errors.Is(serviceError, authr.ErrPermission) {
				logger.Info().Err(serviceError).Str("icon-name", iconName).Str("tag", tag).Msg("icon not found to add/remove tag")
				g.AbortWithStatus(403)
				return
			}
			if errors.Is(serviceError, domain.ErrIconNotFound) {
				logger.Info().Err(serviceError).Str("icon-name", iconName).Str("tag", tag).Msg("icon not found to add/remove tag")
				g.AbortWithStatus(404)
				return
			}
			logger.Error().Err(serviceError).Str("icon-name", iconName).Str("tag", tag).Msg("failed to add/remove tag")
			g.AbortWithStatus(500)
			return
		}
		g.Status(204)
	}
}
