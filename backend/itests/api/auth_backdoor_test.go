package api

import (
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
	creds := s.client.makeRequestCredentials(defaultCredentials)
	resp, err := s.client.doRequest("PUT", &testRequest{
		path:        authenticationBackdoorPath,
		credentials: &creds,
		json:        true,
		body: security.Authorization{
			Username:   defaultCredentials.User,
			Privileges: []string{},
		},
	})
	s.NoError(err)
	s.Equal(404, resp.statusCode)
}

func (s *authBackDoorTestSuite) TestBackDoorShouldBeAvailableWhenEnabled() {
	var credentials requestCredentials
	credentials = s.client.makeRequestCredentials(defaultCredentials)
	resp, err := s.client.doRequest("PUT", &testRequest{
		path:        authenticationBackdoorPath,
		credentials: &credentials,
		json:        true,
		body: security.Authorization{
			Username:   defaultCredentials.User,
			Privileges: []string{},
		},
	})
	s.NoError(err)
	s.Equal(200, resp.statusCode)
}

func (s *authBackDoorTestSuite) TestBackDoorShouldAllowToSetPrivileges() {
	var err error
	var resp testResponse

	requestedAuthorization := security.Authorization{
		Username:   defaultCredentials.User,
		Privileges: []string{"galagonya", "ide-oda"},
	}

	cjar, errCreatingJar := cookiejar.New(nil)
	if errCreatingJar != nil {
		panic(errCreatingJar)
	}

	var credentials requestCredentials
	credentials = s.client.makeRequestCredentials(defaultCredentials)
	if err != nil {
		panic(err)
	}
	resp, err = s.client.doRequest("PUT", &testRequest{
		path:        authenticationBackdoorPath,
		credentials: &credentials,
		jar:         cjar,
		json:        true,
		body:        requestedAuthorization,
	})
	s.NoError(err)
	s.Equal(200, resp.statusCode)

	resp, errUserInfo := s.client.doRequest("GET", &testRequest{
		path:          "/user",
		jar:           cjar,
		respBodyProto: &security.Authorization{},
	})
	s.NoError(errUserInfo)
	s.Equal(200, resp.statusCode)
	s.Equal(&requestedAuthorization, resp.body)
}
