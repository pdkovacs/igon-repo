package authn

import (
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Scheme is the authentication scheme
type Scheme int

const (
	// Basic authentication
	Basic Scheme = iota
	// OIDC provides OpenID Connect authentication
	OIDC
)

func basicScheme(c *gin.Context) {
	session := sessions.Default(c)
	fmt.Printf("Basic authentication: session: %v\n", session)
}

func oidcScheme(c *gin.Context) {
	session := sessions.Default(c)
	fmt.Printf("OIDC authentication: session: %v\n", session)
	c.AbortWithStatus(500)
}

// HandlerProvider handles authentication
func HandlerProvider(authnScheme Scheme) gin.HandlerFunc {
	switch authnScheme {
	case Basic:
		return basicScheme
	case OIDC:
		return oidcScheme
	}
	return nil
}
