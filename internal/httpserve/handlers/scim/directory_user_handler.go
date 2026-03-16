package scim

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	scimoptional "github.com/elimity-com/scim/optional"
	scimschema "github.com/elimity-com/scim/schema"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/directorysyncrun"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/middleware/transaction"
)

const (
	// scimSyncRunCursor is the sentinel source_cursor value used to identify SCIM-sourced sync runs
	scimSyncRunCursor = "scim"
	// defaultSCIMPageSize is the default page size for SCIM list operations
	defaultSCIMPageSize = 100
)

// DirectoryUserHandler implements scim.ResourceHandler writing to DirectoryAccount instead of User.
// All records are scoped to the integration identified in the request context.
type DirectoryUserHandler struct{}

// NewDirectoryUserHandler creates a new DirectoryUserHandler
func NewDirectoryUserHandler() *DirectoryUserHandler {
	return &DirectoryUserHandler{}
}

// Create stores a new DirectoryAccount record derived from SCIM user attributes,
// upserting by (integration_id, external_id) when a match exists.
func (h *DirectoryUserHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx).Client()

	ic, ok := IntegrationContextFromContext(ctx)
	if !ok || ic == nil {
		return scim.Resource{}, ErrIntegrationIDRequired
	}

	da, err := h.syncDirectoryAccount(ctx, client, ic, attributes, "")
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryAccountToSCIMResource(da), nil
}

// Get returns the DirectoryAccount corresponding to the given identifier, scoped by integration
func (h *DirectoryUserHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx).Client()

	ic, ok := IntegrationContextFromContext(ctx)
	if !ok || ic == nil {
		return scim.Resource{}, ErrIntegrationIDRequired
	}

	da, err := client.DirectoryAccount.Query().
		Where(directoryaccount.ID(id), directoryaccount.IntegrationID(ic.IntegrationID)).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return scim.Resource{}, scimerrors.ScimErrorResourceNotFound(id)
		}

		return scim.Resource{}, fmt.Errorf("failed to get directory account: %w", err)
	}

	return directoryAccountToSCIMResource(da), nil
}

// GetAll returns a paginated list of DirectoryAccount resources scoped by integration
func (h *DirectoryUserHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx).Client()

	ic, ok := IntegrationContextFromContext(ctx)
	if !ok || ic == nil {
		return scim.Page{}, ErrIntegrationIDRequired
	}

	query := client.DirectoryAccount.Query().Where(directoryaccount.IntegrationID(ic.IntegrationID))

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
	ctx := r.Context()
	client := transaction.FromContext(ctx).Client()

	ic, ok := IntegrationContextFromContext(ctx)
	if !ok || ic == nil {
		return scim.Resource{}, ErrIntegrationIDRequired
	}

	da, err := client.DirectoryAccount.Query().
		Where(directoryaccount.ID(id), directoryaccount.IntegrationID(ic.IntegrationID)).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return scim.Resource{}, scimerrors.ScimErrorResourceNotFound(id)
		}

		return scim.Resource{}, fmt.Errorf("failed to get directory account: %w", err)
	}

	replacement := cloneSCIMAttributes(attributes)
	replacement["externalId"] = da.ExternalID

	updated, err := h.syncDirectoryAccount(ctx, client, ic, replacement, "")
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryAccountToSCIMResource(updated), nil
}

// Patch applies a set of patch operations to the DirectoryAccount identified by id
func (h *DirectoryUserHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx).Client()

	ic, ok := IntegrationContextFromContext(ctx)
	if !ok || ic == nil {
		return scim.Resource{}, ErrIntegrationIDRequired
	}

	da, err := client.DirectoryAccount.Query().
		Where(directoryaccount.ID(id), directoryaccount.IntegrationID(ic.IntegrationID)).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return scim.Resource{}, scimerrors.ScimErrorResourceNotFound(id)
		}

		return scim.Resource{}, fmt.Errorf("failed to get directory account: %w", err)
	}

	patched := directoryAccountAttributesFromRecord(da)
	if err := applyDirectoryAccountPatchOperations(patched, operations); err != nil {
		return scim.Resource{}, err
	}

	updated, err := h.syncDirectoryAccount(ctx, client, ic, patched, "")
	if err != nil {
		return scim.Resource{}, err
	}

	return directoryAccountToSCIMResource(updated), nil
}

// Delete sets the DirectoryAccount status to DELETED
func (h *DirectoryUserHandler) Delete(r *http.Request, id string) error {
	ctx := r.Context()
	client := transaction.FromContext(ctx).Client()

	ic, ok := IntegrationContextFromContext(ctx)
	if !ok || ic == nil {
		return ErrIntegrationIDRequired
	}

	da, err := client.DirectoryAccount.Query().
		Where(directoryaccount.ID(id), directoryaccount.IntegrationID(ic.IntegrationID)).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return scimerrors.ScimErrorResourceNotFound(id)
		}

		return fmt.Errorf("failed to get directory account: %w", err)
	}

	return h.deleteDirectoryAccount(ctx, client, ic, da)
}

