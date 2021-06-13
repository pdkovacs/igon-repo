package web

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authn"
	"github.com/pdkovacs/igo-repo/backend/pkg/services"
	log "github.com/sirupsen/logrus"
)

func UserInfoHandler(userService services.UserService) func(c *gin.Context) {
	return func(c *gin.Context) {
		domainId := c.Query("domainId")
		userId := c.Query("userId")
		logger := log.WithField("prefix", "UserInfoHandler")
		session := sessions.Default(c)
		user := session.Get(UserKey)

		usession, ok := user.(SessionData)
		if !ok {
			logger.Errorf("failed to cast user session of type %T", user)
		}

		if domainId == "" {
			domainId = usession.UserInfo.UserId.DomainID
		} else if domainId != usession.UserInfo.UserId.DomainID {
			c.JSON(400, "Cross domain user info queries not yet supported")
			return
		}

		userInfo := userService.GetUserInfo(authn.UserID{IDInDomain: userId, DomainID: domainId})
		logger.Infof("User info found: %v", usession)
		c.JSON(200, userInfo)
	}
}
