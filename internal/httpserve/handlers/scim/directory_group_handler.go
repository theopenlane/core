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

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/pkg/middleware/transaction"
)

// DirectoryGroupHandler implements scim.ResourceHandler writing to DirectoryGroup instead of Group.
// All records are scoped to the integration identified in the request context.
type DirectoryGroupHandler struct{}

// NewDirectoryGroupHandler creates a new DirectoryGroupHandler
func NewDirectoryGroupHandler() *DirectoryGroupHandler {
	return &DirectoryGroupHandler{}
}

// Create stores a new DirectoryGroup record derived from SCIM group attributes,
// upserting by (integration_id, external_id) when a match exists.
func (h *DirectoryGroupHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	ic, ok := IntegrationContextFromContext(ctx)
	if !ok || ic == nil {
		return scim.Resource{}, ErrIntegrationIDRequired
	}

	ga, err := ExtractGroupAttributes(attributes)
	if err != nil {
		return scim.Resource{}, err
	}

	existing, err := client.DirectoryGroup.Query().
		Where(directorygroup.IntegrationID(ic.IntegrationID), directorygroup.ExternalID(ga.ExternalID)).
		Only(ctx)
	if err != nil && !generated.IsNotFound(err) {
		return scim.Resource{}, fmt.Errorf("failed to query directory group: %w", err)
	}

	if existing != nil {
		return h.updateDirectoryGroup(ctx, client, ic, existing, ga)
	}

	syncRunID, err := ensureScimSyncRun(ctx, client, ic.IntegrationID, ic.OrgID)
	if err != nil {
		return scim.Resource{}, ErrSyncRunRequired
	}

	return h.createDirectoryGroup(ctx, client, ic, syncRunID, ga)
}

// Get returns the DirectoryGroup corresponding to the given identifier, scoped by integration
func (h *DirectoryGroupHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	ic, ok := IntegrationContextFromContext(ctx)
	if !ok || ic == nil {
		return scim.Resource{}, ErrIntegrationIDRequired
	}

	dg, err := client.DirectoryGroup.Query().
		Where(directorygroup.ID(id), directorygroup.IntegrationID(ic.IntegrationID)).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return scim.Resource{}, scimerrors.ScimErrorResourceNotFound(id)
		}

		return scim.Resource{}, fmt.Errorf("failed to get directory group: %w", err)
	}

	members, err := h.loadGroupMembers(ctx, client, dg.ID)
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryGroupToSCIMResource(dg, members), nil
}

// GetAll returns a paginated list of DirectoryGroup resources scoped by integration
func (h *DirectoryGroupHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	ic, ok := IntegrationContextFromContext(ctx)
	if !ok || ic == nil {
		return scim.Page{}, ErrIntegrationIDRequired
	}

	query := client.DirectoryGroup.Query().Where(directorygroup.IntegrationID(ic.IntegrationID))

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return scim.Page{}, fmt.Errorf("failed to count directory groups: %w", err)
	}

	offset := max(params.StartIndex-1, 0)

	count := params.Count
	if count <= 0 {
		count = defaultSCIMPageSize
	}

	groups, err := query.Offset(offset).Limit(count).All(ctx)
	if err != nil {
		return scim.Page{}, fmt.Errorf("failed to list directory groups: %w", err)
	}

	resources := make([]scim.Resource, 0, len(groups))
	for _, dg := range groups {
		members, err := h.loadGroupMembers(ctx, client, dg.ID)
		if err != nil {
			return scim.Page{}, err
		}

		resources = append(resources, directoryGroupToSCIMResource(dg, members))
	}

	return scim.Page{
		TotalResults: total,
		Resources:    resources,
	}, nil
}

// Replace replaces all attributes on the DirectoryGroup identified by id
func (h *DirectoryGroupHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	ic, ok := IntegrationContextFromContext(ctx)
	if !ok || ic == nil {
		return scim.Resource{}, ErrIntegrationIDRequired
	}

	dg, err := client.DirectoryGroup.Query().
		Where(directorygroup.ID(id), directorygroup.IntegrationID(ic.IntegrationID)).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return scim.Resource{}, scimerrors.ScimErrorResourceNotFound(id)
		}

		return scim.Resource{}, fmt.Errorf("failed to get directory group: %w", err)
	}

	ga, err := ExtractGroupAttributes(attributes)
	if err != nil {
		return scim.Resource{}, err
	}

	return h.updateDirectoryGroup(ctx, client, ic, dg, ga)
}