// syncDirectoryAccount upserts one SCIM directory account through Runtime and reloads the normalized record
func (h *DirectoryUserHandler) syncDirectoryAccount(ctx context.Context, client *generated.Client, ic *IntegrationContext, attributes scim.ResourceAttributes, action string) (*generated.DirectoryAccount, error) {
	payloadSet, err := buildDirectoryAccountPayloadSet(attributes, action)
	if err != nil {
		return nil, handleDirectoryIngestError(err, "directory account payload is invalid")
	}

	if err := ingestDirectoryPayloadSets(ctx, client, ic, []integrationtypes.IngestPayloadSet{payloadSet}); err != nil {
		return nil, handleDirectoryIngestError(err, "directory account payload is invalid")
	}

	externalID := directoryAccountExternalID(attributes)
	da, err := client.DirectoryAccount.Query().
		Where(directoryaccount.IntegrationID(ic.IntegrationID), directoryaccount.ExternalID(externalID)).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, ErrDirectoryAccountNotFound
		}

		return nil, fmt.Errorf("failed to reload directory account: %w", err)
	}

	return da, nil
}

// deleteDirectoryAccount marks a directory account as deleted through Runtime ingest
func (h *DirectoryUserHandler) deleteDirectoryAccount(ctx context.Context, client *generated.Client, ic *IntegrationContext, da *generated.DirectoryAccount) error {
	attributes := directoryAccountAttributesFromRecord(da)
	_, err := h.syncDirectoryAccount(ctx, client, ic, attributes, scimDeleteAction)

	return err
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
			return fmt.Errorf("%w: unsupported patch operation %s", ErrInvalidAttributes, op.Op)
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

		mergeSCIMMap(attributes, valueMap)

		return
	}

	switch pathStr {
	case "displayname":
		if v, ok := op.Value.(string); ok && v != "" {
			attributes["displayName"] = v
		}
	case "name.givenname":
		if v, ok := op.Value.(string); ok && v != "" {
			name := ensureSCIMMap(attributes, "name")
			name["givenName"] = v
		}
	case "name.familyname":
		if v, ok := op.Value.(string); ok && v != "" {
			name := ensureSCIMMap(attributes, "name")
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
			attributes["emails"] = cloneSCIMValue(v)
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
		delete(ensureSCIMMap(attributes, "name"), "givenName")
	case "name.familyname":
		delete(ensureSCIMMap(attributes, "name"), "familyName")
	case "displayname":
		delete(attributes, "displayName")
	case "emails":
		delete(attributes, "emails")
	}
}

// directoryAccountToSCIMResource converts a DirectoryAccount entity to a SCIM Resource
func directoryAccountToSCIMResource(da *generated.DirectoryAccount) scim.Resource {
	attrs := directoryAccountAttributesFromRecord(da)
	delete(attrs, "externalId")

	externalID := scimoptional.NewString("")
	if da.ExternalID != "" {
		externalID = scimoptional.NewString(da.ExternalID)
	}

	meta := scim.Meta{
		Created:      &da.CreatedAt,
		LastModified: &da.UpdatedAt,
		Version:      fmt.Sprintf("W/\"%d\"", da.UpdatedAt.Unix()),
	}

	return scim.Resource{
		ID:         da.ID,
		ExternalID: externalID,
		Attributes: attrs,
		Meta:       meta,
	}
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

// ensureScimSyncRun gets or creates a sentinel DirectorySyncRun for SCIM-sourced directory writes.
// It queries for an existing run with source_cursor="scim" for the given integration, creating one
// if none exists, and returns its ID.
func ensureScimSyncRun(ctx context.Context, client *generated.Client, integrationID, orgID string) (string, error) {
	existing, err := client.DirectorySyncRun.Query().
		Where(
			directorysyncrun.IntegrationID(integrationID),
			directorysyncrun.SourceCursor(scimSyncRunCursor),
		).
		Order(directorysyncrun.ByCreatedAt(sql.OrderDesc())).
		First(ctx)
	if err != nil && !generated.IsNotFound(err) {
		return "", fmt.Errorf("failed to query scim sync run: %w", err)
	}

	if existing != nil {
		return existing.ID, nil
	}

	run, err := client.DirectorySyncRun.Create().
		SetIntegrationID(integrationID).
		SetOwnerID(orgID).
		SetStatus(enums.DirectorySyncRunStatusRunning).
		SetSourceCursor(scimSyncRunCursor).
		Save(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create scim sync run: %w", err)
	}

	return run.ID, nil
}
