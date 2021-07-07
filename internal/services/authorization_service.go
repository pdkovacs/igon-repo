package services

import (
	"github.com/pdkovacs/igo-repo/internal/auxiliaries"
	"github.com/pdkovacs/igo-repo/internal/security/authn"
	"github.com/pdkovacs/igo-repo/internal/security/authr"
)

type UsersByGroups map[authr.GroupID][]authn.UserID

type AuthorizationService interface {
	GetGroupsForUser(userId authn.UserID) []authr.GroupID
	GetPermissionsForGroup(group authr.GroupID) []authr.PermissionID
}

func NewAuthorizationService(config auxiliaries.Options) authRService {
	return authRService{config.UsersByRoles}
}

type authRService struct {
	usersByGroups auxiliaries.UsersByRoles
}

func (as *authRService) GetGroupsForUser(userId authn.UserID) []authr.GroupID {
	return nil
}

func (as *authRService) GetPermissionsForGroup(group authr.GroupID) []authr.PermissionID {
	return authr.GetPermissionsForGroup(group)
}
