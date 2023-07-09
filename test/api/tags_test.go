package api_tests

import (
	"testing"

	"iconrepo/internal/app/security/authr"
	"iconrepo/test/testdata"

	"github.com/stretchr/testify/suite"
)

type tagsTestSuite struct {
	IconTestSuite
}

func TestTagsTestSuite(t *testing.T) {
	t.Parallel()
	for _, iconSuite := range IconTestSuites("api_tags") {
		suite.Run(t, &tagsTestSuite{IconTestSuite: iconSuite})
	}
}

func (s *tagsTestSuite) TestAddingFailsWithoutPermission() {
	dataIn, dataOut := testdata.Get()
	iconIn := dataIn[0]
	tag := "Ahoj"

	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)

	session.mustSetAllPermsExcept([]authr.PermissionID{authr.ADD_TAG})

	statusCode, err := session.addTag(iconIn.Name, tag)
	s.NoError(err)
	s.Equal(403, statusCode)

	respIcons := session.mustDescribeAllIcons()
	s.AssertResponseIconSetsEqual(dataOut, respIcons)
}

func (s *tagsTestSuite) TestAddingSucceedssWithRequiredPermission() {
	dataIn, dataOut := testdata.Get()
	iconIn := dataIn[0]
	iconOut := &dataOut[0]
	tag := "Ahoj"

	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)

	session.setAuthorization([]authr.PermissionID{authr.ADD_TAG})

	statusCode, err := session.addTag(iconIn.Name, tag)
	s.NoError(err)
	s.Equal(201, statusCode)

	iconOut.Tags = []string{tag}

	respIcons := session.mustDescribeAllIcons()
	s.AssertResponseIconSetsEqual(dataOut, respIcons)
}

func (s *tagsTestSuite) TestDeletingFailsWithoutPermission() {
	dataIn, dataOut := testdata.Get()
	iconIn := dataIn[0]
	iconOut := &dataOut[0]
	tag := "Ahoj"

	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)
	statusCode, err := session.addTag(iconIn.Name, tag)
	s.NoError(err)
	s.Equal(201, statusCode)

	session.mustSetAllPermsExcept([]authr.PermissionID{authr.REMOVE_TAG})
	statusCode, err = session.removeTag(iconIn.Name, tag)
	s.NoError(err)
	s.Equal(403, statusCode)

	iconOut.Tags = []string{tag}

	respIcons := session.mustDescribeAllIcons()
	s.AssertResponseIconSetsEqual(dataOut, respIcons)
}

func (s *tagsTestSuite) TestDeletingSucceedssWithRequiredPermission() {
	dataIn, dataOut := testdata.Get()
	iconIn := dataIn[0]
	tag := "Ahoj"

	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)
	statusCode, err := session.addTag(iconIn.Name, tag)
	s.NoError(err)
	s.Equal(201, statusCode)

	session.mustSetAuthorization([]authr.PermissionID{authr.REMOVE_TAG})
	statusCode, err = session.removeTag(iconIn.Name, tag)
	s.NoError(err)
	s.Equal(204, statusCode)

	respIcons := session.mustDescribeAllIcons()
	s.AssertResponseIconSetsEqual(dataOut, respIcons)
}
