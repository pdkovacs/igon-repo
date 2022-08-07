package http

import (
	"fmt"

	"igo-repo/internal/app/security/authr"
	"igo-repo/internal/app/services"
	"igo-repo/internal/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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

func AuthenticationCheck(options config.Options, userService *services.UserService) gin.HandlerFunc {
	logger := log.WithField("prefix", "AuthenticationCheck")
	logger.Debugf("AuthenticationType: %v", options.AuthenticationType)
	switch options.AuthenticationType {
	case config.BasicAuthentication:
		return checkBasicAuthentication(BasicConfig{PasswordCredentialsList: options.PasswordCredentials}, *userService)
	case config.OIDCAuthentication:
		return checkOIDCAuthentication
	}
	panic(fmt.Sprintf("unexpected authentication type: %v", options.AuthenticationType))
}

// Authentication handles authentication
func Authentication(options config.Options, userService *services.UserService) gin.HandlerFunc {
	logger := log.WithField("prefix", "Authentication")
	logger.Debugf("AuthenticationType: %v", options.AuthenticationType)
	switch options.AuthenticationType {
	case config.BasicAuthentication:
		return basicScheme(BasicConfig{PasswordCredentialsList: options.PasswordCredentials}, userService)
	case config.OIDCAuthentication:
		return oidcScheme(OIDCConfig{
			TokenIssuer:             options.OIDCTokenIssuer,
			UserAuthorizationURL:    options.OIDCUserAuthorizationURL,
			ClientRedirectBackURL:   options.OIDCClientRedirectBackURL,
			AccessTokenURL:          options.OIDCAccessTokenURL,
			IpJwtPublicKeyURL:       options.OIDCIpJwtPublicKeyURL,
			IpJwtPublicKeyPemBase64: options.OIDCIpJwtPublicKeyPemBase64,
			IpLogoutURL:             options.OIDCIpLogoutURL,
			ClientID:                options.OIDCClientID,
			ClientSecret:            options.OIDCClientSecret,
			ServerURLContext:        options.ServerURLContext,
		}, userService)
	}
	return nil
}
