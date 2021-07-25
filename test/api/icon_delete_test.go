package api

import (
	"testing"

	"github.com/pdkovacs/igo-repo/api"
	"github.com/pdkovacs/igo-repo/app/security/authr"
	"github.com/pdkovacs/igo-repo/test/api/testdata"
	"github.com/stretchr/testify/suite"
)

type deleteIconTestSuite struct {
	iconTestSuite
}

func TestDeleteIconTestSuite(t *testing.T) {
	suite.Run(t, &deleteIconTestSuite{})
}

func (s *deleteIconTestSuite) TestFailWith403WithoutPrivilege() {
	dataIn, dataOut := testdata.Get()
	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(dataIn)
	session.setAuthorization([]authr.PermissionID{})
	statusCode, deleteError := session.deleteIcon(dataIn[0].Name)
	s.NoError(deleteError)
	s.Equal(403, statusCode)
	respIcons, listError := session.describeAllIcons()
	s.NoError((listError))
	s.assertResponseIconSetsEqual(dataOut, respIcons)

	s.assertEndState()
}

func (s *deleteIconTestSuite) TestSucceedsWithPrivilege() {
	dataIn, dataOut := testdata.Get()
	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(dataIn)
	session.setAuthorization([]authr.PermissionID{authr.REMOVE_ICON})
	statusCode, deleteError := session.deleteIcon(dataIn[0].Name)
	s.NoError(deleteError)
	s.Equal(204, statusCode)
	respIcons, listError := session.describeAllIcons()
	s.NoError((listError))
	s.assertResponseIconSetsEqual([]api.ResponseIcon{dataOut[1]}, respIcons)

	s.assertEndState()
}
