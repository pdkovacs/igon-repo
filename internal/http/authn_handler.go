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

func MustGetUserSession(c *gin.Context) SessionData {
	session := sessions.Default(c)
	user := session.Get(UserKey)
	if userSession, ok := user.(SessionData); ok {
		return userSession
	}
	panic(fmt.Errorf("unexpected user session type: %T", user))
}

func AuthenticationCheck(options config.Options, userService *services.UserService, log zerolog.Logger) gin.HandlerFunc {
	switch options.AuthenticationType {
	case config.BasicAuthentication:
		return checkBasicAuthentication(BasicConfig{PasswordCredentialsList: options.PasswordCredentials}, *userService, log)
	case config.OIDCAuthentication:
		return checkOIDCAuthentication(log)
	}
	panic(fmt.Sprintf("unexpected authentication type: %v", options.AuthenticationType))
}

// Authentication handles authentication
func Authentication(options config.Options, userService *services.UserService, log zerolog.Logger) gin.HandlerFunc {
	switch options.AuthenticationType {
	case config.BasicAuthentication:
		return basicScheme(BasicConfig{PasswordCredentialsList: options.PasswordCredentials}, userService)
	case config.OIDCAuthentication:
		return CreateOIDCSCheme(OIDCConfig{
			TokenIssuer:           options.OIDCTokenIssuer,
			ClientRedirectBackURL: options.OIDCClientRedirectBackURL,
			ClientID:              options.OIDCClientID,
			ClientSecret:          options.OIDCClientSecret,
			ServerURLContext:      options.ServerURLContext,
		}, userService, log)
	}
	return nil
}
