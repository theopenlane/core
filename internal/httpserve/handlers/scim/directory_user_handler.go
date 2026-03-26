package scim

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/elimity-com/scim"
	scimschema "github.com/elimity-com/scim/schema"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	definitionscim "github.com/theopenlane/core/internal/integrations/definitions/scim"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

const (
	// defaultSCIMPageSize is the default page size for SCIM list operations
	defaultSCIMPageSize = 100
)

// DirectoryUserHandler implements scim.ResourceHandler writing to DirectoryAccount instead of User.
// All records are scoped to the integration identified in the request context
type DirectoryUserHandler struct{}

// NewDirectoryUserHandler creates a new DirectoryUserHandler
func NewDirectoryUserHandler() *DirectoryUserHandler {
	return &DirectoryUserHandler{}
}

// Create stores a new DirectoryAccount record derived from SCIM user attributes,
// upserting by (integration_id, external_id) when a match exists
func (h *DirectoryUserHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx, client, sr, err := ResolveRequest(r)
	if err != nil {
		return scim.Resource{}, err
	}

	da, err := h.syncDirectoryAccount(ctx, client, sr, attributes, "")
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryAccountToSCIMResource(da), nil
}

// Get returns the DirectoryAccount corresponding to the given identifier, scoped by integration
func (h *DirectoryUserHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	ctx, client, sr, err := ResolveRequest(r)
	if err != nil {
		return scim.Resource{}, err
	}

	q := client.DirectoryAccount.Query().
		Where(directoryaccount.ID(id), directoryaccount.IntegrationID(sr.Installation.ID))

	da, err := lookupByID(ctx, id, q.Only)
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryAccountToSCIMResource(da), nil
}

// GetAll returns a paginated list of DirectoryAccount resources scoped by integration
func (h *DirectoryUserHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	ctx, client, sr, err := ResolveRequest(r)
	if err != nil {
		return scim.Page{}, err
	}

	query := client.DirectoryAccount.Query().Where(directoryaccount.IntegrationID(sr.Installation.ID))

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return scim.Page{}, fmt.Errorf("failed to count directory accounts: %w", err)
	}

	offset := max(params.StartIndex-1, 0)

	count := params.Count
	if count <= 0 {
		count = defaultSCIMPageSize
	}

	accounts, err := query.Offset(offset).Limit(count).All(ctx)
	if err != nil {
		return scim.Page{}, fmt.Errorf("failed to list directory accounts: %w", err)
	}

	resources := make([]scim.Resource, 0, len(accounts))
	for _, da := range accounts {
		resources = append(resources, directoryAccountToSCIMResource(da))
	}

	return scim.Page{
		TotalResults: total,
		Resources:    resources,
	}, nil
}

// Replace replaces all attributes on the DirectoryAccount identified by id
func (h *DirectoryUserHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx, client, sr, err := ResolveRequest(r)
	if err != nil {
		return scim.Resource{}, err
	}

	q := client.DirectoryAccount.Query().
		Where(directoryaccount.ID(id), directoryaccount.IntegrationID(sr.Installation.ID))

	da, err := lookupByID(ctx, id, q.Only)
	if err != nil {
		return scim.Resource{}, err
	}

	replacement := definitionscim.CloneSCIMAttributes(attributes)
	replacement["externalId"] = da.ExternalID

	updated, err := h.syncDirectoryAccount(ctx, client, sr, replacement, "")
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryAccountToSCIMResource(updated), nil
}

// Patch applies a set of patch operations to the DirectoryAccount identified by id
func (h *DirectoryUserHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	ctx, client, sr, err := ResolveRequest(r)
	if err != nil {
		return scim.Resource{}, err
	}

	q := client.DirectoryAccount.Query().
		Where(directoryaccount.ID(id), directoryaccount.IntegrationID(sr.Installation.ID))

	da, err := lookupByID(ctx, id, q.Only)
	if err != nil {
		return scim.Resource{}, err
	}

	patched := directoryAccountAttributesFromRecord(da)
	if err := applyDirectoryAccountPatchOperations(patched, operations); err != nil {
		return scim.Resource{}, err
	}

	updated, err := h.syncDirectoryAccount(ctx, client, sr, patched, "")
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryAccountToSCIMResource(updated), nil
}

// Delete sets the DirectoryAccount status to DELETED
func (h *DirectoryUserHandler) Delete(r *http.Request, id string) error {
	ctx, client, sr, err := ResolveRequest(r)
	if err != nil {
		return err
	}

	q := client.DirectoryAccount.Query().
		Where(directoryaccount.ID(id), directoryaccount.IntegrationID(sr.Installation.ID))

	da, err := lookupByID(ctx, id, q.Only)
	if err != nil {
		return err
	}

	attributes := directoryAccountAttributesFromRecord(da)
	_, err = h.syncDirectoryAccount(ctx, client, sr, attributes, definitionscim.DeleteAction)

	return err
}

