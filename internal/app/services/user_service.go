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

func (us *UserService) GetUserInfo(userId authn.UserID) authr.UserInfo {
	memberIn := us.authorizationService.GetGroupsForUser(userId)
	return authr.UserInfo{
		UserId:      userId,
		Groups:      memberIn,
		Permissions: us.getPermissionsForUser(userId, memberIn),
		DisplayName: us.getDisplayName(userId),
	}
}
