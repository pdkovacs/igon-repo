package api

import (
	"errors"
	"os"
	"testing"

	"github.com/pdkovacs/igo-repo/api"
	"github.com/pdkovacs/igo-repo/domain"
	"github.com/pdkovacs/igo-repo/repositories"
	"github.com/pdkovacs/igo-repo/security/authn"
	"github.com/pdkovacs/igo-repo/security/authr"
	"github.com/pdkovacs/igo-repo/test/api/testdata"
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
	iconfileContent := testdata.GetDemoIconfileContent(iconName, iconFile)

	session := s.client.mustLogin(nil)
	session.mustSetAllPermsExcept([]authr.PermissionID{authr.CREATE_ICON})
	statusCode, _, err := session.createIcon(iconName, iconfileContent)
	s.True(errors.Is(err, errJSONUnmarshal))
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
	iconfileContent := testdata.GetDemoIconfileContent(iconName, iconfileDescriptor)

	expectedUserID := authn.LocalDomain.CreateUserID(testdata.DefaultCredentials.Username)
	expectedIconfileDescriptor := domain.IconfileDescriptor{
		Format: iconfileDescriptor.Format,
		Size:   testdata.DP2PX[iconfileDescriptor.Size],
	}
	expectedResponse := api.ResponseIcon{
		Name:       iconName,
		ModifiedBy: expectedUserID.String(),
		Tags:       []string{},
		Paths: []api.IconPath{
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

	s.assertEndState()
}

func (s *iconCreateTestSuite) TestAddMultipleIconsInARow() {
	testInput, testOutput := testdata.Get()
	sampleIconName1 := testInput[0].Name
	sampleIconfileDesc1 := testInput[0].Iconfiles[0]
	sampleIconName2 := testInput[1].Name
	sampleIconfileDesc2 := testInput[1].Iconfiles[1]

	session := s.client.mustLoginSetAllPerms()

	session.mustAddTestData(testInput)
	s.getCheckIconfile(session, sampleIconName1, sampleIconfileDesc1)
	s.getCheckIconfile(session, sampleIconName2, sampleIconfileDesc2)

	iconDescriptors, describeError := session.describeAllIcons()
	s.NoError(describeError)
	s.assertResponseIconSetsEqual(testOutput, iconDescriptors)

	s.assertEndState()
}

func (s *iconCreateTestSuite) TestRollbackToLastConsistentStateOnError() {
	dataIn, dataOut := testdata.Get()
	moreDataIn, _ := testdata.Get()

	session := s.client.mustLoginSetAllPerms()
	session.mustAddTestData(dataIn)

	lastStableSHA1, beforeIncidentGitErr := s.testGitRepo.GetCurrentCommit()
	s.NoError(beforeIncidentGitErr)

	os.Setenv(repositories.IntrusiveGitTestEnvvarName, "true")

	statusCode, _, _ := session.createIcon(moreDataIn[1].Name, moreDataIn[1].Iconfiles[0].Content)
	s.Equal(500, statusCode)

	afterIncidentSHA1, afterIncidentGitErr := s.testGitRepo.GetCurrentCommit()
	s.NoError(afterIncidentGitErr)

	s.Equal(lastStableSHA1, afterIncidentSHA1)

	iconDescriptors, describeError := session.describeAllIcons()
	s.NoError(describeError)
	s.assertResponseIconSetsEqual(dataOut, iconDescriptors)

	s.assertEndState()
}
