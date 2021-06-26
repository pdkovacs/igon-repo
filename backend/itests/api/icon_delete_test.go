package api

import (
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/security/authr"
	"github.com/pdkovacs/igo-repo/backend/pkg/web"
	"github.com/stretchr/testify/suite"
)

type deleteIconTestSuite struct {
	iconTestSuite
}

func TestDeleteIconTestSuite(t *testing.T) {
	suite.Run(t, &deleteIconTestSuite{})
}

func (s *deleteIconTestSuite) TestFailWith403WithoutPrivilege() {
	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(testIconInputData)
	session.setAuthorization([]authr.PermissionID{})
	statusCode, deleteError := session.deleteIcon(testIconInputData[0].Name)
	s.NoError(deleteError)
	s.Equal(403, statusCode)
	respIcons, listError := session.describeAllIcons()
	s.NoError((listError))
	s.Equal(testIconDataResponse, respIcons)

	s.assertEndState()
}

func (s *deleteIconTestSuite) TestSucceedsWithPrivilege() {
	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(testIconInputData)
	session.setAuthorization([]authr.PermissionID{authr.REMOVE_ICON})
	statusCode, deleteError := session.deleteIcon(testIconInputData[0].Name)
	s.NoError(deleteError)
	s.Equal(204, statusCode)
	respIcons, listError := session.describeAllIcons()
	s.NoError((listError))
	s.Equal([]web.ResponseIcon{testIconDataResponse[1]}, respIcons)

	s.assertEndState()
}
