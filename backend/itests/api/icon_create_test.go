package api

import (
	"testing"

	"github.com/pdkovacs/igo-repo/backend/itests/common"
	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
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

	cjar := s.client.MustCreateCookieJar()
	s.client.mustSetAuthorization(cjar, []authr.PermissionID{})
	statusCode, _, err := s.client.createIcon(cjar, iconName, iconfileContent)
	s.NoError(err)
	s.Equal(403, statusCode)
}

func (s *iconCreateTestSuite) TestPOSTShouldCompleteWithCREATE_ICONPrivilegeTest() {
	const iconName = "dock"
	var iconFile = domain.Iconfile{
		Format: "png",
		Size:   "36dp",
	}
	iconfileContent := getDemoIconfileContent(iconName, iconFile)

	cjar := s.client.MustCreateCookieJar()
	s.client.mustSetAuthorization(cjar, []authr.PermissionID{authr.CREATE_ICON})
	statusCode, _, err := s.client.createIcon(cjar, iconName, iconfileContent)
	s.NoError(err)
	s.Equal(201, statusCode)
}
