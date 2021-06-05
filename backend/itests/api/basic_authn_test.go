package api

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type basicAuthnTestSuite struct {
	apiTestSuite
}

func TestBasicAuthnTestSuite(t *testing.T) {
	suite.Run(t, &basicAuthnTestSuite{})
}

func (s *basicAuthnTestSuite) TestShouldFailWith401WithoutCredentials() {
	req := request{
		path:               "/info",
		credentials:        &requestCredentials{"", ""},
		expectedStatusCode: 401,
		body:               nil,
		testSuite:          &s.Suite,
	}
	resp, requestError := s.get(&req)
	s.NoError(requestError)
	challenge, hasChallange := resp.headers["Www-Authenticate"]
	s.True(hasChallange)
	s.Equal("Basic", challenge[0])
}

func (s *basicAuthnTestSuite) TestShouldFailWith401WithWrongCredentials() {
	reqCreds, makeReqCredErr := makeRequestCredentials(basicAuthScheme, "ux", "definitely-wrong-password....!~")
	s.Require().NoError(makeReqCredErr)
	req := request{
		path:               "/info",
		credentials:        &reqCreds,
		expectedStatusCode: 401,
		body:               nil,
		testSuite:          &s.Suite,
	}
	resp, requestError := s.get(&req)
	s.NoError(requestError)
	challenge, hasChallange := resp.headers["Www-Authenticate"]
	s.True(hasChallange)
	s.Equal("Basic", challenge[0])
}

func (s *basicAuthnTestSuite) TestShouldPasssWithCorrectCredentials() {
	reqCreds, makeReqCredErr := makeRequestCredentials(basicAuthScheme, defaultCredentials.User, defaultCredentials.Password)
	s.Require().NoError(makeReqCredErr)
	req := request{
		path:               "/info",
		credentials:        &reqCreds,
		expectedStatusCode: 200,
		body:               nil,
		testSuite:          &s.Suite,
	}
	resp, requestError := s.get(&req)
	s.NoError(requestError)
	_, hasChallange := resp.headers["Www-Authenticate"]
	s.False(hasChallange)
}
