package server

import (
	"iconrepo/internal/config"
	blobstore_tests "iconrepo/test/repositories/blobstore"
	"iconrepo/test/repositories/indexing"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type buildinfoAPITestSuite struct {
	suite.Suite
	ApiTestSuite
}

func TestBuildinfoAPITestSuite(t *testing.T) {
	suite.Run(
		t,
		&buildinfoAPITestSuite{
			ApiTestSuite: apiTestSuites(
				"apitests_buildinfo",
				[]blobstore_tests.TestBlobstoreController{blobstore_tests.DefaultBlobstoreController},
				[]indexing.IndexTestRepoController{*indexing.DefaultIndexTestRepoController()},
			)[0]})
}

func (s *buildinfoAPITestSuite) TestMustIncludeVersionInfo() {
	expected := config.GetBuildInfo()

	session := s.Client.mustLogin(nil)
	req := testRequest{
		path:          "/app-info",
		respBodyProto: &config.BuildInfo{},
	}
	resp, err := session.get(&req)
	s.NoError(err)
	s.Equal(http.StatusOK, resp.statusCode)
	s.Equal(&expected, resp.body)
}
