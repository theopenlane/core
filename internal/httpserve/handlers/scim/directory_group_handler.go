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
	client := transaction.FromContext(ctx).Client()

	ic, ok := IntegrationContextFromContext(ctx)
	if !ok || ic == nil {
		return scim.Resource{}, ErrIntegrationIDRequired
	}

	dg, members, err := h.syncDirectoryGroup(ctx, client, ic, attributes, "")
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryGroupToSCIMResource(dg, members), nil
}

// Get returns the DirectoryGroup corresponding to the given identifier, scoped by integration
func (h *DirectoryGroupHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx).Client()

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
	client := transaction.FromContext(ctx).Client()

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
	client := transaction.FromContext(ctx).Client()

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

	replacement := cloneSCIMAttributes(attributes)
	replacement["externalId"] = dg.ExternalID

	if err := h.clearGroupMembers(ctx, client, dg.ID); err != nil {
		return scim.Resource{}, err
	}

	updated, members, err := h.syncDirectoryGroup(ctx, client, ic, replacement, "")
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryGroupToSCIMResource(updated, members), nil
}

// Patch applies a set of patch operations to the DirectoryGroup identified by id
func (h *DirectoryGroupHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx).Client()

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

	currentMembers, err := h.loadGroupMembers(ctx, client, dg.ID)
	if err != nil {
		return scim.Resource{}, err
	}

	patched := directoryGroupAttributesFromRecord(dg, currentMembers)
	membersTouched, err := applyGroupPatchOperations(patched, operations)
	if err != nil {
		return scim.Resource{}, err
	}

	if membersTouched {
		if err := h.clearGroupMembers(ctx, client, dg.ID); err != nil {
			return scim.Resource{}, err
		}
	}

	updated, members, err := h.syncDirectoryGroup(ctx, client, ic, patched, "")
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryGroupToSCIMResource(updated, members), nil
}

// Delete sets the DirectoryGroup status to DELETED
func (h *DirectoryGroupHandler) Delete(r *http.Request, id string) error {
	ctx := r.Context()
	client := transaction.FromContext(ctx).Client()

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

	members, err := h.loadGroupMembers(ctx, client, dg.ID)
	if err != nil {
		return err
	}

	attributes := directoryGroupAttributesFromRecord(dg, members)
	_, _, err = h.syncDirectoryGroup(ctx, client, ic, attributes, scimDeleteAction)

	return err
}

// syncDirectoryGroup upserts one SCIM directory group and its memberships through Runtime and reloads the normalized records
func (h *DirectoryGroupHandler) syncDirectoryGroup(ctx context.Context, client *generated.Client, ic *IntegrationContext, attributes scim.ResourceAttributes, action string) (*generated.DirectoryGroup, []*generated.DirectoryMembership, error) {
	payloadSets, err := buildDirectoryGroupPayloadSets(attributes, action)
	if err != nil {
		return nil, nil, handleDirectoryIngestError(err, "directory group payload is invalid")
	}

	if err := ingestDirectoryPayloadSets(ctx, client, ic, payloadSets); err != nil {
		return nil, nil, handleDirectoryIngestError(err, "directory group payload is invalid")
	}

	externalID := directoryGroupExternalID(attributes)
	dg, err := client.DirectoryGroup.Query().
		Where(directorygroup.IntegrationID(ic.IntegrationID), directorygroup.ExternalID(externalID)).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, nil, ErrDirectoryGroupNotFound
		}

		return nil, nil, fmt.Errorf("failed to reload directory group: %w", err)
	}

	members, err := h.loadGroupMembers(ctx, client, dg.ID)
	if err != nil {
		return nil, nil, err
	}

	return dg, members, nil
}

// clearGroupMembers removes all DirectoryMembership records for the given group
func (h *DirectoryGroupHandler) clearGroupMembers(ctx context.Context, client *generated.Client, groupID string) error {
	_, err := client.DirectoryMembership.Delete().
		Where(directorymembership.DirectoryGroupID(groupID)).
		Exec(ctx)

	return err
}

// loadGroupMembers returns the DirectoryAccount IDs and display names for all members of a group
func (h *DirectoryGroupHandler) loadGroupMembers(ctx context.Context, client *generated.Client, groupID string) ([]*generated.DirectoryMembership, error) {
	return client.DirectoryMembership.Query().
		Where(directorymembership.DirectoryGroupID(groupID)).
		All(ctx)
}

// applyGroupPatchOperations applies SCIM PATCH operations to group attributes before ingest
func applyGroupPatchOperations(attributes scim.ResourceAttributes, operations []scim.PatchOperation) (bool, error) {
	membersTouched := false

	for _, op := range operations {
		switch strings.ToLower(op.Op) {
		case scim.PatchOperationReplace:
			membersTouched = applyGroupReplaceOp(attributes, op) || membersTouched
		case scim.PatchOperationAdd:
			membersTouched = applyGroupAddOp(attributes, op) || membersTouched
		case scim.PatchOperationRemove:
			membersTouched = applyGroupRemoveOp(attributes, op) || membersTouched
		default:
			return false, fmt.Errorf("%w: unsupported patch operation %s", ErrInvalidAttributes, op.Op)
		}
	}

	return membersTouched, nil
}

