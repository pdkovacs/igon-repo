package web

import (
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pdkovacs/igo-repo/backend/pkg/services"
)

// OIDCConfig holds the configuration for the OIDCConfig authentication scheme
type OIDCConfig struct{}

const UserKey = "igo-user"

type SessionData struct {
	UserInfo services.UserInfo
}

func MustGetUserSession(c *gin.Context) SessionData {
	session := sessions.Default(c)
	user := session.Get(UserKey)
	if userSession, ok := user.(SessionData); ok {
		return userSession
	}
	panic(fmt.Errorf("unexpected user session type: %T", user))
}

func oidcScheme(c *gin.Context) {
	session := sessions.Default(c)
	fmt.Printf("OIDC authentication: session: %v\n", session)
	c.AbortWithStatus(500)
}

// HandlerProvider handles authentication
func HandlerProvider(authnConfig interface{}, userService *services.UserService) gin.HandlerFunc {
	switch authnConfig := authnConfig.(type) {
	case BasicConfig:
		return basicScheme(authnConfig, userService)
	case OIDCConfig:
		return oidcScheme
	}
	return nil
}