// Patch applies a set of patch operations to the DirectoryGroup identified by id
func (h *DirectoryGroupHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	ic, ok := IntegrationContextFromContext(ctx)
	if !ok || ic == nil {
		return scim.Resource{}, ErrIntegrationIDRequired
	}

	dg, err := client.DirectoryGroup.Query().
		Where(directorygroup.ID(id), directorygroup.IntegrationID(ic.IntegrationID)).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return scim.Resource{}, scimerrors.ScimErrorResourceNotFound(id)
		}

		return scim.Resource{}, fmt.Errorf("failed to get directory group: %w", err)
	}

	update := client.DirectoryGroup.UpdateOne(dg)
	modified := false

	for _, op := range operations {
		switch strings.ToLower(op.Op) {
		case scim.PatchOperationReplace:
			if err := h.applyGroupReplaceOp(ctx, client, ic, update, op, id, &modified); err != nil {
				return scim.Resource{}, err
			}
		case scim.PatchOperationAdd:
			if err := h.applyGroupAddOp(ctx, client, ic, op, id, &modified); err != nil {
				return scim.Resource{}, err
			}
		case scim.PatchOperationRemove:
			if err := h.applyGroupRemoveOp(ctx, client, op, id, &modified); err != nil {
				return scim.Resource{}, err
			}
		}
	}

	if modified {
		if err := update.Exec(ctx); err != nil {
			return scim.Resource{}, HandleEntError(err, "failed to patch directory group", "constraint violation")
		}
	}

	dg, err = client.DirectoryGroup.Query().
		Where(directorygroup.ID(id), directorygroup.IntegrationID(ic.IntegrationID)).
		Only(ctx)
	if err != nil {
		return scim.Resource{}, fmt.Errorf("failed to reload directory group: %w", err)
	}

	members, err := h.loadGroupMembers(ctx, client, id)
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryGroupToSCIMResource(dg, members), nil
}

// Delete sets the DirectoryGroup status to DELETED
func (h *DirectoryGroupHandler) Delete(r *http.Request, id string) error {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	ic, ok := IntegrationContextFromContext(ctx)
	if !ok || ic == nil {
		return ErrIntegrationIDRequired
	}

	dg, err := client.DirectoryGroup.Query().
		Where(directorygroup.ID(id), directorygroup.IntegrationID(ic.IntegrationID)).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return scimerrors.ScimErrorResourceNotFound(id)
		}

		return fmt.Errorf("failed to get directory group: %w", err)
	}

	return client.DirectoryGroup.UpdateOne(dg).SetStatus(enums.DirectoryGroupStatusDeleted).Exec(ctx)
}

// createDirectoryGroup creates a new DirectoryGroup record from SCIM group attributes
func (h *DirectoryGroupHandler) createDirectoryGroup(ctx context.Context, client *generated.Tx, ic *IntegrationContext, syncRunID string, ga *GroupAttributes) (scim.Resource, error) {
	create := client.DirectoryGroup.Create().
		SetIntegrationID(ic.IntegrationID).
		SetDirectorySyncRunID(syncRunID).
		SetExternalID(ga.ExternalID).
		SetDisplayName(ga.DisplayName)

	status := enums.DirectoryGroupStatusActive
	if !ga.Active {
		status = enums.DirectoryGroupStatusInactive
	}

	create.SetStatus(status)

	dg, err := create.Save(ctx)
	if err != nil {
		return scim.Resource{}, HandleEntError(err, "failed to create directory group", fmt.Sprintf("directory group with externalId %s already exists", ga.ExternalID))
	}

	if len(ga.MemberIDs) > 0 {
		if err := h.addGroupMembers(ctx, client, ic, dg.ID, syncRunID, ga.MemberIDs); err != nil {
			return scim.Resource{}, err
		}
	}

	members, err := h.loadGroupMembers(ctx, client, dg.ID)
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryGroupToSCIMResource(dg, members), nil
}

