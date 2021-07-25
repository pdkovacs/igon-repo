package authr

import "github.com/pdkovacs/igo-repo/app/security/authn"

type UserInfo struct {
	UserId      authn.UserID   `json:"userID"`
	Groups      []GroupID      `json:"groups"`
	Permissions []PermissionID `json:"permissions"`
	DisplayName string         `json:"displayName"`
}
