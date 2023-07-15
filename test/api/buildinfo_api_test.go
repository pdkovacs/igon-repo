package api_tests

import (
	"iconrepo/internal/config"
	"iconrepo/internal/repositories/blobstore/git"
	git_tests "iconrepo/test/repositories/blobstore/git"
	"testing"

	"github.com/stretchr/testify/suite"
)

type buildinfoAPITestSuite struct {
	suite.Suite
	ApiTestSuite
}

func TestBuildinfoAPITestSuite(t *testing.T) {
	suite.Run(t, &buildinfoAPITestSuite{ApiTestSuite: apiTestSuites("apitests_buildinfo", []git_tests.GitTestRepo{git.Local{}})[0]})
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
	s.Equal(200, resp.statusCode)
	s.Equal(&expected, resp.body)
}
