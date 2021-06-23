package api

import (
	"encoding/base64"
	"testing"

	"github.com/pdkovacs/igo-repo/backend/itests/common"
	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authn"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authr"
	"github.com/stretchr/testify/suite"
)

type iconCreateTestSuite struct {
	iconTestSuite
}

func TestIconCreateTestSuite(t *testing.T) {
	suite.Run(t, &iconCreateTestSuite{})
}

func (s *iconCreateTestSuite) BeforeTest(suiteName string, testName string) {
	serverConfig := common.CloneConfig(s.defaultConfig)
	serverConfig.EnableBackdoors = true
	s.startTestServer(serverConfig)
}

func (s *iconCreateTestSuite) TestFailsWith403WithoutPrivilege() {
	const iconName = "dock"
	var iconFile = domain.IconfileDescriptor{
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

func (s *iconCreateTestSuite) TestCompletesWithPrivilege() {
	const iconName = "dock"
	var iconfileDescriptor = domain.IconfileDescriptor{
		Format: "png",
		Size:   "36dp",
	}
	iconfileContent := getDemoIconfileContent(iconName, iconfileDescriptor)

	expectedIconfile := createIconfile(iconfileDescriptor, nil)
	expectedIconfile.Size = "54px" // TODO: preserve size in DP
	expectedIconfile.Content = []byte(base64.StdEncoding.EncodeToString(iconfileContent))
	expectedModifier := authn.LocalDomain.CreateUserID(defaultCredentials.Username)
	expectedIcon := domain.Icon{
		IconAttributes: domain.IconAttributes{
			Name:       iconName,
			ModifiedBy: expectedModifier.String(),
			Tags:       []string{},
		},
		Iconfiles: []domain.Iconfile{
			expectedIconfile,
		},
	}

	expectedIconfileDescriptor := iconfileDescriptor
	expectedIconfileDescriptor.Size = dp2px[iconfileDescriptor.Size]

	expectedIconDescriptor := domain.IconDescriptor{
		IconAttributes: domain.IconAttributes{
			Name:       iconName,
			ModifiedBy: expectedModifier.String(),
			Tags:       []string{},
		},
		Iconfiles: []domain.IconfileDescriptor{
			expectedIconfileDescriptor,
		},
	}

	session := s.client.mustLogin(nil)
	session.mustSetAuthorization([]authr.PermissionID{authr.CREATE_ICON})
	statusCode, resultIcon, err := session.createIcon(iconName, iconfileContent)
	s.NoError(err)
	s.Equal(201, statusCode)
	s.Equal(expectedIcon, resultIcon)

	icons, errDesc := session.describeAllIcons()
	s.NoError(errDesc)
	s.Equal(1, len(icons))
	s.Equal(expectedIconDescriptor, icons[0])
}

func (s *iconCreateTestSuite) TestAddMultipleIconsInARow() {
	sampleIconName1 := testIconInputData[0].Name
	sampleIconfileDesc1 := testIconInputData[0].Iconfiles[0]
	sampleIconName2 := testIconInputData[1].Name
	sampleIconfileDesc2 := testIconInputData[1].Iconfiles[1]

	session := s.client.mustLogin(nil)
	session.mustSetAuthorization([]authr.PermissionID{authr.CREATE_ICON})

	mustAddTestData(&session, &testIconInputData)
	s.getCheckIconfile(session, sampleIconName1, sampleIconfileDesc1)
	s.getCheckIconfile(session, sampleIconName2, sampleIconfileDesc2)
	s.assertGitCleanStatus()

	iconDescriptors, describeError := session.describeAllIcons()
	s.NoError(describeError)
	s.Equal(ingestedTestIconDataDescription, iconDescriptors)
}
