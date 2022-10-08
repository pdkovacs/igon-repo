package api_tests

import (
	"igo-repo/internal/config"
	"testing"

	"github.com/stretchr/testify/suite"
)

type buildinfoAPITestSuite struct {
	suite.Suite
	apiTestSuite
}

func TestBuildinfoAPITestSuite(t *testing.T) {
	suite.Run(t, &buildinfoAPITestSuite{})
}

func (s *buildinfoAPITestSuite) TestMustIncludeVersionInfo() {
	expected := config.GetBuildInfo()

	session := s.client.mustLogin(nil)
	req := testRequest{
		path:          "/app-info",
		respBodyProto: &config.BuildInfo{},
	}
	resp, err := session.get(&req)
	s.NoError(err)
	s.Equal(200, resp.statusCode)
	s.Equal(&expected, resp.body)
}
