package graphapi

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/enums"
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

func getGroupByIDWithPermissionsEdges(ctx context.Context, groupID *string) (*generated.Group, error) {
	if groupID != nil {
		groupWithPermissions, err := withTransactionalMutation(ctx).Group.
			Query().
			Where(group.IDEQ(*groupID)).
			// Control permissions
			WithControlEditors().
			WithControlViewers().
			WithControlBlockedGroups().

			// Control Objective permissions
			WithControlObjectiveEditors().
			WithControlObjectiveViewers().
			WithControlObjectiveBlockedGroups().

			// Program permissions
			WithProgramViewers().
			WithProgramEditors().
			WithProgramBlockedGroups().

			// Risk permissions
			WithRiskViewers().
			WithRiskEditors().
			WithRiskBlockedGroups().

			// Internal Policy permissions
			WithInternalPolicyEditors().
			WithInternalPolicyBlockedGroups().

			// Procedure permissions
			WithProcedureEditors().
			WithProcedureBlockedGroups().

			// Narrative permissions
			WithNarrativeViewers().
			WithNarrativeEditors().
			WithNarrativeBlockedGroups().
			Only(ctx)
		if err != nil {
			return nil, err
		}

		return groupWithPermissions, nil
	}

	return nil, nil
}

func (r *mutationResolver) createGroupMembersViaClone(ctx context.Context, cloneGroupID *string, groupID string, existingMembers []*generated.GroupMembership) error {
	if cloneGroupID == nil {
		return nil
	}

	clonedGroupMembers, err := withTransactionalMutation(ctx).GroupMembership.Query().Where(groupmembership.GroupID(*cloneGroupID)).All(ctx)
	if err != nil {
		return err
	}

	var memberInput []*generated.CreateGroupMembershipInput

	for _, member := range clonedGroupMembers {
		memberExists := false
		for _, existingMember := range existingMembers {
			if member.UserID == existingMember.UserID {
				memberExists = true
			}
		}

		if !memberExists {
			memberInput = append(memberInput, &generated.CreateGroupMembershipInput{
				GroupID: groupID,
				UserID:  member.UserID,
				Role:    &member.Role,
			})
		}
	}

	if _, err := r.CreateBulkGroupMembership(ctx, memberInput); err != nil {
		return err
	}

	return nil
}
