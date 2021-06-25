package api

import (
	"errors"
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/security/authr"
	"github.com/stretchr/testify/suite"
)

type iconUpdateTestSuite struct {
	iconTestSuite
}

func TestIconUpdateTestSuite(t *testing.T) {
	suite.Run(t, &iconUpdateTestSuite{})
}

func (s *iconTestSuite) TestAddingIconfileFailsWith403WithoutPermission() {
	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(testIconInputData)

	session.mustSetAllPermsExcept([]authr.PermissionID{authr.UPDATE_ICON, authr.ADD_ICONFILE})

	iconName := testIconInputData[0].Name

	statusCode, _, updateError := session.addIconfile(iconName, moreIconInputData[0].Iconfiles[1])

	s.True(errors.Is(updateError, errJSONUnmarshal))
	s.Equal(403, statusCode)

	statusCode, resp, descError := session.describeIcon(iconName)
	s.Equal(200, statusCode)
	s.NoError(descError)
	s.Equal(testIconDataResponse[0], resp)
}

func (s *iconTestSuite) TestCanAddIconfilesWithProperPermission() {
	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(testIconInputData)

	session.setAuthorization([]authr.PermissionID{authr.UPDATE_ICON, authr.ADD_ICONFILE})

	iconName := testIconInputData[0].Name

	statusCode, resp, updateError := session.addIconfile(iconName, moreIconInputData[0].Iconfiles[1])

	s.NoError(updateError)
	s.Equal(200, statusCode)
	s.Equal(moreTestIconDataResponse[0].Paths[1], resp)

	expectedIconDesc := testIconDataResponse[0]
	expectedIconDesc.Paths = append(expectedIconDesc.Paths, moreTestIconDataResponse[0].Paths[1])

	descStatus, iconDesc, descError := session.describeIcon(iconName)
	s.Equal(200, descStatus)
	s.NoError(descError)

	s.Equal(expectedIconDesc, iconDesc)
}

func (s *iconTestSuite) TestAddingIconfilesWithExistingFormatSizeComboToFail() {
	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(testIconInputData)

	session.setAuthorization([]authr.PermissionID{authr.UPDATE_ICON, authr.ADD_ICONFILE})

	iconName := testIconInputData[0].Name

	statusCode, _, updateError := session.addIconfile(iconName, moreIconInputData[0].Iconfiles[0])

	s.True(errors.Is(updateError, errJSONUnmarshal))
	s.Equal(409, statusCode)

	descStatus, iconDesc, descError := session.describeIcon(iconName)
	s.Equal(200, descStatus)
	s.NoError(descError)
	s.Equal(testIconDataResponse[0], iconDesc)
}
