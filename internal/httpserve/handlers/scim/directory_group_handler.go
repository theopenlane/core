package scim

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/elimity-com/scim"
	scimschema "github.com/elimity-com/scim/schema"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/internal/ent/generated/directorymembership"
	definitionscim "github.com/theopenlane/core/internal/integrations/definitions/scim"
)

// DirectoryGroupHandler implements scim.ResourceHandler writing to DirectoryGroup instead of Group.
// All records are scoped to the integration identified in the request context
type DirectoryGroupHandler struct{}

// NewDirectoryGroupHandler creates a new DirectoryGroupHandler
func NewDirectoryGroupHandler() *DirectoryGroupHandler {
	return &DirectoryGroupHandler{}
}

// Create stores a new DirectoryGroup record derived from SCIM group attributes,
// upserting by (integration_id, external_id) when a match exists
func (h *DirectoryGroupHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx, client, sr, err := ResolveRequest(r)
	if err != nil {
		return scim.Resource{}, err
	}

	dg, members, err := h.syncDirectoryGroup(ctx, client, sr, attributes, "")
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryGroupToSCIMResource(sr.BasePath, dg, members), nil
}

// Get returns the DirectoryGroup corresponding to the given identifier, scoped by integration
func (h *DirectoryGroupHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	ctx, client, sr, err := ResolveRequest(r)
	if err != nil {
		return scim.Resource{}, err
	}

	q := client.DirectoryGroup.Query().
		Where(directorygroup.ID(id), directorygroup.IntegrationID(sr.Installation.ID))

	dg, err := lookupByID(ctx, id, q.Only)
	if err != nil {
		return scim.Resource{}, err
	}

	members, err := h.loadGroupMembers(ctx, client, dg.ID)
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryGroupToSCIMResource(sr.BasePath, dg, members), nil
}

// GetAll returns a paginated list of DirectoryGroup resources scoped by integration
func (h *DirectoryGroupHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	ctx, client, sr, err := ResolveRequest(r)
	if err != nil {
		return scim.Page{}, err
	}

	query := client.DirectoryGroup.Query().Where(directorygroup.IntegrationID(sr.Installation.ID))

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

		resources = append(resources, directoryGroupToSCIMResource(sr.BasePath, dg, members))
	}

	return scim.Page{
		TotalResults: total,
		Resources:    resources,
	}, nil
}

// Replace replaces all attributes on the DirectoryGroup identified by id
func (h *DirectoryGroupHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx, client, sr, err := ResolveRequest(r)
	if err != nil {
		return scim.Resource{}, err
	}

	q := client.DirectoryGroup.Query().
		Where(directorygroup.ID(id), directorygroup.IntegrationID(sr.Installation.ID))

	dg, err := lookupByID(ctx, id, q.Only)
	if err != nil {
		return scim.Resource{}, err
	}

	replacement := definitionscim.CloneSCIMAttributes(attributes)
	replacement["externalId"] = dg.ExternalID

	if err := h.clearGroupMembers(ctx, client, dg.ID); err != nil {
		return scim.Resource{}, err
	}

	updated, members, err := h.syncDirectoryGroup(ctx, client, sr, replacement, "")
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryGroupToSCIMResource(sr.BasePath, updated, members), nil
}

// Patch applies a set of patch operations to the DirectoryGroup identified by id
func (h *DirectoryGroupHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	ctx, client, sr, err := ResolveRequest(r)
	if err != nil {
		return scim.Resource{}, err
	}

	q := client.DirectoryGroup.Query().
		Where(directorygroup.ID(id), directorygroup.IntegrationID(sr.Installation.ID))

	dg, err := lookupByID(ctx, id, q.Only)
	if err != nil {
		return scim.Resource{}, err
	}

	currentMembers, err := h.loadGroupMembers(ctx, client, dg.ID)
	if err != nil {
		return scim.Resource{}, err
	}

	patched := directoryGroupAttributesFromRecord(sr.BasePath, dg, currentMembers)
	membersTouched, err := applyGroupPatchOperations(patched, operations)
	if err != nil {
		return scim.Resource{}, err
	}

	if membersTouched {
		if err := h.clearGroupMembers(ctx, client, dg.ID); err != nil {
			return scim.Resource{}, err
		}
	}

	updated, members, err := h.syncDirectoryGroup(ctx, client, sr, patched, "")
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryGroupToSCIMResource(sr.BasePath, updated, members), nil
}

// Delete sets the DirectoryGroup status to DELETED
func (h *DirectoryGroupHandler) Delete(r *http.Request, id string) error {
	ctx, client, sr, err := ResolveRequest(r)
	if err != nil {
		return err
	}

	q := client.DirectoryGroup.Query().
		Where(directorygroup.ID(id), directorygroup.IntegrationID(sr.Installation.ID))

	dg, err := lookupByID(ctx, id, q.Only)
	if err != nil {
		return err
	}

	members, err := h.loadGroupMembers(ctx, client, dg.ID)
	if err != nil {
		return err
	}

	attributes := directoryGroupAttributesFromRecord(sr.BasePath, dg, members)
	_, _, err = h.syncDirectoryGroup(ctx, client, sr, attributes, definitionscim.DeleteAction)

	return err
}

