package api

import (
	"os"
	"testing"

	test_repositories "github.com/pdkovacs/igo-repo/backend/itests/repositories"
	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/repositories"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authn"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authr"
	"github.com/pdkovacs/igo-repo/backend/pkg/web"
	"github.com/stretchr/testify/suite"
)

type iconCreateTestSuite struct {
	iconTestSuite
}

func TestIconCreateTestSuite(t *testing.T) {
	suite.Run(t, &iconCreateTestSuite{})
}

func (s *iconCreateTestSuite) TestFailsWith403WithoutPrivilege() {
	const iconName = "dock"
	var iconFile = domain.IconfileDescriptor{
		Format: "png",
		Size:   "36dp",
	}
	iconfileContent := getDemoIconfileContent(iconName, iconFile)

	session := s.client.mustLogin(nil)
	session.mustSetAllPermsExcept([]authr.PermissionID{authr.CREATE_ICON})
	statusCode, _, err := session.createIcon(iconName, iconfileContent)
	s.NoError(err)
	s.Equal(403, statusCode)

	icons, errDesc := session.describeAllIcons()
	s.NoError(errDesc)
	s.Equal(0, len(icons))
}

func (s *iconCreateTestSuite) TestCompletesWithPrivilege() {
	const iconName = "dock"
	var iconfileDescriptor = domain.IconfileDescriptor{
		Format: "png",
		Size:   "36dp",
	}
	iconfileContent := getDemoIconfileContent(iconName, iconfileDescriptor)

	expectedUserID := authn.LocalDomain.CreateUserID(defaultCredentials.Username)
	expectedIconfileDescriptor := domain.IconfileDescriptor{
		Format: iconfileDescriptor.Format,
		Size:   dp2px[iconfileDescriptor.Size],
	}
	expectedResponse := web.ResponseIcon{
		Name:       iconName,
		ModifiedBy: expectedUserID.String(),
		Tags:       []string{},
		Paths: []web.IconPath{
			{
				IconfileDescriptor: expectedIconfileDescriptor,
				Path:               getFilePath(iconName, expectedIconfileDescriptor),
			},
		},
	}

	session := s.client.mustLogin(nil)
	session.mustSetAuthorization([]authr.PermissionID{authr.CREATE_ICON})
	statusCode, resultIcon, err := session.createIcon(iconName, iconfileContent)
	s.NoError(err)
	s.Equal(201, statusCode)
	s.Equal(expectedResponse, resultIcon)

	icons, errDesc := session.describeAllIcons()
	s.NoError(errDesc)
	s.Equal(1, len(icons))
	s.Equal(expectedResponse, icons[0])
}

func (s *iconCreateTestSuite) TestAddMultipleIconsInARow() {
	sampleIconName1 := testIconInputData[0].Name
	sampleIconfileDesc1 := testIconInputData[0].Iconfiles[0]
	sampleIconName2 := testIconInputData[1].Name
	sampleIconfileDesc2 := testIconInputData[1].Iconfiles[1]

	session := s.client.mustLoginSetAllPerms()

	session.mustAddTestData(testIconInputData)
	s.getCheckIconfile(session, sampleIconName1, sampleIconfileDesc1)
	s.getCheckIconfile(session, sampleIconName2, sampleIconfileDesc2)
	s.assertGitCleanStatus()

	iconDescriptors, describeError := session.describeAllIcons()
	s.NoError(describeError)
	s.Equal(testIconDataResponse, iconDescriptors)
}

func (s *iconCreateTestSuite) TestRollbackToLastConsistentStateOnError() {
	intactIcon := testIconInputData[0]

	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData([]domain.Icon{intactIcon})

	lastStableSHA1, beforeIncidentGitErr := test_repositories.GetCurrentCommit(s.server.Repositories.Git)
	s.NoError(beforeIncidentGitErr)

	os.Setenv(repositories.IntrusiveGitTestEnvvarName, "true")

	statusCode, _, _ := session.createIcon(testIconInputData[1].Name, testIconInputData[1].Iconfiles[0].Content)
	s.Equal(500, statusCode)

	test_repositories.AssertGitCleanStatus(&s.Suite, s.server.Repositories.Git)
	afterIncidentSHA1, afterIncidentGitErr := test_repositories.GetCurrentCommit(s.server.Repositories.Git)
	s.NoError(afterIncidentGitErr)

	s.Equal(lastStableSHA1, afterIncidentSHA1)

	iconDescriptors, describeError := session.describeAllIcons()
	s.NoError(describeError)
	s.Equal(1, len(iconDescriptors))
	s.Equal([]web.ResponseIcon{testIconDataResponse[0]}, iconDescriptors)
}
