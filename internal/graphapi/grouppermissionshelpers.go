package graphapi

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/common/enums"
)

// EntObject is a struct that contains the id, displayID, and name of an object
type EntObject struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	RefCode   string `json:"ref_code,omitempty"`
	DisplayID string `json:"display_id,omitempty"`
}

// getGroupPermissions returns a slice of GroupPermissions for the given object type and permission
func getGroupPermissions[T any](obj []T, objectType string, permission enums.Permission) (perms []*model.GroupPermissionEdge) {
	for _, e := range obj {
		eo, err := convertToEntObject(e)
		if err != nil {
			return nil
		}

		name := eo.Name
		if name == "" {
			name = eo.RefCode
		}

		perms = append(perms, &model.GroupPermissionEdge{
			Node: &model.GroupPermission{
				ObjectType:  objectType,
				ID:          eo.ID,
				Permissions: permission,
				DisplayID:   &eo.DisplayID,
				Name:        &name,
			},
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

// getGroupByIDWithPermissionsEdges returns a group object with all the permissions edges
// TODO (sfunk): This function is a good candidate for a generated function with the group permissions mixin
func getGroupByIDWithPermissionsEdges(ctx context.Context, groupID *string) (*generated.Group, error) {
	if groupID == nil || *groupID == "" {
		return nil, nil
	}

	groupWithPermissions, err := withTransactionalMutation(ctx).Group.
		Query().
		Where(group.IDEQ(*groupID)).
		// Control permissions
		WithControlEditors().
		WithControlBlockedGroups().

		// Control Mapped permissions
		WithMappedControlEditors().
		WithMappedControlBlockedGroups().

		// Control Implementation permissions
		WithControlImplementationViewers().
		WithControlImplementationEditors().
		WithControlImplementationBlockedGroups().

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

func (r *mutationResolver) createGroupMembersViaClone(ctx context.Context, cloneGroupID *string, groupID string, existingMembers []*generated.GroupMembership) error {
	// if there is no clone group id, return
	if cloneGroupID == nil || *cloneGroupID == "" {
		return nil
	}

	// get all the members of the cloned group
	clonedGroupMembers, err := withTransactionalMutation(ctx).GroupMembership.Query().Where(groupmembership.GroupID(*cloneGroupID)).All(ctx)
	if err != nil {
		return err
	}

	var memberInput []*generated.CreateGroupMembershipInput

	// this happens after the original group mutation, so we need to ensure
	// we don't add the same members to the group
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

	// no members to add
	if len(memberInput) == 0 {
		return nil
	}

	// create all the group memberships
	if _, err := r.CreateBulkGroupMembership(ctx, memberInput); err != nil {
		return err
	}

	return nil
}
