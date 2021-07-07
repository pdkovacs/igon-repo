package api

import (
	"testing"

	"github.com/pdkovacs/igo-repo/internal/auxiliaries"
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
	expected := auxiliaries.GetBuildInfo()

	session := s.client.mustLogin(nil)
	req := testRequest{
		path:          "/info",
		respBodyProto: &auxiliaries.BuildInfo{},
	}
	resp, err := session.get(&req)
	s.NoError(err)
	s.Equal(200, resp.statusCode)
	s.Equal(&expected, resp.body)
}