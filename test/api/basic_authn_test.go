package api

import (
	"testing"

	"igo-repo/internal/config"
	"igo-repo/test/testdata"

	"github.com/stretchr/testify/suite"
)

type basicAuthnTestSuite struct {
	apiTestSuite
}

func TestBasicAuthnTestSuite(t *testing.T) {
	suite.Run(t, &basicAuthnTestSuite{})
}

func (s *basicAuthnTestSuite) TestShouldFailWith401WithoutCredentials() {
	req := testRequest{
		path:          "/icon",
		credentials:   &requestCredentials{"", ""},
		respBodyProto: nil,
	}
	resp, requestError := s.client.get(&req)
	s.NoError(requestError)
	s.Equal(401, resp.statusCode)
	challenge, hasChallange := resp.headers["Www-Authenticate"]
	s.True(hasChallange)
	s.Equal("Basic", challenge[0])
}

func (s *basicAuthnTestSuite) TestShouldFailWith401WithWrongCredentials() {
	reqCreds, makeReqCredErr := makeRequestCredentials(config.BasicAuthentication, "ux", "definitely-wrong-password....!~")
	s.Require().NoError(makeReqCredErr)
	req := testRequest{
		path:          "/icon",
		credentials:   &reqCreds,
		respBodyProto: nil,
	}
	resp, requestError := s.client.get(&req)
	s.NoError(requestError)
	s.Equal(401, resp.statusCode)
	challenge, hasChallange := resp.headers["Www-Authenticate"]
	s.True(hasChallange)
	s.Equal("Basic", challenge[0])
}

func (s *basicAuthnTestSuite) TestShouldPasssWithCorrectCredentials() {
	reqCreds, makeReqCredErr := makeRequestCredentials(config.BasicAuthentication, testdata.DefaultCredentials.Username, testdata.DefaultCredentials.Password)
	s.Require().NoError(makeReqCredErr)
	req := testRequest{
		path:          "/icon",
		credentials:   &reqCreds,
		respBodyProto: nil,
	}
	resp, requestError := s.client.get(&req)
	s.NoError(requestError)
	s.Equal(200, resp.statusCode)
	_, hasChallange := resp.headers["Www-Authenticate"]
	s.False(hasChallange)
}
