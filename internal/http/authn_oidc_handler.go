package http

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"igo-repo/internal/app/security/authn"
	"igo-repo/internal/app/security/authr"
	"igo-repo/internal/app/services"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

func checkOIDCAuthentication(c *gin.Context) {
	logger := log.WithField("prefix", "checkOIDCAuthentication")
	session := sessions.Default(c)
	user := session.Get(UserKey)
	if user == nil {
		logger.Debugf("No user session redirecting to /login...")
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
	}
	logger.Debugf("User session: %v", user)
}

type OIDCConfig struct {
	ClientID              string
	ClientSecret          string
	ClientRedirectBackURL string
	TokenIssuer           string
	ServerURLContext      string
}

const oidcTokenRequestStateKey = "oidcTokenRequestState"

type claims struct {
	Email    string   `json:"email"`
	Verified bool     `json:"email_verified"`
	Groups   []string `json:"groups"`
}

func oidcScheme(config OIDCConfig, userService *services.UserService) gin.HandlerFunc {
	logger := log.WithField("prefix", "oidc-authn")

	provider, err := oidc.NewProvider(context.TODO(), config.TokenIssuer)
	if err != nil {
		panic(err)
	}

	var verifier = provider.Verifier(&oidc.Config{ClientID: config.ClientID})

	oauth2Config := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.ClientRedirectBackURL,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
	}

	handleOAuth2Callback := handleOAuth2Callback(oauth2Config, verifier)

	logger.Debug("Returning oidc-authn handler")

	return func(c *gin.Context) {
		logger := log.WithField("prefix", "oidc-authn")
		logger.Debugf("Incoming request %v...", c.Request.URL)

		queryError := c.Query("error")
		if queryError != "" {
			logger.Errorf("callback error: %s", queryError)
			c.Writer.WriteString("callback error")
			c.AbortWithStatus(401)
			return
		}

		session := sessions.Default(c)
		user := session.Get(UserKey)
		if userSession, ok := user.(SessionData); ok {
			if len(userSession.UserInfo.UserId.IDInDomain) > 0 {
				logger.Debugf("session already authenticated")
				return
			}
			logger.Errorf("has user-session, but no user-id")
			c.AbortWithStatus(401)
			return
		}

		authrCode := c.Query("code")
		if authrCode != "" {
			logger.Debugf("incoming user approval with autherization code %v", authrCode)
			state := session.Get(oidcTokenRequestStateKey)
			storedState, ok := state.(string)
			if !ok {
				logger.Error("no suitable auth2-state stored for session")
				c.AbortWithStatus(401)
				return
			}
			claims, handleCallbackErr := handleOAuth2Callback(c, storedState)
			if handleCallbackErr == nil && claims != nil {
				logger.Infof("claims collected: %+v", claims)
				// FIXME: Use other than local-domain
				userId := authn.LocalDomain.CreateUserID(claims.Email)
				if claims.Groups != nil {
					userService.UpdateUserInfo(userId, authr.GroupNamesToGroupIDs(claims.Groups))
				}
				userInfo := userService.GetUserInfo(userId)
				session.Set(UserKey, SessionData{userInfo})
				session.Save()
				c.Abort()
				c.Redirect(http.StatusFound, fmt.Sprintf("%s/", config.ServerURLContext))
				return
			}
			if handleCallbackErr != nil {
				logger.Errorf("error while processing authorization code: %v", handleCallbackErr)
				c.AbortWithStatus(401)
				return
			}
			logger.Errorf("No claims found")
			c.AbortWithStatus(401)
			return
		}

		state := randSeq(32)
		session.Set(oidcTokenRequestStateKey, state)
		session.Save()

		logger.Debugf("new authn round started, state %v saved to session", state)

		c.Abort()
		c.Redirect(http.StatusFound, oauth2Config.AuthCodeURL(state))
	}
}

func handleOAuth2Callback(oauth2Config oauth2.Config, verifier *oidc.IDTokenVerifier) func(c *gin.Context, storedState string) (*claims, error) {

	logger := log.WithField("prefix", "handleOAuth2Callback")

	return func(c *gin.Context, storedState string) (*claims, error) {
		r := c.Request

		responseState := r.URL.Query().Get("state")
		if responseState != storedState {
			return nil, fmt.Errorf("response state %v doesn't equal the stored state: %v", responseState, storedState)
		}

		oauth2Token, err := oauth2Config.Exchange(context.TODO(), r.URL.Query().Get("code"))
		if err != nil {
			logger.Errorf("failed to obtain OAuth2 token: %v", err)
			return nil, fmt.Errorf("failed to obtain OAuth2 token: %w", err)
		}

		// Extract the ID Token from OAuth2 token.
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			logger.Errorf("failed to extract ID token: %v", err)
			return nil, fmt.Errorf("failed to extract ID token: %w", err)
		}

		// Parse and verify ID Token payload.
		idToken, err := verifier.Verify(context.TODO(), rawIDToken)
		if err != nil {
			logger.Errorf("failed to verify ID token: %v", err)
			return nil, fmt.Errorf("failed to verify ID token: %w", err)
		}

		// Extract custom claims
		var claims claims
		if err := idToken.Claims(&claims); err != nil {
			logger.Errorf("failed to extract claims from ID token: %v", err)
			return nil, fmt.Errorf("failed to extract claims from ID token: %w", err)
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
