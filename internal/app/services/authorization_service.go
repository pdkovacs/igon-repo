package services

import (
	"iconrepo/internal/app/security/authn"
	"iconrepo/internal/app/security/authr"
	"iconrepo/internal/config"
	"strings"

	"github.com/rs/zerolog/log"
)

type UsersByGroups map[authr.GroupID][]authn.UserID

type AuthorizationService interface {
	GetGroupsForUser(userID authn.UserID) []authr.GroupID
	GetPermissionsForGroup(group authr.GroupID) []authr.PermissionID
	GetPermissionsForGroups(group []authr.GroupID) []authr.PermissionID
	UpdateUser(userId authn.UserID, groups []authr.GroupID)
}

func NewAuthorizationService(options config.Options) authRService {
	return authRService{options.UsersByRoles}
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

func (as *authRService) GetPermissionsForGroups(groups []authr.GroupID) []authr.PermissionID {
	permissions := []authr.PermissionID{}
	for _, group := range groups {
		permissions = append(permissions, authr.GetPermissionsForGroup(group)...)
	}
	return permissions
}

func (as *authRService) UpdateUser(userId authn.UserID, groups []authr.GroupID) {
	if as.usersByGroups == nil {
		as.usersByGroups = make(map[string][]string)
	}
	for _, group := range groups {
		as.usersByGroups[string(group)] = append(as.usersByGroups[string(group)], userId.IDInDomain)
	}
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
	log.Debug().Str("user_id", userID.String()).Str("group_names", strings.Join(groupNames, ", ")).Msg("user's group memberships collected")
	return authr.GroupNamesToGroupIDs(groupNames)
}
