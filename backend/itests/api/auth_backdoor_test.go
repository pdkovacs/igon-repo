package api

import (
	"errors"
	"net/http/cookiejar"
	"testing"

	"github.com/pdkovacs/igo-repo/backend/itests/common"
	"github.com/pdkovacs/igo-repo/backend/pkg/security"
	"github.com/stretchr/testify/suite"
)

type authBackDoorTestSuite struct {
	apiTestSuite
}

func TestAuthBackDoorTestSuite(t *testing.T) {
	suite.Run(t, &authBackDoorTestSuite{})
}

func (s *authBackDoorTestSuite) BeforeTest(suiteName string, testName string) {
	serverConfig := common.CloneConfig(s.defaultConfig)
	if suiteName != "authBackDoorTestSuite" {
		return
	}
	switch testName {
	case "TestBackDoorMustntBeAvailableByDefault":
		{
		}
	case "TestBackDoorShouldBeAvailableWhenEnabled":
		{
			serverConfig.EnableBackdoors = true
		}
	case "TestBackDoorShouldAllowToSetPrivileges":
		{
			serverConfig.EnableBackdoors = true
		}
	default:
		{
			panic("Unexpected testName: " + testName)
		}
	}
	s.startTestServer(serverConfig)
}

func (s *authBackDoorTestSuite) TestBackDoorMustntBeAvailableByDefault() {
	var userName = defaultCredentials.User
	var password = defaultCredentials.Password
	_, err := s.client.req("PUT", &requestType{
		path: authenticationBackdoorPath,
		credentials: &requestCredentials{
			headerName:  userName,
			headerValue: password,
		},
		expectedStatusCode: 404,
		json:               true,
		body: security.Authorization{
			Username:   userName,
			Privileges: []string{},
		},
	})
	s.True(errors.Is(err, errUnexpectedStatusCode))
}

func (s *authBackDoorTestSuite) TestBackDoorShouldBeAvailableWhenEnabled() {
	var err error

	userName := defaultCredentials.User
	password := defaultCredentials.Password

	var credentials requestCredentials
	credentials, err = makeRequestCredentials(s.server.Configuration.AuthenticationType, userName, password)
	if err != nil {
		panic(err)
	}
	_, err = s.client.req("PUT", &requestType{
		path:               authenticationBackdoorPath,
		credentials:        &credentials,
		expectedStatusCode: 200,
		json:               true,
		body: security.Authorization{
			Username:   userName,
			Privileges: []string{},
		},
	})
	s.NoError(err)
}

func (s *authBackDoorTestSuite) TestBackDoorShouldAllowToSetPrivileges() {
	var err error

	userName := defaultCredentials.User
	password := defaultCredentials.Password

	requestedAuthorization := security.Authorization{
		Username:   userName,
		Privileges: []string{"galagonya", "ide-oda"},
	}

	cjar, errCreatingJar := cookiejar.New(nil)
	if errCreatingJar != nil {
		panic(errCreatingJar)
	}

	var credentials requestCredentials
	credentials, err = makeRequestCredentials(s.server.Configuration.AuthenticationType, userName, password)
	if err != nil {
		panic(err)
	}
	_, err = s.client.req("PUT", &requestType{
		path:               authenticationBackdoorPath,
		credentials:        &credentials,
		jar:                cjar,
		expectedStatusCode: 200,
		json:               true,
		body:               requestedAuthorization,
	})
	s.NoError(err)

	userInfoResponse, errUserInfo := s.client.req("GET", &requestType{
		path:               "/user",
		jar:                cjar,
		expectedStatusCode: 200,
		respBodyProto:      &security.Authorization{},
	})
	s.NoError(errUserInfo)
	s.Equal(&requestedAuthorization, userInfoResponse.body)
}
