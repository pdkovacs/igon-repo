package http

import (
	"igo-repo/internal/app/security/authn"
	"igo-repo/internal/app/security/authr"
	"igo-repo/internal/app/services"
	"igo-repo/internal/logging"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type userInfoDTO struct {
	Username    string               `json:"username"`
	Groups      []authr.GroupID      `json:"groups"`
	Permissions []authr.PermissionID `json:"permissions"`
	DisplayName string               `json:"displayName"`
}

func userInfoHandler(userService services.UserService, log zerolog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		userId := c.Query("userId")
		logger := logging.CreateMethodLogger(log, "UserInfoHandler")
		session := sessions.Default(c)
		user := session.Get(UserKey)

		usession, ok := user.(SessionData)
		if !ok {
			logger.Error().Msgf("failed to cast user session of type %T", user)
		}

		var userInfo authr.UserInfo
		if userId == "" {
			userInfo = usession.UserInfo
		} else {
			userInfo = userService.GetUserInfo(authn.UserID{IDInDomain: userId})
		}
		logger.Debug().Msgf("User info: %v", userInfo)

		responseUserInfo := userInfoDTO{
			Username:    userInfo.UserId.IDInDomain,
			Groups:      userInfo.Groups,
			Permissions: userInfo.Permissions,
			DisplayName: userInfo.DisplayName,
		}

		c.JSON(200, responseUserInfo)
	}
}
