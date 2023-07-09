package api_tests

import (
	"testing"

	"iconrepo/internal/app/security/authr"
	"iconrepo/internal/httpadapter"
	"iconrepo/test/testdata"

	"github.com/stretchr/testify/suite"
)

type deleteIconTestSuite struct {
	IconTestSuite
}

func TestDeleteIconTestSuite(t *testing.T) {
	t.Parallel()
	for _, iconSuite := range IconTestSuites("api_deleteicon") {
		suite.Run(t, &deleteIconTestSuite{IconTestSuite: iconSuite})
	}
}

func (s *deleteIconTestSuite) TestFailWith403WithoutPrivilege() {
	dataIn, dataOut := testdata.Get()
	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)
	session.setAuthorization([]authr.PermissionID{})
	statusCode, deleteError := session.deleteIcon(dataIn[0].Name)
	s.NoError(deleteError)
	s.Equal(403, statusCode)
	respIcons, listError := session.DescribeAllIcons()
	s.NoError((listError))
	s.AssertResponseIconSetsEqual(dataOut, respIcons)

	s.AssertEndState()
}

func (s *deleteIconTestSuite) TestSucceedsWithPrivilege() {
	dataIn, dataOut := testdata.Get()
	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)
	session.setAuthorization([]authr.PermissionID{authr.REMOVE_ICON})
	statusCode, deleteError := session.deleteIcon(dataIn[0].Name)
	s.NoError(deleteError)
	s.Equal(204, statusCode)
	respIcons, listError := session.DescribeAllIcons()
	s.NoError((listError))
	s.AssertResponseIconSetsEqual([]httpadapter.IconDTO{dataOut[1]}, respIcons)

	s.AssertEndState()
}
