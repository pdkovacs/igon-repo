package api

import (
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/stretchr/testify/suite"
)

type iconCreateTestSuite struct {
	apiTestSuite
}

func TestIconCreateTestSuite(t *testing.T) {
	suite.Run(t, &iconCreateTestSuite{})
}

func (s *iconCreateTestSuite) TestPOSTShouldFailWith403WithoutCREATE_ICONPrivilegeTest() {
	const iconName = "dock"
	var iconFile = domain.Iconfile{
		Format: "png",
		Size:   "36dp",
	}
	iconfileContent := getDemoIconfileContent(iconName, iconFile)

	statusCode, _, err := s.client.createIcon(iconName, iconfileContent, []string{})
	s.NoError(err)
	s.Equal(403, statusCode)
}