// updateDirectoryGroup updates an existing DirectoryGroup record from SCIM group attributes
func (h *DirectoryGroupHandler) updateDirectoryGroup(ctx context.Context, client *generated.Tx, ic *IntegrationContext, dg *generated.DirectoryGroup, ga *GroupAttributes) (scim.Resource, error) {
	update := client.DirectoryGroup.UpdateOne(dg).SetDisplayName(ga.DisplayName)

	status := enums.DirectoryGroupStatusActive
	if !ga.Active {
		status = enums.DirectoryGroupStatusInactive
	}

	update.SetStatus(status)

	if err := update.Exec(ctx); err != nil {
		return scim.Resource{}, HandleEntError(err, "failed to update directory group", "constraint violation")
	}

	if err := h.clearGroupMembers(ctx, client, dg.ID); err != nil {
		return scim.Resource{}, err
	}

	if len(ga.MemberIDs) > 0 {
		syncRunID, err := ensureScimSyncRun(ctx, client, ic.IntegrationID, ic.OrgID)
		if err != nil {
			return scim.Resource{}, ErrSyncRunRequired
		}

		if err := h.addGroupMembers(ctx, client, ic, dg.ID, syncRunID, ga.MemberIDs); err != nil {
			return scim.Resource{}, err
		}
	}

	updatedDG, err := client.DirectoryGroup.Query().
		Where(directorygroup.ID(dg.ID), directorygroup.IntegrationID(ic.IntegrationID)).
		Only(ctx)
	if err != nil {
		return scim.Resource{}, fmt.Errorf("failed to reload directory group: %w", err)
	}

	members, err := h.loadGroupMembers(ctx, client, dg.ID)
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryGroupToSCIMResource(updatedDG, members), nil
}

// addGroupMembers creates DirectoryMembership records linking DirectoryAccount IDs to the group
func (h *DirectoryGroupHandler) addGroupMembers(ctx context.Context, client *generated.Tx, ic *IntegrationContext, groupID, syncRunID string, memberIDs []string) error {
	for _, memberID := range memberIDs {
		if _, err := client.DirectoryMembership.Create().
			SetIntegrationID(ic.IntegrationID).
			SetDirectorySyncRunID(syncRunID).
			SetDirectoryAccountID(memberID).
			SetDirectoryGroupID(groupID).
			SetOwnerID(ic.OrgID).
			Save(ctx); err != nil {
			if generated.IsConstraintError(err) {
				continue
			}

			return fmt.Errorf("failed to add directory group member: %w", err)
		}
	}

	return nil
}

// removeGroupMembers removes specified DirectoryAccount IDs from the group
func (h *DirectoryGroupHandler) removeGroupMembers(ctx context.Context, client *generated.Tx, groupID string, memberIDs []string) error {
	for _, memberID := range memberIDs {
		if _, err := client.DirectoryMembership.Delete().
			Where(directorymembership.DirectoryGroupID(groupID), directorymembership.DirectoryAccountID(memberID)).
			Exec(ctx); err != nil && !generated.IsNotFound(err) {
			return fmt.Errorf("failed to remove directory group member: %w", err)
		}
	}

	return nil
}

// clearGroupMembers removes all DirectoryMembership records for the given group
func (h *DirectoryGroupHandler) clearGroupMembers(ctx context.Context, client *generated.Tx, groupID string) error {
	_, err := client.DirectoryMembership.Delete().
		Where(directorymembership.DirectoryGroupID(groupID)).
		Exec(ctx)

	return err
}

// loadGroupMembers returns the DirectoryAccount IDs and display names for all members of a group
func (h *DirectoryGroupHandler) loadGroupMembers(ctx context.Context, client *generated.Tx, groupID string) ([]*generated.DirectoryMembership, error) {
	return client.DirectoryMembership.Query().
		Where(directorymembership.DirectoryGroupID(groupID)).
		All(ctx)
}

