package scim

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/elimity-com/scim"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	definitionscim "github.com/theopenlane/core/internal/integrations/definitions/scim"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

const (
	// defaultSCIMPageSize is the default page size for SCIM list operations
	defaultSCIMPageSize = 100
)

// DirectoryUserHandler implements scim.ResourceHandler writing to DirectoryAccount instead of User.
// All records are scoped to the integration identified in the request context
type DirectoryUserHandler struct {
	// Runtime provides shared integration execution capabilities
	Runtime *integrationsruntime.Runtime
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
	applyDirectoryAccountPatchOperations(patched, operations)

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
func (h *DirectoryUserHandler) syncDirectoryAccount(ctx context.Context, client *generated.Client, sr *Request, attributes scim.ResourceAttributes, action string) (*generated.DirectoryAccount, error) {
	payloadSet, err := definitionscim.BuildDirectoryAccountPayloadSet(attributes, action)
	if err != nil {
		return nil, handleIngestError(err, "directory account payload is invalid")
	}

	externalID := definitionscim.DirectoryAccountExternalID(attributes)
	if err := ingestPayloadSets(ctx, client, h.Runtime, sr.Installation, []integrationtypes.IngestPayloadSet{payloadSet}); err != nil {
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

// applyDirectoryAccountPatchOperations applies SCIM PATCH operations to account attributes before ingest.
// The library has already validated operations against the composed schema, so all paths
// and value types are guaranteed valid
func applyDirectoryAccountPatchOperations(attributes scim.ResourceAttributes, operations []scim.PatchOperation) {
	for _, op := range operations {
		switch strings.ToLower(op.Op) {
		case scim.PatchOperationReplace, scim.PatchOperationAdd:
			applyPatchValue(attributes, op)
		case scim.PatchOperationRemove:
			removePatchValue(attributes, op)
		}
	}
}

// directoryAccountToSCIMResource converts a DirectoryAccount entity to a SCIM Resource
func directoryAccountToSCIMResource(da *generated.DirectoryAccount) scim.Resource {
	return buildSCIMResource(da.ID, da.ExternalID, da.CreatedAt, da.UpdatedAt, directoryAccountAttributesFromRecord(da))
}

// directoryAccountAttributesFromRecord renders a DirectoryAccount as SCIM resource attributes.
// For patching and delete ingest the externalId is re-added by the caller
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
		"externalId":  da.ExternalID,
		"userName":    canonicalEmail,
		"displayName": da.DisplayName,
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
