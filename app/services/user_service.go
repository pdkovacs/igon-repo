package services

import (
	"github.com/pdkovacs/igo-repo/app/security/authn"
	"github.com/pdkovacs/igo-repo/app/security/authr"
)

type UserInfo struct {
	UserId      authn.UserID         `json:"userID"`
	Groups      []authr.GroupID      `json:"groups"`
	Permissions []authr.PermissionID `json:"permissions"`
	DisplayName string               `json:"displayName"`
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
		userPermissions = append(userPermissions, authr.GetPermissionsForGroup(group)...)
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
