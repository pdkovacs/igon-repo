package services

import (
	"iconrepo/internal/app/security/authn"
	"iconrepo/internal/app/security/authr"
)

func NewUserService(authorizationService AuthorizationService) UserService {
	return UserService{
		AuthorizationService: authorizationService,
	}
}

type UserService struct {
	AuthorizationService AuthorizationService
}

func (us *UserService) getPermissionsForUser(userId authn.UserID) []authr.PermissionID {
	userPermissions := []authr.PermissionID{}

	memberIn := us.AuthorizationService.GetGroupsForUser(userId)
	for _, group := range memberIn {
		userPermissions = append(userPermissions, authr.GetPermissionsForGroup(group)...)
	}

	return userPermissions
}

func (us *UserService) getDisplayName(userId authn.UserID) string {
	return userId.String()
}

func (us *UserService) GetUserInfo(userId authn.UserID) authr.UserInfo {
	memberIn := us.AuthorizationService.GetGroupsForUser(userId)
	return authr.UserInfo{
		UserId:      userId,
		Groups:      memberIn,
		Permissions: us.getPermissionsForUser(userId),
		DisplayName: us.getDisplayName(userId),
	}
}

func (us *UserService) UpdateUserInfo(userId authn.UserID, memberIn []authr.GroupID) {
	us.AuthorizationService.UpdateUser(userId, memberIn)
}
