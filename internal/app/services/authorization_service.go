package services

import (
	"igo-repo/internal/app/security/authn"
	"igo-repo/internal/app/security/authr"
	"igo-repo/internal/config"

	log "github.com/sirupsen/logrus"
)

type UsersByGroups map[authr.GroupID][]authn.UserID

type AuthorizationService interface {
	GetGroupsForUser(userID authn.UserID) []authr.GroupID
	GetPermissionsForGroup(group authr.GroupID) []authr.PermissionID
}

func NewAuthorizationService(config config.Options) authRService {
	return authRService{config.UsersByRoles}
}

type authRService struct {
	// TODO: if this data structure is to serve both the local and the OIDC domain,
	// "usersByGroups" should be more abstract/indirect, like let it be at least a function or something
	usersByGroups config.UsersByRoles
}

func (as *authRService) GetGroupsForUser(userID authn.UserID) []authr.GroupID {
	return getLocalGroupsFor(userID, as.usersByGroups)
}

func (as *authRService) GetPermissionsForGroup(group authr.GroupID) []authr.PermissionID {
	return authr.GetPermissionsForGroup(group)
}

func str2GroupID(s string) authr.GroupID {
	return authr.GroupID(s)
}

func getLocalGroupsFor(userID authn.UserID, usersByGroups map[string][]string) []authr.GroupID {
	groupNames := []string{}
	for groupName, members := range usersByGroups {
		for _, member := range members {
			if userID.IDInDomain == member {
				groupNames = append(groupNames, groupName)
				break
			}
		}
	}
	log.Debugf("Groups of %s: %v", userID, groupNames)
	return groupNames2GroupIDs(groupNames)
}

func groupNames2GroupIDs(strs []string) []authr.GroupID {
	groupIDs := []authr.GroupID{}
	for _, s := range strs {
		groupIDs = append(groupIDs, str2GroupID((s)))
	}
	return groupIDs
}
