package api

import (
	"errors"
	"testing"

	"github.com/pdkovacs/igo-repo/app/security/authr"
	"github.com/pdkovacs/igo-repo/test/api/testdata"
	"github.com/stretchr/testify/suite"
)

type iconUpdateTestSuite struct {
	iconTestSuite
}

func TestIconUpdateTestSuite(t *testing.T) {
	suite.Run(t, &iconUpdateTestSuite{})
}

func (s *iconTestSuite) TestAddingIconfileFailsWith403WithoutPermission() {
	dataIn, dataOut := testdata.Get()
	moreDataIn, _ := testdata.GetMore()

	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(dataIn)

	session.mustSetAllPermsExcept([]authr.PermissionID{authr.UPDATE_ICON, authr.ADD_ICONFILE})

	iconName := moreDataIn[0].Name

	statusCode, _, updateError := session.addIconfile(iconName, moreDataIn[0].Iconfiles[1])

	s.True(errors.Is(updateError, errJSONUnmarshal))
	s.Equal(403, statusCode)

	resp, descError := session.describeAllIcons()
	s.NoError(descError)
	s.assertResponseIconSetsEqual(dataOut, resp)

	s.assertEndState()
}

func (s *iconTestSuite) TestCanAddIconfilesWithProperPermission() {
	dataIn, dataOut := testdata.Get()
	moreDataIn, _ := testdata.GetMore()

	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(dataIn)

	session.setAuthorization([]authr.PermissionID{authr.UPDATE_ICON, authr.ADD_ICONFILE})

	iconName := dataIn[0].Name
	newIconfile := moreDataIn[0].Iconfiles[1]
	iconfilePath := s.createIconfilePaths(iconName, moreDataIn[0].Iconfiles[1].IconfileDescriptor)

	statusCode, resp, updateError := session.addIconfile(iconName, newIconfile)

	s.NoError(updateError)
	s.Equal(200, statusCode)
	s.Equal(iconfilePath, resp)

	expectedIconDesc := dataOut
	expectedIconDesc[0].Paths = append(expectedIconDesc[0].Paths, iconfilePath)

	iconDesc, descError := session.describeAllIcons()
	s.NoError(descError)

	s.assertResponseIconSetsEqual(dataOut, iconDesc)

	s.assertEndState()
}

func (s *iconTestSuite) TestAddingIconfilesWithExistingFormatSizeComboToFail() {
	dataIn, dataOut := testdata.Get()

	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(dataIn)

	session.setAuthorization([]authr.PermissionID{authr.UPDATE_ICON, authr.ADD_ICONFILE})

	iconName := dataIn[0].Name
	newIconfile := dataIn[0].Iconfiles[1]

	statusCode, _, updateError := session.addIconfile(iconName, newIconfile)

	s.True(errors.Is(updateError, errJSONUnmarshal))
	s.Equal(409, statusCode)

	iconDesc, descError := session.describeAllIcons()
	s.NoError(descError)

	s.assertResponseIconSetsEqual(dataOut, iconDesc)

	s.assertEndState()
}
