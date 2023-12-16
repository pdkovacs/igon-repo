package server

import (
	"net/http"
	"testing"

	"iconrepo/test/testdata"

	"github.com/stretchr/testify/suite"
)

type iconGetTestSuite struct {
	IconTestSuite
}

func TestIconGetTestSuite(t *testing.T) {
	t.Parallel()
	for _, iconSuite := range IconTestSuites("api_iconget") {
		suite.Run(t, &iconGetTestSuite{IconTestSuite: iconSuite})
	}
}

func (s *iconGetTestSuite) TestReturnAllIconDescriptions() {
	dataIn, dataOut := testdata.Get()
	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)
	all, err := session.DescribeAllIcons(s.Ctx)
	s.NoError(err)
	s.AssertResponseIconSetsEqual(dataOut, all)

	s.AssertEndState()
}

func (s *iconGetTestSuite) TestDescribeSingleIcon() {
	dataIn, dataOut := testdata.Get()
	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)
	statusCode, one, err := session.describeIcon(dataIn[0].Name)
	s.NoError(err)
	s.Equal(http.StatusOK, statusCode)
	s.assertResponseIconsEqual(dataOut[0], one)

	s.AssertEndState()
}

func (s *iconGetTestSuite) TestReturn404ForNonExistentIcon() {
	dataIn, _ := testdata.Get()
	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)
	statusCode, _, err := session.describeIcon("somenonexistentname")
	s.Error(err)
	s.ErrorIs(err, errJSONUnmarshal)
	s.Equal(http.StatusNotFound, statusCode)

	s.AssertEndState()
}
