package app_tests

import (
	"errors"
	"testing"

	_ "image/jpeg"
	_ "image/png"

	"igo-repo/internal/app"
	"igo-repo/internal/app/domain"
	"igo-repo/internal/app/security/authn"
	"igo-repo/internal/app/security/authr"
	"igo-repo/internal/app/services"
	"igo-repo/internal/logging"
	"igo-repo/test/mocks"
	"igo-repo/test/testdata"

	"github.com/stretchr/testify/suite"
)

var appTestLogger = logging.CreateRootLogger(logging.DebugLevel)
var appTestApiLogger = logging.CreateUnitLogger(appTestLogger, "app-test-api")

func createUserInfo(withPerms []authr.PermissionID) authr.UserInfo {
	return authr.UserInfo{
		UserId:      authn.UserID{IDInDomain: "testuser"},
		Groups:      []authr.GroupID{},
		Permissions: withPerms,
	}
}

func getTestIconfile() domain.Iconfile {
	const iconName = "dock"
	var iconfileDescriptor = domain.IconfileDescriptor{
		Format: "png",
		Size:   "36dp",
	}
	content := testdata.GetDemoIconfileContent(iconName, iconfileDescriptor)
	return domain.Iconfile{
		IconfileDescriptor: domain.IconfileDescriptor{
			Format: iconfileDescriptor.Format,
			Size:   testdata.DP2PX[iconfileDescriptor.Size],
		},
		Content: content,
	}
}

type appTestSuite struct {
	suite.Suite
	t *testing.T
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, &appTestSuite{t: t})
}

func (s *appTestSuite) SetupSuite() {
	services.RegisterSVGDecoder(logging.CreateUnitLogger(appTestLogger, "app-test-suite"))
}

func (s *appTestSuite) TestCreateIconNoPerm() {
	testUser := createUserInfo(nil)
	iconName := "test-icon"
	iconfile := getTestIconfile()
	mockRepo := mocks.Repository{}
	application := app.AppCore{Repository: &mockRepo}
	api := application.GetAPI(appTestApiLogger)
	_, err := api.IconService.CreateIcon(iconName, iconfile.Content, testUser)
	s.True(errors.Is(err, authr.ErrPermission))
	mockRepo.AssertExpectations(s.t)
}

func (s *appTestSuite) TestCreateIcon() {
	testUser := createUserInfo([]authr.PermissionID{authr.CREATE_ICON})
	iconName := "test-icon"
	iconfile := getTestIconfile()
	expectedResponseIcon := domain.Icon{
		IconAttributes: domain.IconAttributes{
			Name:       iconName,
			ModifiedBy: testUser.UserId.IDInDomain,
			Tags:       []string{},
		},
		Iconfiles: []domain.Iconfile{iconfile},
	}
	mockRepo := mocks.Repository{}
	mockRepo.On("CreateIcon", iconName, iconfile, testUser).Return(nil)
	application := app.AppCore{Repository: &mockRepo}
	api := application.GetAPI(appTestApiLogger)
	icon, err := api.IconService.CreateIcon(iconName, iconfile.Content, testUser)
	s.NoError(err)
	s.Equal(expectedResponseIcon, icon)
	mockRepo.AssertExpectations(s.t)
}