// applyGroupReplaceOp applies a SCIM PATCH replace operation to a directory group
func (h *DirectoryGroupHandler) applyGroupReplaceOp(ctx context.Context, client *generated.Tx, ic *IntegrationContext, update *generated.DirectoryGroupUpdateOne, op scim.PatchOperation, groupID string, modified *bool) error {
	pathStr := ""
	if op.Path != nil {
		pathStr = strings.ToLower(op.Path.String())
	}

	switch pathStr {
	case "displayname":
		if v, ok := op.Value.(string); ok && v != "" {
			update.SetDisplayName(v)
			*modified = true
		}
	case "active":
		if v, ok := op.Value.(bool); ok {
			status := enums.DirectoryGroupStatusActive
			if !v {
				status = enums.DirectoryGroupStatusInactive
			}

			update.SetStatus(status)
			*modified = true
		}
	case "members":
		if err := h.clearGroupMembers(ctx, client, groupID); err != nil {
			return err
		}

		memberIDs := extractMemberIDsFromValue(op.Value)
		if len(memberIDs) > 0 {
			syncRunID, err := ensureScimSyncRun(ctx, client, ic.IntegrationID, ic.OrgID)
			if err != nil {
				return ErrSyncRunRequired
			}

			if err := h.addGroupMembers(ctx, client, ic, groupID, syncRunID, memberIDs); err != nil {
				return err
			}
		}

		*modified = true
	}

	return nil
}

// applyGroupAddOp applies a SCIM PATCH add operation to a directory group
func (h *DirectoryGroupHandler) applyGroupAddOp(ctx context.Context, client *generated.Tx, ic *IntegrationContext, op scim.PatchOperation, groupID string, modified *bool) error {
	if op.Path == nil {
		return fmt.Errorf("%w: add operation requires path", ErrInvalidAttributes)
	}

	pathStr := strings.ToLower(op.Path.String())
	if pathStr != "members" {
		return nil
	}

	memberIDs := extractMemberIDsFromValue(op.Value)
	if len(memberIDs) == 0 {
		return nil
	}

	syncRunID, err := ensureScimSyncRun(ctx, client, ic.IntegrationID, ic.OrgID)
	if err != nil {
		return ErrSyncRunRequired
	}

	if err := h.addGroupMembers(ctx, client, ic, groupID, syncRunID, memberIDs); err != nil {
		return err
	}

	*modified = true

	return nil
}

// applyGroupRemoveOp applies a SCIM PATCH remove operation to a directory group
func (h *DirectoryGroupHandler) applyGroupRemoveOp(ctx context.Context, client *generated.Tx, op scim.PatchOperation, groupID string, modified *bool) error {
	if op.Path == nil {
		return fmt.Errorf("%w: remove operation requires path", ErrInvalidAttributes)
	}

	pathStr := strings.ToLower(op.Path.String())
	if pathStr != "members" {
		return nil
	}

	memberIDs := extractMemberIDsFromValue(op.Value)
	if len(memberIDs) == 0 {
		return nil
	}

	if err := h.removeGroupMembers(ctx, client, groupID, memberIDs); err != nil {
		return err
	}

	*modified = true

	return nil
}

// directoryGroupToSCIMResource converts a DirectoryGroup entity and its memberships to a SCIM Resource
func directoryGroupToSCIMResource(dg *generated.DirectoryGroup, memberships []*generated.DirectoryMembership) scim.Resource {
	members := make([]map[string]any, 0, len(memberships))
	for _, m := range memberships {
		members = append(members, map[string]any{
			"value": m.DirectoryAccountID,
			"$ref":  fmt.Sprintf("/v1/scim/Users/%s", m.DirectoryAccountID),
		})
	}

	displayName := dg.DisplayName

	active := dg.Status == enums.DirectoryGroupStatusActive

	attrs := scim.ResourceAttributes{
		scimschema.CommonAttributeID: dg.ID,
		"displayName":                displayName,
		"members":                    members,
		"active":                     active,
	}

	externalID := scimoptional.NewString("")
	if dg.ExternalID != "" {
		externalID = scimoptional.NewString(dg.ExternalID)
	}

	meta := scim.Meta{
		Created:      &dg.CreatedAt,
		LastModified: &dg.UpdatedAt,
		Version:      fmt.Sprintf("W/\"%d\"", dg.UpdatedAt.Unix()),
	}

	return scim.Resource{
		ID:         dg.ID,
		ExternalID: externalID,
		Attributes: attrs,
		Meta:       meta,
	}
}
