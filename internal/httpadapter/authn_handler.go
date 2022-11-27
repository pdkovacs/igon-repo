package httpadapter

import (
	"fmt"

	"igo-repo/internal/app/security/authn"
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

func mustGetUserSession_(c *gin.Context) SessionData {
	session := sessions.Default(c)
	user := session.Get(UserKey)
	if userSession, ok := user.(SessionData); ok {
		return userSession
	}
	panic(fmt.Errorf("unexpected user session type: %T", user))
}

func getUserInfo(authnType authn.AuthenticationScheme) func(c *gin.Context) (authr.UserInfo, error) {
	return func(c *gin.Context) (authr.UserInfo, error) {
		switch authnType {
		case authn.SchemeBasic:
			{
				sessionData := mustGetUserSession_(c)
				return sessionData.UserInfo, nil
			}
		case authn.SchemeOIDC:
			{
				sessionData := mustGetUserSession_(c)
				return sessionData.UserInfo, nil
			}
		case authn.SchemeOIDCProxy:
			{
				uInfoCandidate, exists := c.Get(forwardedUserInfoKey)
				if !exists {
					return authr.UserInfo{}, fmt.Errorf("user-info key missing from request context")
				}
				if userInfo, ok := uInfoCandidate.(authr.UserInfo); ok {
					return userInfo, nil
				}
				panic(fmt.Sprintf("failed to cast to UserInfo: %#v", uInfoCandidate))
			}
		}
		panic(fmt.Sprintf("unexpected authentication type: %v", authnType))
	}
}

func authenticationCheck(options config.Options, userService *services.UserService, log zerolog.Logger) gin.HandlerFunc {
	switch options.AuthenticationType {
	case authn.SchemeBasic:
		return checkBasicAuthentication(basicConfig{PasswordCredentialsList: options.PasswordCredentials}, *userService, log)
	case authn.SchemeOIDC:
		return checkOIDCAuthentication(log)
	case authn.SchemeOIDCProxy:
		return checkOIDCProxyAuthentication(userService.AuthorizationService, log)
	}
	panic(fmt.Sprintf("unexpected authentication type: %v", options.AuthenticationType))
}

// authentication handles authentication
func authentication(options config.Options, userService *services.UserService, log zerolog.Logger) gin.HandlerFunc {
	switch options.AuthenticationType {
	case authn.SchemeBasic:
		return basicScheme(basicConfig{PasswordCredentialsList: options.PasswordCredentials}, userService)
	case authn.SchemeOIDC:
		return CreateOIDCSChemeHandler(oidcConfig{
			tokenIssuer:           options.OIDCTokenIssuer,
			clientRedirectBackURL: options.OIDCClientRedirectBackURL,
			clientID:              options.OIDCClientID,
			clientSecret:          options.OIDCClientSecret,
			serverURLContext:      options.ServerURLContext,
		}, userService, log)
	case authn.SchemeOIDCProxy:
		return func(g *gin.Context) {
		}
	}
	return nil
}