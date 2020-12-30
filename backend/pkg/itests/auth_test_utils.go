package itests

import (
	"encoding/base64"
	"fmt"
)

// authScheme is the authentication authScheme
type authScheme int

const (
	// basic is the basic authentication type
	basicAuthScheme authScheme = iota
	// oidc is the basic authentication type
	oidcAuthScheme
)

func makeRequestCredentials(authnScheme authScheme, usernameOrToken string, password string) (requestCredentials, error) {
	switch authnScheme {
	case basicAuthScheme:
		{
			username := usernameOrToken
			auth := username + ":" + password
			return requestCredentials{
				headerName:  "Authorization",
				headerValue: "Basic " + base64.StdEncoding.EncodeToString([]byte(auth)),
			}, nil
		}
	case oidcAuthScheme:
		return requestCredentials{}, fmt.Errorf("Unsupported authorization scheme: OIDC")
	default:
		return requestCredentials{}, fmt.Errorf("Unexpected authorization scheme: %v", authnScheme)
	}
}
