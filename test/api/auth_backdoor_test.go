package api_tests

import (
	"testing"

	"igo-repo/internal/app/security/authn"
	"igo-repo/internal/app/security/authr"
	"igo-repo/internal/repositories/gitrepo"
	"igo-repo/test/repositories/git_tests"
	"igo-repo/test/testdata"

	"github.com/stretchr/testify/suite"
)

type authBackDoorTestSuite struct {
	ApiTestSuite
}

func TestAuthBackDoorTestSuite(t *testing.T) {
	suite.Run(t, &authBackDoorTestSuite{ApiTestSuite: apiTestSuites("apitests_backdoor", []git_tests.GitTestRepo{gitrepo.Local{}})[0]})
}

func (s *authBackDoorTestSuite) BeforeTest(suiteName string, testName string) {
	s.ApiTestSuite.config = s.ApiTestSuite.initConfig()
	if suiteName != "authBackDoorTestSuite" {
		return
	}
	switch testName {
	case "TestBackDoorMustntBeAvailableByDefault":
		{
		}
	case "TestBackDoorShouldBeAvailableWhenEnabled":
		{
			s.ApiTestSuite.config.EnableBackdoors = true
		}
	case "TestBackDoorShouldAllowToSetPrivileges":
		{
			s.ApiTestSuite.config.EnableBackdoors = true
		}
	default:
		{
			panic("Unexpected testName: " + testName)
		}
	}
	startErr := s.startApp(s.ApiTestSuite.config)
	if startErr != nil {
		panic(startErr)
	}
}

func (s *authBackDoorTestSuite) TestBackDoorMustntBeAvailableByDefault() {
	session := s.Client.mustLogin(nil)
	resp, err := session.put(&testRequest{
		path: authenticationBackdoorPath,
		json: true,
		body: []authr.PermissionID{},
	})
	s.NoError(err)
	s.Equal(405, resp.statusCode)
}

func (s *authBackDoorTestSuite) TestBackDoorShouldBeAvailableWhenEnabled() {
	session := s.Client.mustLogin(nil)
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

	session := s.Client.mustLogin(nil)

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