// syncDirectoryGroup upserts one SCIM directory group and its memberships through the inline operation path
func (h *DirectoryGroupHandler) syncDirectoryGroup(ctx context.Context, client *generated.Client, sr *SCIMRequest, attributes scim.ResourceAttributes, action string) (*generated.DirectoryGroup, []*generated.DirectoryMembership, error) {
	payloadSets, err := definitionscim.BuildDirectoryGroupPayloadSets(attributes, action)
	if err != nil {
		return nil, nil, handleIngestError(err, "directory group payload is invalid")
	}

	externalID := definitionscim.DirectoryGroupExternalID(attributes)
	if err := ingestPayloadSets(ctx, client, sr, payloadSets); err != nil {
		return nil, nil, handleIngestError(err, "directory group payload is invalid")
	}

	dg, err := client.DirectoryGroup.Query().
		Where(
			directorygroup.IntegrationID(sr.Installation.ID),
			directorygroup.ExternalID(externalID),
		).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, nil, definitionscim.ErrDirectoryGroupNotFound
		}

		return nil, nil, fmt.Errorf("failed to reload directory group: %w", err)
	}

	members, err := h.loadGroupMembers(ctx, client, dg.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to reload directory memberships: %w", err)
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
			return false, fmt.Errorf("%w: unsupported patch operation %s", definitionscim.ErrInvalidAttributes, op.Op)
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

		definitionscim.MergeSCIMMap(attributes, valueMap)

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
		attributes["members"] = memberRefsFromIDs("", definitionscim.ExtractMemberIDsFromValue(op.Value))
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

		definitionscim.MergeSCIMMap(attributes, valueMap)

		_, membersTouched := valueMap["members"]

		return membersTouched
	}

	pathStr := strings.ToLower(op.Path.String())
	if pathStr != "members" {
		return false
	}

	current := definitionscim.ExtractMemberIDsFromValue(attributes["members"])
	additions := definitionscim.ExtractMemberIDsFromValue(op.Value)
	attributes["members"] = memberRefsFromIDs("", append(current, additions...))

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

	removals := definitionscim.ExtractMemberIDsFromValue(op.Value)
	if len(removals) == 0 {
		delete(attributes, "members")
		return true
	}

	current := definitionscim.ExtractMemberIDsFromValue(attributes["members"])
	attributes["members"] = memberRefsFromIDs("", lo.Without(current, removals...))

	return true
}

// directoryGroupToSCIMResource converts a DirectoryGroup entity and its memberships to a SCIM Resource
func directoryGroupToSCIMResource(basePath string, dg *generated.DirectoryGroup, memberships []*generated.DirectoryMembership) scim.Resource {
	return buildSCIMResource(dg.ID, dg.ExternalID, dg.CreatedAt, dg.UpdatedAt, directoryGroupAttributesFromRecord(basePath, dg, memberships))
}

// directoryGroupAttributesFromRecord renders a DirectoryGroup as SCIM attributes for patching and delete ingest
func directoryGroupAttributesFromRecord(basePath string, dg *generated.DirectoryGroup, memberships []*generated.DirectoryMembership) scim.ResourceAttributes {
	return scim.ResourceAttributes{
		scimschema.CommonAttributeID: dg.ID,
		"externalId":                 dg.ExternalID,
		"displayName":                dg.DisplayName,
		"members":                    memberRefsFromMemberships(basePath, memberships),
		"active":                     dg.Status == enums.DirectoryGroupStatusActive,
	}
}

// memberRefsFromMemberships renders SCIM member references from DirectoryMembership rows
func memberRefsFromMemberships(basePath string, memberships []*generated.DirectoryMembership) []map[string]any {
	return lo.Map(memberships, func(m *generated.DirectoryMembership, _ int) map[string]any {
		return map[string]any{
			"value": m.DirectoryAccountID,
			"$ref":  scimUserRef(basePath, m.DirectoryAccountID),
		}
	})
}

// memberRefsFromIDs renders SCIM member references from account IDs
func memberRefsFromIDs(basePath string, memberIDs []string) []map[string]any {
	return lo.Map(lo.Uniq(memberIDs), func(id string, _ int) map[string]any {
		return map[string]any{
			"value": id,
			"$ref":  scimUserRef(basePath, id),
		}
	})
}

func scimUserRef(basePath, memberID string) string {
	if basePath == "" {
		return fmt.Sprintf("Users/%s", memberID)
	}

	return fmt.Sprintf("%s/Users/%s", basePath, memberID)
}
