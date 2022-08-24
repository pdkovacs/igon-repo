package services

import (
	"igo-repo/internal/app/security/authn"
	"igo-repo/internal/app/security/authr"
)

func NewUserService(authorizationService AuthorizationService) UserService {
	return UserService{
		authorizationService: authorizationService,
	}
}

type UserService struct {
	authorizationService AuthorizationService
}

func (us *UserService) getPermissionsForUser(userId authn.UserID) []authr.PermissionID {
	userPermissions := []authr.PermissionID{}

	memberIn := us.authorizationService.GetGroupsForUser(userId)
	for _, group := range memberIn {
		userPermissions = append(userPermissions, authr.GetPermissionsForGroup(group)...)
	}

	return userPermissions
}

func (us *UserService) getDisplayName(userId authn.UserID) string {
	return userId.String()
}

func (us *UserService) GetUserInfo(userId authn.UserID) authr.UserInfo {
	memberIn := us.authorizationService.GetGroupsForUser(userId)
	return authr.UserInfo{
		UserId:      userId,
		Groups:      memberIn,
		Permissions: us.getPermissionsForUser(userId),
		DisplayName: us.getDisplayName(userId),
	}
}

func (us *UserService) UpdateUserInfo(userId authn.UserID, memberIn []authr.GroupID) {
	us.authorizationService.UpdateUser(userId, memberIn)
}
