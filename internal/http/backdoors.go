package http

import (
	"encoding/json"
	"io"

	"igo-repo/internal/app/security/authr"
	"igo-repo/internal/logging"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func HandlePutIntoBackdoorRequest(log zerolog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		logger := logging.CreateMethodLogger(log, "PUT /backdoor/authentication")

		requestBody, errReadRequest := io.ReadAll(c.Request.Body)
		if errReadRequest != nil {
			logger.Error().Msgf("failed to read request body %T: %v", c.Request.Body, errReadRequest)
			c.JSON(500, nil)
		}
		permissions := []authr.PermissionID{}
		errBodyUnmarshal := json.Unmarshal(requestBody, &permissions)
		if errBodyUnmarshal != nil {
			logger.Error().Msgf("failed to unmarshal request body %T: %v", requestBody, errBodyUnmarshal)
			c.JSON(400, nil)
		}
		session := sessions.Default(c)
		user := session.Get(UserKey)
		logger.Info().Msgf("%v requested authorization: %v", user, permissions)

		sessionData, ok := user.(SessionData)
		if !ok {
			logger.Error().Msgf("failed to cast %T to SessionData: ", user)
			c.AbortWithStatus(500)
			return
		}

		updatedCachedUserInfo := sessionData
		updatedCachedUserInfo.UserInfo.Permissions = permissions
		session.Set(UserKey, SessionData{updatedCachedUserInfo.UserInfo})
		session.Save()
		c.JSON(200, nil)
	}
}

func HandleGetIntoBackdoorRequest(log zerolog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		logger := logging.CreateMethodLogger(log, "GET /backdoor/authentication")

		session := sessions.Default(c)
		user := session.Get(UserKey)
		sessionData, ok := user.(SessionData)
		if !ok {
			logger.Error().Msgf("failed to cast %T to SessionData: ", user)
			c.AbortWithStatus(500)
			return
		}
		c.JSON(200, sessionData.UserInfo)
	}
}
