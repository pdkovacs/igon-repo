package api

import (
	"encoding/base64"
	"fmt"

	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
)

func makeRequestCredentials(authnScheme string, usernameOrToken string, password string) (requestCredentials, error) {
	switch authnScheme {
	case auxiliaries.BasicAuthentication:
		{
			username := usernameOrToken
			auth := username + ":" + password
			return requestCredentials{
				headerName:  "Authorization",
				headerValue: "Basic " + base64.StdEncoding.EncodeToString([]byte(auth)),
			}, nil
		}
	case auxiliaries.OIDCAuthentication:
		return requestCredentials{}, fmt.Errorf("Unsupported authorization scheme: OIDC")
	default:
		return requestCredentials{}, fmt.Errorf("Unexpected authorization scheme: %v", authnScheme)
	}
}
