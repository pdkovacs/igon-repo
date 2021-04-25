package itests

import (
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
	"github.com/stretchr/testify/suite"
)

type buildinfoAPITestSuite struct {
	suite.Suite
}

func TestBuildinfoAPITestSuite(t *testing.T) {
	suite.Run(t, &buildinfoAPITestSuite{})
}

func (s *buildinfoAPITestSuite) BeforeTest(suiteName, testName string) {
	startTestServer(defaultOptions)
}

func (s *buildinfoAPITestSuite) AfterTest(suiteName, testName string) {
	terminateTestServer()
}

func (s *buildinfoAPITestSuite) TestMustIncludeVersionInfo() {
	expected := auxiliaries.GetBuildInfo()

	req := request{
		path:               "/info",
		testSuite:          &s.Suite,
		expectedStatusCode: 200,
		body:               &auxiliaries.BuildInfo{},
	}
	resp, err := get(&req)
	s.Require().NoError(err)
	s.Equal(&expected, resp.body)
}
