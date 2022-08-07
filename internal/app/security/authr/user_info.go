package authr

import "igo-repo/internal/app/security/authn"

type UserInfo struct {
	UserId      authn.UserID   `json:"userID"`
	Groups      []GroupID      `json:"groups"`
	Permissions []PermissionID `json:"permissions"`
	DisplayName string         `json:"displayName"`
}
