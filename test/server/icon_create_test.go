package server

import (
	"testing"

	"iconrepo/internal/app/domain"
	"iconrepo/internal/app/security/authn"
	"iconrepo/internal/app/security/authr"
	"iconrepo/internal/httpadapter"
	"iconrepo/test/testdata"

	"github.com/stretchr/testify/suite"
)

type iconCreateTestSuite struct {
	IconTestSuite
}

func TestIconCreateTestSuite(t *testing.T) {
	t.Parallel()
	for _, iconSuite := range IconTestSuites("api_iconcreate") {
		suite.Run(t, &iconCreateTestSuite{IconTestSuite: iconSuite})
	}
}

func (s *iconCreateTestSuite) TestFailsWith403WithoutPrivilege() {
	const iconName = "dock"
	var iconFile = domain.IconfileDescriptor{
		Format: "png",
		Size:   "36dp",
	}
	iconfileContent := testdata.GetDemoIconfileContent(iconName, iconFile)

	session := s.Client.mustLogin(nil)
	session.mustSetAllPermsExcept([]authr.PermissionID{authr.CREATE_ICON})
	statusCode, _, err := session.CreateIcon(iconName, iconfileContent)
	s.Error(err)
	s.ErrorIs(err, errJSONUnmarshal)
	s.Equal(403, statusCode)

	icons, errDesc := session.DescribeAllIcons(s.Ctx)
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
	expectedResponse := httpadapter.IconDTO{
		Name:       iconName,
		ModifiedBy: expectedUserID.String(),
		Tags:       []string{},
		Paths: []httpadapter.IconPath{
			{
				IconfileDescriptor: expectedIconfileDescriptor,
				Path:               getFilePath(iconName, expectedIconfileDescriptor),
			},
		},
	}

	session := s.Client.mustLogin(nil)
	session.mustSetAuthorization([]authr.PermissionID{authr.CREATE_ICON})
	statusCode, resultIcon, err := session.CreateIcon(iconName, iconfileContent)
	s.NoError(err)
	s.Equal(201, statusCode)
	s.Equal(expectedResponse, resultIcon)

	icons, errDesc := session.DescribeAllIcons(s.Ctx)
	s.NoError(errDesc)
	s.Equal(1, len(icons))
	s.Equal(expectedResponse, icons[0])

	iconfile, getIconfileError := session.GetIconfile(iconName, expectedIconfileDescriptor)
	s.NoError(getIconfileError)
	s.Equal(iconfileContent, iconfile)

	s.AssertEndState()
}

func (s *iconCreateTestSuite) TestAddMultipleIconsInARow() {
	testInput, testOutput := testdata.Get()
	sampleIconName1 := testInput[0].Name
	sampleIconfileDesc1 := testInput[0].Iconfiles[0]
	sampleIconName2 := testInput[1].Name
	sampleIconfileDesc2 := testInput[1].Iconfiles[1]

	session := s.Client.MustLoginSetAllPerms()

	session.MustAddTestData(testInput)
	s.getCheckIconfile(session, sampleIconName1, sampleIconfileDesc1)
	s.getCheckIconfile(session, sampleIconName2, sampleIconfileDesc2)

	iconDescriptors, describeError := session.DescribeAllIcons(s.Ctx)
	s.NoError(describeError)
	s.AssertResponseIconSetsEqual(testOutput, iconDescriptors)

	s.AssertEndState()
}
