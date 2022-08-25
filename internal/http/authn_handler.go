package http

import (
	"fmt"

	"igo-repo/internal/app/security/authr"
	"igo-repo/internal/app/services"
	"igo-repo/internal/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

const UserKey = "igo-user"

type SessionData struct {
	UserInfo authr.UserInfo
}

func mustGetUserSession(c *gin.Context) SessionData {
	session := sessions.Default(c)
	user := session.Get(UserKey)
	if userSession, ok := user.(SessionData); ok {
		return userSession
	}
	panic(fmt.Errorf("unexpected user session type: %T", user))
}

func authenticationCheck(options config.Options, userService *services.UserService, log zerolog.Logger) gin.HandlerFunc {
	switch options.AuthenticationType {
	case config.BasicAuthentication:
		return checkBasicAuthentication(basicConfig{PasswordCredentialsList: options.PasswordCredentials}, *userService, log)
	case config.OIDCAuthentication:
		return checkOIDCAuthentication(log)
	}
	panic(fmt.Sprintf("unexpected authentication type: %v", options.AuthenticationType))
}

// authentication handles authentication
func authentication(options config.Options, userService *services.UserService, log zerolog.Logger) gin.HandlerFunc {
	switch options.AuthenticationType {
	case config.BasicAuthentication:
		return basicScheme(basicConfig{PasswordCredentialsList: options.PasswordCredentials}, userService)
	case config.OIDCAuthentication:
		return CreateOIDCSChemeHandler(oidcConfig{
			tokenIssuer:           options.OIDCTokenIssuer,
			clientRedirectBackURL: options.OIDCClientRedirectBackURL,
			clientID:              options.OIDCClientID,
			clientSecret:          options.OIDCClientSecret,
			serverURLContext:      options.ServerURLContext,
		}, userService, log)
	}
	return nil
}
