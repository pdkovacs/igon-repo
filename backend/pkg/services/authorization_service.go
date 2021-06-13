package services

import (
	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authn"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authr"
)

type UsersByGroups map[authr.GroupID][]authn.UserID

type AuthorizationService interface {
	GetGroupsForUser(userId authn.UserID) []authr.GroupID
	GetPermissionsByGroup() map[authr.GroupID][]authr.PermissionID
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

func (as *authRService) GetPermissionsByGroup() map[authr.GroupID][]authr.PermissionID {
	return authr.GetPermissionsByGroup()
}
