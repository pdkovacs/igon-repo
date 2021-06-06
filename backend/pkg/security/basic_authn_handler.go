package security

import (
	"encoding/base64"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
)

// BasicConfig holds the configuration for the Basic authentication scheme
type BasicConfig struct {
	PasswordCredentialsList []auxiliaries.PasswordCredentials
}

func decodeBasicAuthnHeaderValue(headerValue string) (username, password string, decodeOK bool) {
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

func basicScheme(config BasicConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorized := false

		session := sessions.Default(c)
		user := session.Get(userKey)
		if user != nil {
			authorized = true
		} else {
			authnHeaderValue, hasHeader := c.Request.Header["Authorization"]
			if hasHeader {
				username, password, decodeOK := decodeBasicAuthnHeaderValue(authnHeaderValue[0])
				if decodeOK {
					for _, pc := range config.PasswordCredentialsList {
						if pc.User == username && pc.Password == password {
							session.Set(userKey, UserSession{username, []string{}})
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
