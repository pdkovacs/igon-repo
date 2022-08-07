package api

import (
	"encoding/base64"
	"fmt"
	"igo-repo/internal/config"
)

func makeRequestCredentials(authnScheme string, usernameOrToken string, password string) (requestCredentials, error) {
	switch authnScheme {
	case config.BasicAuthentication:
		{
			username := usernameOrToken
			auth := username + ":" + password
			return requestCredentials{
				headerName:  "Authorization",
				headerValue: "Basic " + base64.StdEncoding.EncodeToString([]byte(auth)),
			}, nil
		}
	case config.OIDCAuthentication:
		return requestCredentials{}, fmt.Errorf("Unsupported authorization scheme: OIDC")
	default:
		return requestCredentials{}, fmt.Errorf("Unexpected authorization scheme: %v", authnScheme)
	}
}
