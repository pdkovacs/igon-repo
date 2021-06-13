package web

import (
	"encoding/json"
	"io"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authn"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authr"
	log "github.com/sirupsen/logrus"
)

type BackdoorAuthorization struct {
	Username    string
	Permissions []authr.PermissionID
}

func HandlePutIntoBackdoorRequest(c *gin.Context) {
	logger := log.WithField("prefix", "PUT /backdoor/authentication")

	requestBody, errReadRequest := io.ReadAll(c.Request.Body)
	if errReadRequest != nil {
		logger.Errorf("failed to read request body %T: %v", c.Request.Body, errReadRequest)
		c.JSON(500, nil)
	}
	var requestedAuthorization BackdoorAuthorization
	errBodyUnmarshal := json.Unmarshal(requestBody, &requestedAuthorization)
	if errBodyUnmarshal != nil {
		logger.Errorf("failed to unmarshal request body %T: %v", requestBody, errBodyUnmarshal)
		c.JSON(400, nil)
	}
	session := sessions.Default(c)
	user := session.Get(UserKey)
	logger.Infof("%v requested authorization: %v", user, requestedAuthorization)

	sessionData, ok := user.(SessionData)
	if !ok {
		logger.Errorf("failed to cast %T to SessionData: ", user)
		c.AbortWithStatus(500)
		return
	}

	updatedCachedUserInfo := sessionData
	updatedCachedUserInfo.UserInfo.UserId = authn.LocalDomain.CreateUserID(requestedAuthorization.Username)
	updatedCachedUserInfo.UserInfo.Permissions = requestedAuthorization.Permissions
	session.Set(UserKey, SessionData{updatedCachedUserInfo.UserInfo})
	session.Save()
	c.JSON(200, nil)
}

func HandleGetIntoBackdoorRequest(c *gin.Context) {
	logger := log.WithField("prefix", "GET /backdoor/authentication")

	session := sessions.Default(c)
	user := session.Get(UserKey)
	sessionData, ok := user.(SessionData)
	if !ok {
		logger.Errorf("failed to cast %T to SessionData: ", user)
		c.AbortWithStatus(500)
		return
	}
	c.JSON(200, sessionData.UserInfo)
}
