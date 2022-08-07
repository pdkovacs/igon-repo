package http

import (
	"igo-repo/internal/app/security/authn"
	"igo-repo/internal/app/security/authr"
	"igo-repo/internal/app/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type UserInfoDTO struct {
	Username    string               `json:"username"`
	Groups      []authr.GroupID      `json:"groups"`
	Permissions []authr.PermissionID `json:"permissions"`
	DisplayName string               `json:"displayName"`
}

func UserInfoHandler(userService services.UserService) func(c *gin.Context) {
	return func(c *gin.Context) {
		userId := c.Query("userId")
		logger := log.WithField("prefix", "UserInfoHandler")
		session := sessions.Default(c)
		user := session.Get(UserKey)

		usession, ok := user.(SessionData)
		if !ok {
			logger.Errorf("failed to cast user session of type %T", user)
		}

		var userInfo authr.UserInfo
		if userId == "" {
			userInfo = usession.UserInfo
		} else {
			userInfo = userService.GetUserInfo(authn.UserID{IDInDomain: userId})
		}
		logger.Debugf("User info: %v", userInfo)

		responseUserInfo := UserInfoDTO{
			Username:    userInfo.UserId.IDInDomain,
			Groups:      userInfo.Groups,
			Permissions: userInfo.Permissions,
			DisplayName: userInfo.DisplayName,
		}

		c.JSON(200, responseUserInfo)
	}
}
