package authr

import (
	"errors"
	"fmt"

	"github.com/pdkovacs/igo-repo/backend/pkg/security/authn"
)

type PermissionID string

const (
	CREATE_ICON     PermissionID = "CREATE_ICON"
	UPDATE_ICON     PermissionID = "UPDATE_ICON"
	ADD_ICONFILE    PermissionID = "ADD_ICONFILE"
	REMOVE_ICONFILE PermissionID = "REMOVE_ICONFILE"
	REMOVE_ICON     PermissionID = "REMOVE_ICON"
	ADD_TAG         PermissionID = "ADD_TAG"
	REMOVE_TAG      PermissionID = "REMOVE_TAG"
)

var permissionDictionary = map[PermissionID]string{
	CREATE_ICON:     "CREATE_ICON",
	UPDATE_ICON:     "UPDATE_ICON",
	ADD_ICONFILE:    "ADD_ICONFILE",
	REMOVE_ICONFILE: "REMOVE_ICONFILE",
	REMOVE_ICON:     "REMOVE_ICON",
	ADD_TAG:         "ADD_TAG",
	REMOVE_TAG:      "REMOVE_TAG",
}

func GetPrivilegeString(id PermissionID) string {
	return permissionDictionary[id]
}

type GroupID string

const (
	ICON_EDITOR GroupID = "ICON_EDITOR"
)

var permissionsByGroup = map[GroupID][]PermissionID{
	ICON_EDITOR: {
		CREATE_ICON,
		UPDATE_ICON,
		REMOVE_ICONFILE,
		REMOVE_ICON,
		ADD_TAG,
		REMOVE_TAG,
	},
}

func GetPermissionsByGroup() map[GroupID][]PermissionID {
	return permissionsByGroup
}

var ErrPermission = errors.New("permission error")

func HasRequiredPermissions(userId authn.UserID, userPermissions []PermissionID, requiredPermissions []PermissionID) error {
	for _, reqPerm := range requiredPermissions {
		found := false
		for _, uPerm := range userPermissions {
			if reqPerm == uPerm {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("not all of %v is included in %v granted to %v, %w", requiredPermissions, userPermissions, userId, ErrPermission)
		}
	}
	return nil
}
