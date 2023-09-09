package server

import (
	"testing"

	"iconrepo/internal/app/security/authn"
	blobstore_tests "iconrepo/test/repositories/blobstore"
	"iconrepo/test/repositories/indexing"
	"iconrepo/test/testdata"

	"github.com/stretchr/testify/suite"
)

type basicAuthnTestSuite struct {
	ApiTestSuite
}

func TestBasicAuthnTestSuite(t *testing.T) {
	suite.Run(
		t, &basicAuthnTestSuite{
			ApiTestSuite: apiTestSuites(
				"apitests_basicauthn",
				[]blobstore_tests.TestBlobstoreController{blobstore_tests.DefaultBlobstoreController},
				[]indexing.IndexTestRepoController{*indexing.DefaultIndexTestRepoController()})[0]},
	)
}

func (s *basicAuthnTestSuite) TestShouldFailWith401WithoutCredentials() {
	req := testRequest{
		path:          "/icon",
		credentials:   &requestCredentials{"", ""},
		respBodyProto: nil,
	}
	resp, requestError := s.Client.get(&req)
	s.NoError(requestError)
	s.Equal(401, resp.statusCode)
	challenge, hasChallange := resp.headers["Www-Authenticate"]
	s.True(hasChallange)
	s.Equal("Basic", challenge[0])
}

func (s *basicAuthnTestSuite) TestShouldFailWith401WithWrongCredentials() {
	reqCreds, makeReqCredErr := makeRequestCredentials(authn.SchemeBasic, "ux", "definitely-wrong-password....!~")
	s.Require().NoError(makeReqCredErr)
	req := testRequest{
		path:          "/icon",
		credentials:   &reqCreds,
		respBodyProto: nil,
	}
	resp, requestError := s.Client.get(&req)
	s.NoError(requestError)
	s.Equal(401, resp.statusCode)
	challenge, hasChallange := resp.headers["Www-Authenticate"]
	s.True(hasChallange)
	s.Equal("Basic", challenge[0])
}

func (s *basicAuthnTestSuite) TestShouldPasssWithCorrectCredentials() {
	reqCreds, makeReqCredErr := makeRequestCredentials(authn.SchemeBasic, testdata.DefaultCredentials.Username, testdata.DefaultCredentials.Password)
	s.Require().NoError(makeReqCredErr)
	req := testRequest{
		path:          "/icon",
		credentials:   &reqCreds,
		respBodyProto: nil,
	}
	resp, requestError := s.Client.get(&req)
	s.NoError(requestError)
	s.Equal(200, resp.statusCode)
	_, hasChallange := resp.headers["Www-Authenticate"]
	s.False(hasChallange)
}
