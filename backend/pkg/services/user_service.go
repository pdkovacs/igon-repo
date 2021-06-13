package services

import (
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authn"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authr"
)

type UserInfo struct {
	UserId      authn.UserID
	Groups      []authr.GroupID
	Permissions []authr.PermissionID
	DisplayName string
}

func NewUserService(authorizationService AuthorizationService) UserService {

	return UserService{
		authorizationService: authorizationService,
	}
}

type UserService struct {
	authorizationService AuthorizationService
}

func (us *UserService) getPermissionsForUser(userId authn.UserID, memberIn []authr.GroupID) []authr.PermissionID {
	userPermissions := []authr.PermissionID{}

	for _, group := range memberIn {
		userPermissions = append(userPermissions, authr.GetPermissionsByGroup()[group]...)
	}

	return userPermissions
}

func (us *UserService) getDisplayName(userId authn.UserID) string {
	return userId.String()
}

func (us *UserService) GetUserInfo(userId authn.UserID) UserInfo {
	memberIn := us.authorizationService.GetGroupsForUser(userId)
	return UserInfo{
		UserId:      userId,
		Groups:      memberIn,
		Permissions: us.getPermissionsForUser(userId, memberIn),
		DisplayName: us.getDisplayName(userId),
	}
}
