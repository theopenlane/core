package graphapi

import (
	"encoding/json"

	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/enums"
)

// Permissions for the different types of permissions
const (
	editor  = "Editor"
	viewer  = "Viewer"
	blocked = "Blocked"
	creator = "Creator"
)

// EntObject is a struct that contains the id, displayID, and name of an object
type EntObject struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	DisplayID string `json:"display_id,omitempty"`
}

// getGroupPermissions returns a slice of GroupPermissions for the given object type and permission
func getGroupPermissions[T any](obj []T, objectType string, permission enums.Permission) (perms []*model.GroupPermissions) {
	for _, e := range obj {
		if permission != enums.Creator {
			eo, err := convertToEntObject(e)
			if err != nil {
				return nil
			}

			perms = append(perms, &model.GroupPermissions{
				ObjectType:  objectType,
				ID:          &eo.ID,
				Permissions: permission,
				DisplayID:   &eo.DisplayID,
				Name:        &eo.Name,
			})
		} else {
			// creator permissions are not specific to an object
			perms = append(perms, &model.GroupPermissions{
				ObjectType:  objectType,
				Permissions: permission,
			})
		}
	}

	return perms
}

// convertToEntObject converts an object to an EntObject to be used in the GroupPermissions
// to get the id, displayID, and name of the object
func convertToEntObject(obj any) (*EntObject, error) {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var entObject EntObject

	err = json.Unmarshal(jsonBytes, &entObject)
	if err != nil {
		return nil, err
	}

	return &entObject, nil
}
