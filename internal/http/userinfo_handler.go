package http

import (
	"igo-repo/internal/app/security/authn"
	"igo-repo/internal/app/security/authr"
	"igo-repo/internal/app/services"

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

func userInfoHandler(authType authn.AuthenticationScheme, userService services.UserService, logger zerolog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		userId := c.Query("userId")

		if authType == authn.SchemeOIDCProxy {
			userInfo, getUserInfoErr := getUserInfo(authType)(c)
			if getUserInfoErr != nil {
				logger.Error().Msgf("failed to find user-info")
				c.AbortWithStatus(401)
				return
			}

			responseUserInfo := userInfoDTO{
				Username:    userInfo.UserId.IDInDomain,
				Groups:      userInfo.Groups,
				Permissions: userInfo.Permissions,
				DisplayName: userInfo.DisplayName,
			}

			c.JSON(200, responseUserInfo)
			return
		}

		session := sessions.Default(c)
		user := session.Get(UserKey)

		usession, ok := user.(SessionData)
		if !ok {
			logger.Error().Msgf("failed to cast user session of type %T", user)
			c.AbortWithStatus(500)
			return
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
