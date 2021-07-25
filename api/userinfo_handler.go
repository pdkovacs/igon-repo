package api

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pdkovacs/igo-repo/app/security/authn"
	"github.com/pdkovacs/igo-repo/app/services"
	log "github.com/sirupsen/logrus"
)

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

		var userInfo services.UserInfo
		if userId == "" {
			userInfo = usession.UserInfo
		} else {
			userInfo = userService.GetUserInfo(authn.UserID{IDInDomain: userId})
		}

		logger.Debugf("User info: %v", userInfo)
		c.JSON(200, userInfo)
	}
}
