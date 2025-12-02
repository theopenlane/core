package scim

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	scimoptional "github.com/elimity-com/scim/optional"
	scimschema "github.com/elimity-com/scim/schema"
	"github.com/samber/lo"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/group"
	"github.com/theopenlane/ent/generated/groupmembership"
	"github.com/theopenlane/ent/generated/organization"
	"github.com/theopenlane/ent/generated/orgmembership"
	"github.com/theopenlane/ent/hooks"
	"github.com/theopenlane/utils/contextx"
)

// GroupHandler implements scim.ResourceHandler for Group resources.
type GroupHandler struct{}

// NewGroupHandler creates a new GroupHandler.
func NewGroupHandler() *GroupHandler {
	return &GroupHandler{}
}

// Create stores given attributes and returns a resource with the attributes that are stored and a unique identifier.
func (h *GroupHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return scim.Resource{}, fmt.Errorf("%w: %w", ErrOrgNotFound, err)
	}

	if err := ValidateSSOEnforced(ctx, orgID); err != nil {
		return scim.Resource{}, err
	}

	ctx = contextx.With(ctx, hooks.ManagedContextKey{})

	ga, err := ExtractGroupAttributes(attributes)
	if err != nil {
		return scim.Resource{}, err
	}

	name := strings.ReplaceAll(strings.ToLower(ga.DisplayName), " ", "-")

	input := generated.CreateGroupInput{
		Name:            name,
		DisplayName:     &ga.DisplayName,
		OwnerID:         &orgID,
		ScimExternalID:  &ga.ExternalID,
		ScimDisplayName: &ga.DisplayName,
		ScimActive:      &ga.Active,
	}

	entGroup, err := client.Group.Create().SetInput(input).SetIsManaged(true).Save(ctx)
	if err != nil {
		return scim.Resource{}, HandleEntError(err, "failed to create group", fmt.Sprintf("Group with name %s already exists", name))
	}

	if len(ga.MemberIDs) > 0 {
		if err := h.addGroupMembers(ctx, entGroup.ID, orgID, ga.MemberIDs); err != nil {
			return scim.Resource{}, err
		}
	}

	entGroup, err = client.Group.Query().Where(group.ID(entGroup.ID)).WithMembers(func(gmq *generated.GroupMembershipQuery) { gmq.WithUser() }).Only(ctx)
	if err != nil {
		return scim.Resource{}, fmt.Errorf("failed to reload group: %w", err)
	}

	return h.toSCIMResource(ctx, entGroup, orgID)
}

// Get returns the resource corresponding with the given identifier.
func (h *GroupHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return scim.Resource{}, fmt.Errorf("%w: %w", ErrOrgNotFound, err)
	}

	if err := ValidateSSOEnforced(ctx, orgID); err != nil {
		return scim.Resource{}, err
	}

	entGroup, err := client.Group.Query().Where(group.ID(id), group.HasOwnerWith(organization.ID(orgID))).WithMembers(func(gmq *generated.GroupMembershipQuery) { gmq.WithUser() }).Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return scim.Resource{}, scimerrors.ScimErrorResourceNotFound(id)
		}

		return scim.Resource{}, fmt.Errorf("failed to get group: %w", err)
	}

	return h.toSCIMResource(ctx, entGroup, orgID)
}

// GetAll returns a paginated list of resources.
func (h *GroupHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return scim.Page{}, fmt.Errorf("%w: %w", ErrOrgNotFound, err)
	}

	if err := ValidateSSOEnforced(ctx, orgID); err != nil {
		return scim.Page{}, err
	}

	query := client.Group.Query().Where(group.HasOwnerWith(organization.ID(orgID)))

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return scim.Page{}, fmt.Errorf("failed to count groups: %w", err)
	}

	offset := params.StartIndex - 1
	if offset < 0 {
		offset = 0
	}

	count := params.Count
	if count <= 0 {
		count = 100
	}

	groups, err := query.Offset(offset).Limit(count).WithMembers(func(gmq *generated.GroupMembershipQuery) { gmq.WithUser() }).All(ctx)
	if err != nil {
		return scim.Page{}, fmt.Errorf("failed to list groups: %w", err)
	}

	resources := make([]scim.Resource, 0, len(groups))
	for _, g := range groups {
		resource, err := h.toSCIMResource(ctx, g, orgID)
		if err != nil {
			return scim.Page{}, err
		}

		resources = append(resources, resource)
	}

	return scim.Page{
		TotalResults: total,
		Resources:    resources,
	}, nil
}

