package api

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type iconGetTestSuite struct {
	iconTestSuite
}

func TestIconGetTestSuite(t *testing.T) {
	suite.Run(t, &iconGetTestSuite{})
}

func (s *iconGetTestSuite) TestReturnAllIconDescriptions() {
	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(testIconInputData)
	all, err := session.describeAllIcons()
	s.NoError(err)
	s.Equal(testIconDataResponse, all)

	s.assertEndState()
}

func (s *iconGetTestSuite) TestDescribeSingleIcon() {
	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(testIconInputData)
	statusCode, one, err := session.describeIcon(testIconInputData[0].Name)
	s.NoError(err)
	s.Equal(200, statusCode)
	s.Equal(testIconDataResponse[0], one)

	s.assertEndState()
}

func (s *iconGetTestSuite) TestReturn404ForNonExistentIcon() {
	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(testIconInputData)
	statusCode, _, err := session.describeIcon("somenonexistentname")
	s.True(errors.Is(err, errJSONUnmarshal))
	s.Equal(404, statusCode)

	s.assertEndState()
}
