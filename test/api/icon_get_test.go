package api

import (
	"errors"
	"testing"

	"github.com/pdkovacs/igo-repo/test/testdata"
	"github.com/stretchr/testify/suite"
)

type iconGetTestSuite struct {
	iconTestSuite
}

func TestIconGetTestSuite(t *testing.T) {
	suite.Run(t, &iconGetTestSuite{})
}

func (s *iconGetTestSuite) TestReturnAllIconDescriptions() {
	dataIn, dataOut := testdata.Get()
	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(dataIn)
	all, err := session.describeAllIcons()
	s.NoError(err)
	s.assertResponseIconSetsEqual(dataOut, all)

	s.assertEndState()
}

func (s *iconGetTestSuite) TestDescribeSingleIcon() {
	dataIn, dataOut := testdata.Get()
	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(dataIn)
	statusCode, one, err := session.describeIcon(dataIn[0].Name)
	s.NoError(err)
	s.Equal(200, statusCode)
	s.assertResponseIconsEqual(dataOut[0], one)

	s.assertEndState()
}

func (s *iconGetTestSuite) TestReturn404ForNonExistentIcon() {
	dataIn, _ := testdata.Get()
	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(dataIn)
	statusCode, _, err := session.describeIcon("somenonexistentname")
	s.True(errors.Is(err, errJSONUnmarshal))
	s.Equal(404, statusCode)

	s.assertEndState()
}
