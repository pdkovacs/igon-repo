package api

import (
	"testing"

	"github.com/pdkovacs/igo-repo/backend/itests/common"
	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authn"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authr"
	"github.com/stretchr/testify/suite"
)

type iconCreateTestSuite struct {
	apiTestSuite
}

func TestIconCreateTestSuite(t *testing.T) {
	suite.Run(t, &iconCreateTestSuite{})
}

func (s *iconCreateTestSuite) BeforeTest(suiteName string, testName string) {
	serverConfig := common.CloneConfig(s.defaultConfig)
	serverConfig.EnableBackdoors = true
	s.startTestServer(serverConfig)
}

func (s *iconCreateTestSuite) TestPOSTShouldFailWith403WithoutCREATE_ICONPrivilegeTest() {
	const iconName = "dock"
	var iconFile = domain.Iconfile{
		Format: "png",
		Size:   "36dp",
	}
	iconfileContent := getDemoIconfileContent(iconName, iconFile)

	session := s.client.mustLogin(nil)
	session.mustSetAuthorization([]authr.PermissionID{})
	statusCode, _, err := session.createIcon(iconName, iconfileContent)
	s.NoError(err)
	s.Equal(403, statusCode)

	icons, errDesc := session.describeAllIcons()
	s.NoError(errDesc)
	s.Equal(0, len(icons))
}

func (s *iconCreateTestSuite) TestPOSTShouldCompleteWithCREATE_ICONPrivilegeTest() {
	const iconName = "dock"
	var iconFile = domain.Iconfile{
		Format: "png",
		Size:   "36dp",
	}
	iconfileContent := getDemoIconfileContent(iconName, iconFile)

	expectedIconfile := iconFile
	expectedIconfile.Size = "54px" // TODO: preserve size in DP
	expectedModifier := authn.LocalDomain.CreateUserID(defaultCredentials.Username)
	expectedResult := domain.Icon{
		Name:       iconName,
		ModifiedBy: expectedModifier.String(),
		Iconfiles: []domain.Iconfile{
			expectedIconfile,
		},
		Tags: []string{},
	}

	session := s.client.mustLogin(nil)
	session.mustSetAuthorization([]authr.PermissionID{authr.CREATE_ICON})
	statusCode, resultIcon, err := session.createIcon(iconName, iconfileContent)
	s.NoError(err)
	s.Equal(201, statusCode)
	s.Equal(expectedResult, resultIcon)

	icons, errDesc := session.describeAllIcons()
	s.NoError(errDesc)
	s.Equal(1, len(icons))
	s.Equal(expectedResult, icons[0])
}
