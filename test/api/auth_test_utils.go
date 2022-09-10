package api

import (
	"encoding/base64"
	"fmt"
	"igo-repo/internal/app/security/authn"
)

func makeRequestCredentials(authnScheme authn.AuthenticationScheme, usernameOrToken string, password string) (requestCredentials, error) {
	switch authnScheme {
	case authn.SchemeBasic:
		{
			username := usernameOrToken
			auth := username + ":" + password
			return requestCredentials{
				headerName:  "Authorization",
				headerValue: "Basic " + base64.StdEncoding.EncodeToString([]byte(auth)),
			}, nil
		}
	case authn.SchemeOIDC:
		return requestCredentials{}, fmt.Errorf("unsupported authorization scheme: OIDC")
	default:
		return requestCredentials{}, fmt.Errorf("unexpected authorization scheme: %v", authnScheme)
	}
}
