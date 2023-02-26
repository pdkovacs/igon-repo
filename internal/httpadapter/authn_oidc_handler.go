package httpadapter

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"igo-repo/internal/app/security/authn"
	"igo-repo/internal/app/security/authr"
	"igo-repo/internal/app/services"
	"igo-repo/internal/logging"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

type HandleOAuth2Callback func(c *gin.Context, storedState string) (*claims, error)

func checkOIDCAuthentication(log zerolog.Logger) func(c *gin.Context) {
	logger := logging.CreateMethodLogger(log, "checkOIDCAuthentication")

	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(UserKey)
		if user == nil {
			logger.Debug().Msgf("Request not authenticated: %v", c.Request.URL)
			c.AbortWithStatus(401)
			return
		}
		logger.Debug().Msgf("User session: %v", user)
	}
}

type oidcConfig struct {
	clientID              string
	clientSecret          string
	clientRedirectBackURL string
	tokenIssuer           string
	serverURLContext      string
}

const oidcTokenRequestStateKey = "oidcTokenRequestState"

type claims struct {
	Email    string   `json:"email"`
	Verified bool     `json:"email_verified"`
	Groups   []string `json:"groups"`
	Name     string   `json:"name"`
}

type oidcScheme struct {
	config         oidcConfig
	logger         zerolog.Logger
	userService    *services.UserService
	usernameCookie string
}

func CreateOIDCSChemeHandler(config oidcConfig, userService *services.UserService, usernameCookie string, clientServerURL string, logger zerolog.Logger) gin.HandlerFunc {
	scheme := oidcScheme{
		config:         config,
		logger:         logger,
		userService:    userService,
		usernameCookie: usernameCookie,
	}
	return scheme.createHandler(clientServerURL)
}

func (scheme *oidcScheme) createHandler(clientServerURL string) gin.HandlerFunc {
	logger := logging.CreateMethodLogger(scheme.logger, "oidc-authn")
	config := scheme.config

	provider, err := oidc.NewProvider(context.TODO(), config.tokenIssuer)
	if err != nil {
		panic(err)
	}

	var verifier = provider.Verifier(&oidc.Config{ClientID: config.clientID})

	oauth2Config := oauth2.Config{
		ClientID:     config.clientID,
		ClientSecret: config.clientSecret,
		RedirectURL:  config.clientRedirectBackURL,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
	}

	handleOAuth2Callback := scheme.getOAuth2CallbackHandler(oauth2Config, verifier, scheme.usernameCookie)

	logger.Debug().Msg("Returning oidc-authn handler")

	return func(c *gin.Context) {
		logging.CreateMethodLogger(logger, "oidc-authn")
		logger.Debug().Msgf("Incoming request %v...", c.Request.URL)

		queryError := c.Query("error")
		if queryError != "" {
			logger.Error().Msgf("callback error: %s", queryError)
			c.Writer.WriteString("callback error")
			c.AbortWithStatus(401)
			return
		}

		session := sessions.Default(c)
		user := session.Get(UserKey)
		if userSession, ok := user.(SessionData); ok {
			if len(userSession.UserInfo.UserId.IDInDomain) > 0 {
				logger.Debug().Msg("session already authenticated")
				// TODO: redirect to proper path
				c.Redirect(302, clientServerURL)
				return
			}
			logger.Error().Msg("has user-session, but no user-id")
			c.AbortWithStatus(401)
			return
		}

		authrCode := c.Query("code")
		if authrCode != "" {
			logger.Debug().Msgf("incoming user approval with autherization code %v", authrCode)
			state := session.Get(oidcTokenRequestStateKey)
			storedState, ok := state.(string)
			if !ok {
				logger.Error().Msg("no suitable auth2-state stored for session")
				c.AbortWithStatus(401)
				return
			}
			claims, handleCallbackErr := handleOAuth2Callback(c, storedState)
			if handleCallbackErr == nil && claims != nil {
				logger.Info().Msgf("claims collected: %+v", claims)
				// FIXME: Use other than local-domain
				userId := authn.LocalDomain.CreateUserID(claims.Email)
				if claims.Groups != nil {
					scheme.userService.UpdateUserInfo(userId, authr.GroupNamesToGroupIDs(claims.Groups))
				}
				userInfo := scheme.userService.GetUserInfo(userId)
				session.Set(UserKey, SessionData{userInfo})
				session.Save()
				c.Abort()
				c.Redirect(http.StatusFound, fmt.Sprintf("%s/", config.serverURLContext))
				return
			}
			if handleCallbackErr != nil {
				logger.Error().Msgf("error while processing authorization code: %v", handleCallbackErr)
				c.AbortWithStatus(401)
				return
			}
			logger.Error().Msg("No claims found")
			c.AbortWithStatus(401)
			return
		}

		state := randSeq(32)
		session.Set(oidcTokenRequestStateKey, state)
		session.Save()

		logger.Debug().Msgf("new authn round started, state %v saved to session", state)

		c.Abort()
		c.Redirect(http.StatusFound, oauth2Config.AuthCodeURL(state))
	}
}

func (scheme *oidcScheme) getOAuth2CallbackHandler(oauth2Config oauth2.Config, verifier *oidc.IDTokenVerifier, usernameCookie string) HandleOAuth2Callback {

	logger := logging.CreateMethodLogger(scheme.logger, "handleOAuth2Callback")

	return func(c *gin.Context, storedState string) (*claims, error) {
		r := c.Request

		responseState := r.URL.Query().Get("state")
		if responseState != storedState {
			return nil, fmt.Errorf("response state %v doesn't equal the stored state: %v", responseState, storedState)
		}

		oauth2Token, err := oauth2Config.Exchange(context.TODO(), r.URL.Query().Get("code"))
		if err != nil {
			logger.Error().Msgf("failed to obtain OAuth2 token: %v", err)
			return nil, fmt.Errorf("failed to obtain OAuth2 token: %w", err)
		}

		// Extract the ID Token from OAuth2 token.
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			logger.Error().Msgf("failed to extract ID token: %v", err)
			return nil, fmt.Errorf("failed to extract ID token: %w", err)
		}

		// Parse and verify ID Token payload.
		idToken, err := verifier.Verify(context.TODO(), rawIDToken)
		if err != nil {
			logger.Error().Msgf("failed to verify ID token: %v", err)
			return nil, fmt.Errorf("failed to verify ID token: %w", err)
		}

		// Extract custom claims
		var claims claims
		if err := idToken.Claims(&claims); err != nil {
			logger.Error().Msgf("failed to extract claims from ID token: %v", err)
			return nil, fmt.Errorf("failed to extract claims from ID token: %w", err)
		}

		if usernameCookie != "" {
			c.SetCookie(usernameCookie, claims.Email, 0, "/", "", false, false)
		}

		return &claims, nil
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
