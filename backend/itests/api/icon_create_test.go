package api

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type iconCreateTestSuite struct {
	apiTestSuite
}

func TestIconCreateTestSuite(t *testing.T) {
	suite.Run(t, &iconCreateTestSuite{})
}

func (s *apiTestSuite) TestPOSTShouldFailWith403WithoutCREATE_ICONPrivilegeTest() {}