// Replace replaces ALL existing attributes of the resource with given identifier.
func (h *GroupHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return scim.Resource{}, fmt.Errorf("%w: %w", ErrOrgNotFound, err)
	}

	if err := ValidateSSOEnforced(ctx, orgID); err != nil {
		return scim.Resource{}, err
	}

	ctx = contextx.With(ctx, hooks.ManagedContextKey{})

	entGroup, err := client.Group.Query().Where(group.ID(id), group.HasOwnerWith(organization.ID(orgID))).Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return scim.Resource{}, scimerrors.ScimErrorResourceNotFound(id)
		}

		return scim.Resource{}, fmt.Errorf("failed to get group: %w", err)
	}

	ga, err := ExtractGroupAttributes(attributes)
	if err != nil {
		return scim.Resource{}, err
	}

	name := strings.ReplaceAll(strings.ToLower(ga.DisplayName), " ", "-")

	update := client.Group.UpdateOne(entGroup).SetName(name).SetDisplayName(ga.DisplayName).SetScimDisplayName(ga.DisplayName).SetScimActive(ga.Active)

	if ga.ExternalID != "" {
		update.SetScimExternalID(ga.ExternalID)
	} else {
		update.ClearScimExternalID()
	}

	if err := update.Exec(ctx); err != nil {
		return scim.Resource{}, HandleEntError(err, "failed to update group", fmt.Sprintf("Group with name %s already exists", name))
	}

	if err := h.clearGroupMembers(ctx, id); err != nil {
		return scim.Resource{}, err
	}

	if len(ga.MemberIDs) > 0 {
		if err := h.addGroupMembers(ctx, id, orgID, ga.MemberIDs); err != nil {
			return scim.Resource{}, err
		}
	}

	updatedGroup, err := client.Group.Query().Where(group.ID(id)).WithMembers(func(gmq *generated.GroupMembershipQuery) { gmq.WithUser() }).Only(ctx)
	if err != nil {
		return scim.Resource{}, fmt.Errorf("failed to reload group: %w", err)
	}

	return h.toSCIMResource(ctx, updatedGroup, orgID)
}

// Delete removes the resource with corresponding ID.
func (h *GroupHandler) Delete(r *http.Request, id string) error {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOrgNotFound, err)
	}

	if err := ValidateSSOEnforced(ctx, orgID); err != nil {
		return err
	}

	ctx = contextx.With(ctx, hooks.ManagedContextKey{})

	count, err := client.Group.Delete().Where(group.ID(id), group.HasOwnerWith(organization.ID(orgID))).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	if count == 0 {
		return scimerrors.ScimErrorResourceNotFound(id)
	}

	return nil
}

// Patch updates one or more attributes of a SCIM resource using a sequence of operations.
func (h *GroupHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return scim.Resource{}, fmt.Errorf("%w: %w", ErrOrgNotFound, err)
	}

	if err := ValidateSSOEnforced(ctx, orgID); err != nil {
		return scim.Resource{}, err
	}

	ctx = contextx.With(ctx, hooks.ManagedContextKey{})

	entGroup, err := client.Group.Query().Where(group.ID(id), group.HasOwnerWith(organization.ID(orgID))).Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return scim.Resource{}, scimerrors.ScimErrorResourceNotFound(id)
		}

		return scim.Resource{}, fmt.Errorf("failed to get group: %w", err)
	}

	update := client.Group.UpdateOne(entGroup)
	modified := false

	for _, op := range operations {
		switch strings.ToLower(op.Op) {
		case scim.PatchOperationReplace:
			if err := h.applyReplaceOperation(ctx, update, op, id, orgID, &modified); err != nil {
				return scim.Resource{}, err
			}
		case scim.PatchOperationAdd:
			if err := h.applyAddOperation(ctx, op, id, orgID, &modified); err != nil {
				return scim.Resource{}, err
			}
		case scim.PatchOperationRemove:
			if err := h.applyRemoveOperation(ctx, op, id, &modified); err != nil {
				return scim.Resource{}, err
			}
		}
	}

	if modified {
		if err = update.Exec(ctx); err != nil {
			return scim.Resource{}, HandleEntError(err, "failed to patch group", "Group name already exists")
		}
	}

	entGroup, err = client.Group.Query().Where(group.ID(id)).WithMembers(func(gmq *generated.GroupMembershipQuery) { gmq.WithUser() }).Only(ctx)
	if err != nil {
		return scim.Resource{}, fmt.Errorf("failed to reload group: %w", err)
	}

	return h.toSCIMResource(ctx, entGroup, orgID)
}

