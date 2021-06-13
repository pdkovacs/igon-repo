package api

import (
	"testing"

	"github.com/pdkovacs/igo-repo/backend/itests/common"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authn"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authr"
	"github.com/pdkovacs/igo-repo/backend/pkg/services"
	"github.com/pdkovacs/igo-repo/backend/pkg/web"
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
		body: web.BackdoorAuthorization{
			Username:    defaultCredentials.Username,
			Permissions: []authr.PermissionID{},
		},
	})
	s.NoError(err)
	s.Equal(404, resp.statusCode)
}

func (s *authBackDoorTestSuite) TestBackDoorShouldBeAvailableWhenEnabled() {
	resp, err := s.client.setAuthorization(
		s.client.MustCreateCookieJar(),
		web.BackdoorAuthorization{
			Username:    defaultCredentials.Username,
			Permissions: []authr.PermissionID{},
		},
	)
	s.NoError(err)
	s.Equal(200, resp.statusCode)
}

func (s *authBackDoorTestSuite) TestBackDoorShouldAllowToSetPrivileges() {
	requestedAuthorization := web.BackdoorAuthorization{
		Username:    defaultCredentials.Username,
		Permissions: []authr.PermissionID{"galagonya", "ide-oda"},
	}
	expectedUserInfo := services.UserInfo{
		UserId:      authn.LocalDomain.CreateUserID(defaultCredentials.Username),
		Permissions: requestedAuthorization.Permissions,
	}

	cjar := s.client.MustCreateCookieJar()

	resp, err := s.client.setAuthorization(cjar, requestedAuthorization)
	s.NoError(err)
	s.Equal(resp.statusCode, 200)

	resp, errUserInfo := s.client.doRequest("GET", &testRequest{
		path:          authenticationBackdoorPath,
		jar:           cjar,
		respBodyProto: &services.UserInfo{},
	})
	s.NoError(errUserInfo)
	s.Equal(200, resp.statusCode)
	s.Equal(&expectedUserInfo, resp.body)
}