// applyGroupReplaceOp applies a SCIM PATCH replace operation to group attributes
func applyGroupReplaceOp(attributes scim.ResourceAttributes, op scim.PatchOperation) bool {
	pathStr := ""
	if op.Path != nil {
		pathStr = strings.ToLower(op.Path.String())
	}

	if pathStr == "" {
		valueMap, ok := op.Value.(map[string]any)
		if !ok {
			return false
		}

		mergeSCIMMap(attributes, valueMap)

		_, membersTouched := valueMap["members"]

		return membersTouched
	}

	switch pathStr {
	case "displayname":
		if v, ok := op.Value.(string); ok && v != "" {
			attributes["displayName"] = v
		}
	case "active":
		if v, ok := op.Value.(bool); ok {
			attributes["active"] = v
		}
	case "externalid":
		if v, ok := op.Value.(string); ok && v != "" {
			attributes["externalId"] = v
		}
	case "members":
		attributes["members"] = memberRefsFromIDs(extractMemberIDsFromValue(op.Value))
		return true
	}

	return false
}

// applyGroupAddOp applies a SCIM PATCH add operation to group attributes
func applyGroupAddOp(attributes scim.ResourceAttributes, op scim.PatchOperation) bool {
	if op.Path == nil {
		valueMap, ok := op.Value.(map[string]any)
		if !ok {
			return false
		}

		mergeSCIMMap(attributes, valueMap)

		_, membersTouched := valueMap["members"]

		return membersTouched
	}

	pathStr := strings.ToLower(op.Path.String())
	if pathStr != "members" {
		return false
	}

	current := extractMemberIDsFromValue(attributes["members"])
	additions := extractMemberIDsFromValue(op.Value)
	attributes["members"] = memberRefsFromIDs(append(current, additions...))

	return len(additions) > 0
}

// applyGroupRemoveOp applies a SCIM PATCH remove operation to group attributes
func applyGroupRemoveOp(attributes scim.ResourceAttributes, op scim.PatchOperation) bool {
	if op.Path == nil {
		return false
	}

	pathStr := strings.ToLower(op.Path.String())
	if pathStr != "members" {
		return false
	}

	removals := extractMemberIDsFromValue(op.Value)
	if len(removals) == 0 {
		delete(attributes, "members")
		return true
	}

	current := extractMemberIDsFromValue(attributes["members"])
	filtered := make([]string, 0, len(current))
	for _, memberID := range current {
		if !containsString(removals, memberID) {
			filtered = append(filtered, memberID)
		}
	}

	attributes["members"] = memberRefsFromIDs(filtered)

	return true
}

// directoryGroupToSCIMResource converts a DirectoryGroup entity and its memberships to a SCIM Resource
func directoryGroupToSCIMResource(dg *generated.DirectoryGroup, memberships []*generated.DirectoryMembership) scim.Resource {
	attrs := directoryGroupAttributesFromRecord(dg, memberships)
	delete(attrs, "externalId")

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

// directoryGroupAttributesFromRecord renders a DirectoryGroup as SCIM attributes for patching and delete ingest
func directoryGroupAttributesFromRecord(dg *generated.DirectoryGroup, memberships []*generated.DirectoryMembership) scim.ResourceAttributes {
	return scim.ResourceAttributes{
		scimschema.CommonAttributeID: dg.ID,
		"externalId":                 dg.ExternalID,
		"displayName":                dg.DisplayName,
		"members":                    memberRefsFromMemberships(memberships),
		"active":                     dg.Status == enums.DirectoryGroupStatusActive,
	}
}

// memberRefsFromMemberships renders SCIM member references from DirectoryMembership rows
func memberRefsFromMemberships(memberships []*generated.DirectoryMembership) []map[string]any {
	members := make([]map[string]any, 0, len(memberships))
	for _, membership := range memberships {
		members = append(members, map[string]any{
			"value": membership.DirectoryAccountID,
			"$ref":  fmt.Sprintf("/v1/scim/Users/%s", membership.DirectoryAccountID),
		})
	}

	return members
}

// memberRefsFromIDs renders SCIM member references from account IDs
func memberRefsFromIDs(memberIDs []string) []map[string]any {
	memberIDs = uniqueStrings(memberIDs)

	members := make([]map[string]any, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		members = append(members, map[string]any{
			"value": memberID,
			"$ref":  fmt.Sprintf("/v1/scim/Users/%s", memberID),
		})
	}

	return members
}

// containsString reports whether a slice contains a given value
func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}

	return false
}
