package api_tests

import (
	"testing"

	"iconrepo/internal/app/security/authr"
	"iconrepo/test/testdata"

	"github.com/stretchr/testify/suite"
)

type iconUpdateTestSuite struct {
	IconTestSuite
}

func TestIconUpdateTestSuite(t *testing.T) {
	t.Parallel()
	for _, iconSuite := range IconTestSuites("api_iconupdate") {
		suite.Run(t, &iconUpdateTestSuite{IconTestSuite: iconSuite})
	}
}

func (s *iconUpdateTestSuite) TestAddingIconfileFailsWith403WithoutPermission() {
	dataIn, dataOut := testdata.Get()
	moreDataIn, _ := testdata.GetMore()

	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)

	session.mustSetAllPermsExcept([]authr.PermissionID{authr.UPDATE_ICON, authr.ADD_ICONFILE})

	iconName := moreDataIn[0].Name

	statusCode, _, updateError := session.addIconfile(iconName, moreDataIn[0].Iconfiles[1])

	s.Error(updateError)
	s.ErrorIs(updateError, errJSONUnmarshal)
	s.Equal(403, statusCode)

	resp, descError := session.DescribeAllIcons(s.Ctx)
	s.NoError(descError)
	s.AssertResponseIconSetsEqual(dataOut, resp)

	s.AssertEndState()
}

func (s *iconUpdateTestSuite) TestCanAddIconfilesWithProperPermission() {
	dataIn, dataOut := testdata.Get()
	moreDataIn, _ := testdata.GetMore()

	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)

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

	iconDesc, descError := session.DescribeAllIcons(s.Ctx)
	s.NoError(descError)

	s.AssertResponseIconSetsEqual(dataOut, iconDesc)

	s.AssertEndState()
}

func (s *iconUpdateTestSuite) TestAddingIconfilesWithExistingFormatSizeComboToFail() {
	dataIn, dataOut := testdata.Get()

	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)

	session.setAuthorization([]authr.PermissionID{authr.UPDATE_ICON, authr.ADD_ICONFILE})

	iconName := dataIn[0].Name
	newIconfile := dataIn[0].Iconfiles[1]

	statusCode, _, updateError := session.addIconfile(iconName, newIconfile)

	s.Error(updateError)
	s.ErrorIs(updateError, errJSONUnmarshal)
	s.Equal(409, statusCode)

	iconDesc, descError := session.DescribeAllIcons(s.Ctx)
	s.NoError(descError)

	s.AssertResponseIconSetsEqual(dataOut, iconDesc)

	s.AssertEndState()
}