// applyReplaceOperation applies a SCIM PATCH replace operation to a group
func (h *GroupHandler) applyReplaceOperation(ctx context.Context, update *generated.GroupUpdateOne, op scim.PatchOperation, groupID, orgID string, modified *bool) error {
	pathStr := ""
	if op.Path != nil {
		pathStr = strings.ToLower(op.Path.String())
	}

	valueMap, isMap := op.Value.(map[string]any)
	if !isMap && pathStr == "" {
		return fmt.Errorf("%w: patch operation requires path or value map", ErrInvalidAttributes)
	}

	if isMap {
		patch := ExtractPatchGroupAttribute(op)

		if patch.ExternalID != nil && *patch.ExternalID != "" {
			update.SetScimExternalID(*patch.ExternalID)
			*modified = true
		}

		if patch.DisplayName != nil && *patch.DisplayName != "" {
			name := strings.ReplaceAll(strings.ToLower(*patch.DisplayName), " ", "-")
			update.SetDisplayName(*patch.DisplayName).SetName(name).SetScimDisplayName(*patch.DisplayName)
			*modified = true
		}

		if patch.Active != nil {
			update.SetScimActive(*patch.Active)
			*modified = true
		}

		if _, ok := valueMap["members"]; ok {
			if err := h.clearGroupMembers(ctx, groupID); err != nil {
				return err
			}

			memberIDs := extractMemberIDsFromValue(op.Value)
			if len(memberIDs) > 0 {
				if err := h.addGroupMembers(ctx, groupID, orgID, memberIDs); err != nil {
					return err
				}
			}

			*modified = true
		}
	} else {
		switch pathStr {
		case "externalid":
			if strVal, ok := op.Value.(string); ok && strVal != "" {
				update.SetScimExternalID(strVal)
				*modified = true
			}
		case "displayname":
			if strVal, ok := op.Value.(string); ok && strVal != "" {
				name := strings.ReplaceAll(strings.ToLower(strVal), " ", "-")
				strippedName := hooks.StripInvalidChars(name)
				update.SetDisplayName(strVal).SetName(strippedName).SetScimDisplayName(strVal)
				*modified = true
			}
		case "active":
			if boolVal, ok := op.Value.(bool); ok {
				update.SetScimActive(boolVal)
				*modified = true
			}
		case "members":
			if err := h.clearGroupMembers(ctx, groupID); err != nil {
				return err
			}

			memberIDs := extractMemberIDsFromValue(op.Value)
			if len(memberIDs) > 0 {
				if err := h.addGroupMembers(ctx, groupID, orgID, memberIDs); err != nil {
					return err
				}
			}

			*modified = true
		}
	}

	return nil
}

// applyAddOperation applies a SCIM PATCH add operation to a group
func (h *GroupHandler) applyAddOperation(ctx context.Context, op scim.PatchOperation, groupID, orgID string, modified *bool) error {
	if op.Path == nil {
		return fmt.Errorf("%w: add operation requires path", ErrInvalidAttributes)
	}

	pathStr := strings.ToLower(op.Path.String())
	if pathStr == "members" {
		memberIDs := extractMemberIDsFromValue(op.Value)
		if len(memberIDs) > 0 {
			if err := h.addGroupMembers(ctx, groupID, orgID, memberIDs); err != nil {
				return err
			}

			*modified = true
		}
	}

	return nil
}

