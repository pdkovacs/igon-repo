package api

import (
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
	"github.com/stretchr/testify/suite"
)

type basicAuthnTestSuite struct {
	apiTestSuite
}

func TestBasicAuthnTestSuite(t *testing.T) {
	suite.Run(t, &basicAuthnTestSuite{})
}

func (s *basicAuthnTestSuite) TestShouldFailWith401WithoutCredentials() {
	req := requestType{
		path:               "/info",
		credentials:        &requestCredentials{"", ""},
		expectedStatusCode: 401,
		respBodyProto:      nil,
	}
	resp, requestError := s.client.get(&req)
	s.NoError(requestError)
	challenge, hasChallange := resp.headers["Www-Authenticate"]
	s.True(hasChallange)
	s.Equal("Basic", challenge[0])
}

func (s *basicAuthnTestSuite) TestShouldFailWith401WithWrongCredentials() {
	reqCreds, makeReqCredErr := makeRequestCredentials(auxiliaries.BasicAuthentication, "ux", "definitely-wrong-password....!~")
	s.Require().NoError(makeReqCredErr)
	req := requestType{
		path:               "/info",
		credentials:        &reqCreds,
		expectedStatusCode: 401,
		respBodyProto:      nil,
	}
	resp, requestError := s.client.get(&req)
	s.NoError(requestError)
	challenge, hasChallange := resp.headers["Www-Authenticate"]
	s.True(hasChallange)
	s.Equal("Basic", challenge[0])
}

func (s *basicAuthnTestSuite) TestShouldPasssWithCorrectCredentials() {
	reqCreds, makeReqCredErr := makeRequestCredentials(auxiliaries.BasicAuthentication, defaultCredentials.User, defaultCredentials.Password)
	s.Require().NoError(makeReqCredErr)
	req := requestType{
		path:               "/info",
		credentials:        &reqCreds,
		expectedStatusCode: 200,
		respBodyProto:      nil,
	}
	resp, requestError := s.client.get(&req)
	s.NoError(requestError)
	_, hasChallange := resp.headers["Www-Authenticate"]
	s.False(hasChallange)
}