// syncDirectoryAccount upserts one SCIM directory account through the inline operation path
func (h *DirectoryUserHandler) syncDirectoryAccount(ctx context.Context, client *generated.Client, sr *SCIMRequest, attributes scim.ResourceAttributes, action string) (*generated.DirectoryAccount, error) {
	payloadSet, err := definitionscim.BuildDirectoryAccountPayloadSet(attributes, action)
	if err != nil {
		return nil, handleIngestError(err, "directory account payload is invalid")
	}

	externalID := definitionscim.DirectoryAccountExternalID(attributes)
	if err := ingestPayloadSets(ctx, client, sr, []integrationtypes.IngestPayloadSet{payloadSet}); err != nil {
		return nil, handleIngestError(err, "directory account payload is invalid")
	}

	q := client.DirectoryAccount.Query().
		Where(directoryaccount.IntegrationID(sr.Installation.ID), directoryaccount.ExternalID(externalID))

	da, err := q.Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, definitionscim.ErrDirectoryAccountNotFound
		}

		return nil, fmt.Errorf("failed to reload directory account: %w", err)
	}

	return da, nil
}

// applyDirectoryAccountPatchOperations applies SCIM PATCH operations to account attributes before ingest
func applyDirectoryAccountPatchOperations(attributes scim.ResourceAttributes, operations []scim.PatchOperation) error {
	for _, op := range operations {
		switch strings.ToLower(op.Op) {
		case scim.PatchOperationReplace, scim.PatchOperationAdd:
			applyDirectoryAccountPatch(attributes, op)
		case scim.PatchOperationRemove:
			applyDirectoryAccountRemove(attributes, op)
		default:
			return fmt.Errorf("%w: unsupported patch operation %s", definitionscim.ErrInvalidAttributes, op.Op)
		}
	}

	return nil
}

// applyDirectoryAccountPatch applies a replace or add patch operation to SCIM account attributes
func applyDirectoryAccountPatch(attributes scim.ResourceAttributes, op scim.PatchOperation) {
	pathStr := ""
	if op.Path != nil {
		pathStr = strings.ToLower(op.Path.String())
	}

	if pathStr == "" {
		valueMap, ok := op.Value.(map[string]any)
		if !ok {
			return
		}

		definitionscim.MergeSCIMMap(attributes, valueMap)

		return
	}

	switch pathStr {
	case "displayname":
		if v, ok := op.Value.(string); ok && v != "" {
			attributes["displayName"] = v
		}
	case "name.givenname":
		if v, ok := op.Value.(string); ok && v != "" {
			name := definitionscim.EnsureSCIMMap(attributes, "name")
			name["givenName"] = v
		}
	case "name.familyname":
		if v, ok := op.Value.(string); ok && v != "" {
			name := definitionscim.EnsureSCIMMap(attributes, "name")
			name["familyName"] = v
		}
	case "active":
		if v, ok := op.Value.(bool); ok {
			attributes["active"] = v
		}
	case "username":
		if v, ok := op.Value.(string); ok && v != "" {
			attributes["userName"] = v
		}
	case "externalid":
		if v, ok := op.Value.(string); ok && v != "" {
			attributes["externalId"] = v
		}
	case "emails":
		if v, ok := op.Value.([]any); ok {
			attributes["emails"] = definitionscim.CloneSCIMValue(v)
		}
	}
}

// applyDirectoryAccountRemove applies a remove patch operation to SCIM account attributes
func applyDirectoryAccountRemove(attributes scim.ResourceAttributes, op scim.PatchOperation) {
	if op.Path == nil {
		return
	}

	pathStr := strings.ToLower(op.Path.String())

	switch pathStr {
	case "name.givenname":
		delete(definitionscim.EnsureSCIMMap(attributes, "name"), "givenName")
	case "name.familyname":
		delete(definitionscim.EnsureSCIMMap(attributes, "name"), "familyName")
	case "displayname":
		delete(attributes, "displayName")
	case "emails":
		delete(attributes, "emails")
	}
}

// directoryAccountToSCIMResource converts a DirectoryAccount entity to a SCIM Resource
func directoryAccountToSCIMResource(da *generated.DirectoryAccount) scim.Resource {
	return buildSCIMResource(da.ID, da.ExternalID, da.CreatedAt, da.UpdatedAt, directoryAccountAttributesFromRecord(da))
}

// directoryAccountAttributesFromRecord renders a DirectoryAccount as SCIM attributes for patching and delete ingest
func directoryAccountAttributesFromRecord(da *generated.DirectoryAccount) scim.ResourceAttributes {
	canonicalEmail := ""
	if da.CanonicalEmail != nil {
		canonicalEmail = *da.CanonicalEmail
	}

	givenName := ""
	if da.GivenName != nil {
		givenName = *da.GivenName
	}

	familyName := ""
	if da.FamilyName != nil {
		familyName = *da.FamilyName
	}

	attrs := scim.ResourceAttributes{
		scimschema.CommonAttributeID: da.ID,
		"externalId":                 da.ExternalID,
		"userName":                   canonicalEmail,
		"displayName":                da.DisplayName,
		"name": map[string]any{
			"givenName":  givenName,
			"familyName": familyName,
		},
		"active": da.Status == enums.DirectoryAccountStatusActive,
	}

	if canonicalEmail != "" {
		attrs["emails"] = []map[string]any{
			{
				"value":   canonicalEmail,
				"primary": true,
			},
		}
	}

	return attrs
}