// applyRemoveOperation applies a SCIM PATCH remove operation to a group
func (h *GroupHandler) applyRemoveOperation(ctx context.Context, op scim.PatchOperation, groupID string, modified *bool) error {
	if op.Path == nil {
		return fmt.Errorf("%w: remove operation requires path", ErrInvalidAttributes)
	}

	pathStr := strings.ToLower(op.Path.String())
	if pathStr == "members" {
		memberIDs := extractMemberIDsFromValue(op.Value)
		if len(memberIDs) > 0 {
			if err := h.removeGroupMembers(ctx, groupID, memberIDs); err != nil {
				return err
			}

			*modified = true
		}
	}

	return nil
}

// addGroupMembers adds users to a group, verifying they are members of the organization
func (h *GroupHandler) addGroupMembers(ctx context.Context, groupID, orgID string, memberIDs []string) error {
	if len(memberIDs) == 0 {
		return nil
	}

	client := transaction.FromContext(ctx)

	for _, memberID := range memberIDs {
		exists, err := client.OrgMembership.Query().Where(orgmembership.UserID(memberID), orgmembership.OrganizationID(orgID)).Exist(ctx)
		if err != nil {
			return fmt.Errorf("failed to check org membership: %w", err)
		}

		if !exists {
			return fmt.Errorf("%w: user %s, organization %s", ErrUserNotMemberOfOrg, memberID, orgID)
		}

		if _, err := client.GroupMembership.Create().SetInput(generated.CreateGroupMembershipInput{GroupID: groupID, UserID: memberID}).Save(ctx); err != nil {
			if generated.IsNotFound(err) {
				return err
			}

			if generated.IsConstraintError(err) {
				continue
			}

			return fmt.Errorf("failed to add group member: %w", err)
		}
	}

	return nil
}

// removeGroupMembers removes specified users from a group
func (h *GroupHandler) removeGroupMembers(ctx context.Context, groupID string, memberIDs []string) error {
	if len(memberIDs) == 0 {
		return nil
	}

	client := transaction.FromContext(ctx)

	for _, memberID := range memberIDs {
		_, err := client.GroupMembership.Delete().Where(groupmembership.GroupID(groupID), groupmembership.UserID(memberID)).Exec(ctx)
		if err != nil && !generated.IsNotFound(err) {
			return fmt.Errorf("failed to remove group member: %w", err)
		}
	}

	return nil
}

// clearGroupMembers removes all users from a group
func (h *GroupHandler) clearGroupMembers(ctx context.Context, groupID string) error {
	client := transaction.FromContext(ctx)

	_, err := client.GroupMembership.Delete().Where(groupmembership.GroupID(groupID)).Exec(ctx)

	return err
}

// toSCIMResource converts an ent Group entity to a SCIM Resource representation
func (h *GroupHandler) toSCIMResource(_ any, entGroup *generated.Group, _ string) (scim.Resource, error) {
	members := make([]map[string]any, 0)
	if entGroup.Edges.Members != nil {
		groupMembers := entGroup.Edges.Members
		members = lo.Map(groupMembers, func(gm *generated.GroupMembership, _ int) map[string]any {
			if gm.Edges.User != nil {
				return map[string]any{
					"value":   gm.Edges.User.ID,
					"display": gm.Edges.User.DisplayName,
					"$ref":    fmt.Sprintf("/v1/scim/Users/%s", gm.Edges.User.ID),
				}
			}

			return map[string]any{}
		})
	}

	displayName := entGroup.DisplayName
	if entGroup.ScimDisplayName != nil && *entGroup.ScimDisplayName != "" {
		displayName = *entGroup.ScimDisplayName
	}

	attrs := scim.ResourceAttributes{
		scimschema.CommonAttributeID: entGroup.ID,
		"displayName":                displayName,
		"members":                    members,
		"active":                     entGroup.ScimActive,
	}

	meta := scim.Meta{
		Created:      &entGroup.CreatedAt,
		LastModified: &entGroup.UpdatedAt,
		Version:      fmt.Sprintf("W/\"%d\"", entGroup.UpdatedAt.Unix()),
	}

	externalID := scimoptional.NewString("")
	if entGroup.ScimExternalID != nil && *entGroup.ScimExternalID != "" {
		externalID = scimoptional.NewString(*entGroup.ScimExternalID)
	}

	return scim.Resource{
		ID:         entGroup.ID,
		ExternalID: externalID,
		Attributes: attrs,
		Meta:       meta,
	}, nil
}
