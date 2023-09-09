package iconservice

import (
	"context"
	"testing"

	_ "image/jpeg"
	_ "image/png"

	"iconrepo/internal/app/domain"
	"iconrepo/internal/app/security/authn"
	"iconrepo/internal/app/security/authr"
	"iconrepo/internal/app/services"
	"iconrepo/test/mocks"
	"iconrepo/test/testdata"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

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
	t   *testing.T
	ctx context.Context
}

func TestIconServiceTestSuite(t *testing.T) {
	suite.Run(t, &appTestSuite{t: t, ctx: context.TODO()})
}

func (s *appTestSuite) SetupSuite() {
	services.RegisterSVGDecoder()
}

func (s *appTestSuite) TestCreateIconNoPerm() {
	testUser := createUserInfo(nil)
	iconName := "test-icon"
	iconfile := getTestIconfile()
	mockRepo := mocks.Repository{}
	api := services.NewIconService(&mockRepo)
	_, err := api.CreateIcon(s.ctx, iconName, iconfile.Content, testUser)
	s.Error(err)
	s.ErrorIs(err, authr.ErrPermission)
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
	// mockRepo.On("CreateIcon", mock.AnythingOfType("*context.emptyCtx"), iconName, iconfile, testUser).Return(nil)
	mockRepo.EXPECT().CreateIcon(mock.AnythingOfType("*context.emptyCtx"), iconName, iconfile, testUser).Return(nil)
	api := services.NewIconService(&mockRepo)
	icon, err := api.CreateIcon(s.ctx, iconName, iconfile.Content, testUser)
	s.NoError(err)
	s.Equal(expectedResponseIcon, icon)
	mockRepo.AssertExpectations(s.t)
}
