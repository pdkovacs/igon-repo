package security

import (
	"encoding/json"
	"io"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Authorization struct {
	Username   string
	Privileges []string
}

func HandlePutIntoBackdoorRequest(c *gin.Context) {
	logger := log.WithField("prefix", "/backdoor/authentication")

	requestBody, errReadRequest := io.ReadAll(c.Request.Body)
	if errReadRequest != nil {
		logger.Errorf("failed to read request body %T: %v", c.Request.Body, errReadRequest)
		c.JSON(500, nil)
	}
	var requestedAuthorization Authorization
	errBodyUnmarshal := json.Unmarshal(requestBody, &requestedAuthorization)
	if errBodyUnmarshal != nil {
		logger.Errorf("failed to unmarshal request body %T: %v", requestBody, errBodyUnmarshal)
		c.JSON(400, nil)
	}
	session := sessions.Default(c)
	user := session.Get(userKey)
	logger.Infof("%v requested authorization: %v", user, requestedAuthorization)
	session.Set(userKey, UserSession{requestedAuthorization.Username, requestedAuthorization.Privileges})
	session.Save()
	c.JSON(200, nil)
}
