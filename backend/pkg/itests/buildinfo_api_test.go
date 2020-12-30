package itests

import (
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/build"
	"github.com/stretchr/testify/suite"
)

type buildinfoAPITestSuite struct {
	suite.Suite
}

func TestBuildinfoAPITestSuite(t *testing.T) {
	suite.Run(t, &buildinfoAPITestSuite{})
}

func (s *buildinfoAPITestSuite) BeforeTest(suiteName, testName string) {
	startTestServer()
}

func (s *buildinfoAPITestSuite) AfterTest(suiteName, testName string) {
	terminateTestServer()
}

func (s *buildinfoAPITestSuite) TestMustIncludeVersionInfo() {
	expected := build.GetInfo()

	buildInfo := build.Info{}
	err := get("/info", 200, &s.Suite, &buildInfo)
	s.Require().NoError(err)
	s.Equal(expected, buildInfo)
}
