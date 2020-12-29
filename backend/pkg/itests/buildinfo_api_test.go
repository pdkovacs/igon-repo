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

func (s *buildinfoAPITestSuite) TestMustIncludeVersionInfo() {
	defer terminateTestServer()
	startTestServer()

	expected := build.GetInfo()

	buildInfo := build.Info{}
	err := get("/info", &buildInfo)
	s.NoError(err)
	s.Equal(expected, buildInfo)
}
