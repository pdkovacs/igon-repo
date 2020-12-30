package authn

import (
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// OIDCConfig holds the configuration for the OIDCConfig authentication scheme
type OIDCConfig struct{}

const userKey = "igo-user"

type userSession struct {
	username    string
	permissions []string
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
