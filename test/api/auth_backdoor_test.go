package api_tests

import (
	"testing"

	"igo-repo/internal/app/security/authn"
	"igo-repo/internal/app/security/authr"
	common_test "igo-repo/test/common"
	"igo-repo/test/testdata"

	"github.com/stretchr/testify/suite"
)

type authBackDoorTestSuite struct {
	apiTestSuite
}

func TestAuthBackDoorTestSuite(t *testing.T) {
	suite.Run(t, &authBackDoorTestSuite{})
}

func (s *authBackDoorTestSuite) BeforeTest(suiteName string, testName string) {
	serverConfig := common_test.CloneConfig(s.defaultConfig)
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
	s.apiTestSuite.startApp(serverConfig)
}

func (s *authBackDoorTestSuite) TestBackDoorMustntBeAvailableByDefault() {
	session := s.client.mustLogin(nil)
	resp, err := session.put(&testRequest{
		path: authenticationBackdoorPath,
		json: true,
		body: []authr.PermissionID{},
	})
	s.NoError(err)
	s.Equal(404, resp.statusCode)
}

func (s *authBackDoorTestSuite) TestBackDoorShouldBeAvailableWhenEnabled() {
	session := s.client.mustLogin(nil)
	resp, err := session.setAuthorization(
		[]authr.PermissionID{},
	)
	s.NoError(err)
	s.Equal(200, resp.statusCode)
}

func (s *authBackDoorTestSuite) TestBackDoorShouldAllowToSetPrivileges() {
	requestedAuthorization := []authr.PermissionID{"galagonya", "ide-oda"}
	userID := authn.LocalDomain.CreateUserID(testdata.DefaultCredentials.Username)
	expectedUserInfo := authr.UserInfo{
		UserId:      userID,
		Permissions: requestedAuthorization,
		DisplayName: userID.String(),
	}

	session := s.client.mustLogin(nil)

	resp, err := session.setAuthorization(requestedAuthorization)
	s.NoError(err)
	s.Equal(resp.statusCode, 200)

	resp, errUserInfo := session.get(&testRequest{
		path:          authenticationBackdoorPath,
		respBodyProto: &authr.UserInfo{},
	})
	s.NoError(errUserInfo)
	s.Equal(200, resp.statusCode)
	s.Equal(&expectedUserInfo, resp.body)
}
