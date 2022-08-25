package http

import (
	"encoding/base64"
	"strings"

	"igo-repo/internal/app/security/authn"
	"igo-repo/internal/app/services"
	"igo-repo/internal/config"
	"igo-repo/internal/logging"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// basicConfig holds the configuration for the Basic authentication scheme
type basicConfig struct {
	PasswordCredentialsList []config.PasswordCredentials
}

func decodeBasicAuthnHeaderValue(headerValue string) (userid string, password string, decodeOK bool) {
	s := strings.SplitN(headerValue, " ", 2)
	if len(s) != 2 {
		return "", "", false
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return "", "", false
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return "", "", false
	}

	return pair[0], pair[1], true
}

func checkBasicAuthentication(options basicConfig, userService services.UserService, log zerolog.Logger) func(c *gin.Context) {
	logger := logging.CreateMethodLogger(log, "basic-authn")
	logger.Debug().Msgf("options.PasswordCredentials size: %d", len(options.PasswordCredentialsList))

	return func(c *gin.Context) {
		authorized := false

		session := sessions.Default(c)
		user := session.Get(UserKey)
		if user != nil {
			authorized = true
		} else {
			authnHeaderValue, hasHeader := c.Request.Header["Authorization"]
			if hasHeader {
				username, password, decodeOK := decodeBasicAuthnHeaderValue(authnHeaderValue[0])
				if decodeOK {
					for _, pc := range options.PasswordCredentialsList {
						if pc.Username == username && pc.Password == password {
							userId := authn.LocalDomain.CreateUserID(username)
							userInfo := userService.GetUserInfo(userId)
							session.Set(UserKey, SessionData{userInfo})
							session.Save()
							authorized = true
							break
						}
					}
				}
			}
		}
		session.Save()

		if authorized {
			c.Next()
		} else {
			c.Header("WWW-Authenticate", "Basic")
			c.AbortWithStatus(401)
		}
	}
}

func basicScheme(options basicConfig, userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Status(200)
	}
}
