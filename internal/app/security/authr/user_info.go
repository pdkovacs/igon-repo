package authr

import "igo-repo/internal/app/security/authn"

func GroupNamesToGroupIDs(groups []string) []GroupID {
	memberIn := []GroupID{}
	for _, group := range groups {
		memberIn = append(memberIn, GroupID(group))
	}
	return memberIn
}

type UserInfo struct {
	UserId      authn.UserID   `json:"userID"`
	Groups      []GroupID      `json:"groups"`
	Permissions []PermissionID `json:"permissions"`
	DisplayName string         `json:"displayName"`
}
