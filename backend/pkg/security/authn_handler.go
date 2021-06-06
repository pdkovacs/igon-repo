package security

import (
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// OIDCConfig holds the configuration for the OIDCConfig authentication scheme
type OIDCConfig struct{}

const userKey = "igo-user"

type UserSession struct {
	Username    string
	Permissions []string
}

func oidcScheme(c *gin.Context) {
	session := sessions.Default(c)
	fmt.Printf("OIDC authentication: session: %v\n", session)
	c.AbortWithStatus(500)
}

// HandlerProvider handles authentication
func HandlerProvider(authnConfig interface{}) gin.HandlerFunc {
	switch authnConfig.(type) {
	case BasicConfig:
		return basicScheme(authnConfig.(BasicConfig))
	case OIDCConfig:
		return oidcScheme
	}
	return nil
}

func UserInfoHandler(c *gin.Context) {
	logger := log.WithField("prefix", "UserInfoHandler")
	session := sessions.Default(c)
	user := session.Get(userKey)

	usession, ok := user.(UserSession)
	if !ok {
		logger.Errorf("failed to cast user session of type %T", user)
	}
	logger.Infof("User info found: %v", usession)
	c.JSON(200, Authorization{Username: usession.Username, Privileges: usession.Permissions})
}
