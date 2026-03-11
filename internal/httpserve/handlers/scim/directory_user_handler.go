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
	"entgo.io/ent/dialect/sql"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/directorysyncrun"
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
	client := transaction.FromContext(ctx)

	ic, ok := IntegrationContextFromContext(ctx)
	if !ok || ic == nil {
		return scim.Resource{}, ErrIntegrationIDRequired
	}

	ua, err := ExtractUserAttributes(attributes)
	if err != nil {
		return scim.Resource{}, err
	}

	existing, err := client.DirectoryAccount.Query().
		Where(directoryaccount.IntegrationID(ic.IntegrationID), directoryaccount.ExternalID(ua.ExternalID)).
		Only(ctx)
	if err != nil && !generated.IsNotFound(err) {
		return scim.Resource{}, fmt.Errorf("failed to query directory account: %w", err)
	}

	if existing != nil {
		return h.updateDirectoryAccount(ctx, client, existing, ua)
	}

	return h.createDirectoryAccount(ctx, client, ua)
}

// Get returns the DirectoryAccount corresponding to the given identifier, scoped by integration
func (h *DirectoryUserHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

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
	client := transaction.FromContext(ctx)

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
	client := transaction.FromContext(ctx)

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

	ua, err := ExtractUserAttributes(attributes)
	if err != nil {
		return scim.Resource{}, err
	}

	return h.updateDirectoryAccount(ctx, client, da, ua)
}

// Patch applies a set of patch operations to the DirectoryAccount identified by id
func (h *DirectoryUserHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

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

	update := client.DirectoryAccount.UpdateOne(da)
	modified := false

	for _, op := range operations {
		switch strings.ToLower(op.Op) {
		case scim.PatchOperationReplace, scim.PatchOperationAdd:
			applyDirectoryAccountPatch(update, op, &modified)
		case scim.PatchOperationRemove:
			applyDirectoryAccountRemove(update, op, &modified)
		}
	}

	if modified {
		da, err = update.Save(ctx)
		if err != nil {
			return scim.Resource{}, HandleEntError(err, "failed to patch directory account", "constraint violation")
		}
	}

	return directoryAccountToSCIMResource(da), nil
}

// Delete sets the DirectoryAccount status to DELETED
func (h *DirectoryUserHandler) Delete(r *http.Request, id string) error {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

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

	return client.DirectoryAccount.UpdateOne(da).SetStatus(enums.DirectoryAccountStatusDeleted).Exec(ctx)
}

// createDirectoryAccount creates a new DirectoryAccount record from SCIM user attributes
func (h *DirectoryUserHandler) createDirectoryAccount(ctx context.Context, client *generated.Tx, ua *UserAttributes) (scim.Resource, error) {
	create := client.DirectoryAccount.Create().
		SetExternalID(ua.ExternalID).
		SetDisplayName(ua.DisplayName)

	if ua.Email != "" {
		create.SetCanonicalEmail(strings.ToLower(ua.Email))
	}

	if ua.FirstName != "" {
		create.SetGivenName(ua.FirstName)
	}

	if ua.LastName != "" {
		create.SetFamilyName(ua.LastName)
	}

	status := enums.DirectoryAccountStatusActive
	if !ua.Active {
		status = enums.DirectoryAccountStatusInactive
	}

	create.SetStatus(status)

	da, err := create.Save(ctx)
	if err != nil {
		return scim.Resource{}, HandleEntError(err, "failed to create directory account", fmt.Sprintf("directory account with externalId %s already exists", ua.ExternalID))
	}

	return directoryAccountToSCIMResource(da), nil
}

// updateDirectoryAccount applies UserAttributes to an existing DirectoryAccount record
func (h *DirectoryUserHandler) updateDirectoryAccount(ctx context.Context, client *generated.Tx, da *generated.DirectoryAccount, ua *UserAttributes) (scim.Resource, error) {
	update := client.DirectoryAccount.UpdateOne(da).
		SetDisplayName(ua.DisplayName)

	if ua.Email != "" {
		update.SetCanonicalEmail(strings.ToLower(ua.Email))
	}

	if ua.FirstName != "" {
		update.SetGivenName(ua.FirstName)
	} else {
		update.ClearGivenName()
	}

	if ua.LastName != "" {
		update.SetFamilyName(ua.LastName)
	} else {
		update.ClearFamilyName()
	}

	status := enums.DirectoryAccountStatusActive
	if !ua.Active {
		status = enums.DirectoryAccountStatusInactive
	}

	update.SetStatus(status)

	updated, err := update.Save(ctx)
	if err != nil {
		return scim.Resource{}, HandleEntError(err, "failed to update directory account", "constraint violation")
	}

	return directoryAccountToSCIMResource(updated), nil
}

// applyDirectoryAccountPatch applies a replace/add patch operation to a DirectoryAccount update
func applyDirectoryAccountPatch(update *generated.DirectoryAccountUpdateOne, op scim.PatchOperation, modified *bool) {
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
	case "name.givenname":
		if v, ok := op.Value.(string); ok && v != "" {
			update.SetGivenName(v)
			*modified = true
		}
	case "name.familyname":
		if v, ok := op.Value.(string); ok && v != "" {
			update.SetFamilyName(v)
			*modified = true
		}
	case "active":
		if v, ok := op.Value.(bool); ok {
			status := enums.DirectoryAccountStatusActive
			if !v {
				status = enums.DirectoryAccountStatusInactive
			}

			update.SetStatus(status)
			*modified = true
		}
	}
}

// applyDirectoryAccountRemove applies a remove patch operation to a DirectoryAccount update
func applyDirectoryAccountRemove(update *generated.DirectoryAccountUpdateOne, op scim.PatchOperation, modified *bool) {
	if op.Path == nil {
		return
	}

	pathStr := strings.ToLower(op.Path.String())

	switch pathStr {
	case "name.givenname":
		update.ClearGivenName()
		*modified = true
	case "name.familyname":
		update.ClearFamilyName()
		*modified = true
	}
}

// directoryAccountToSCIMResource converts a DirectoryAccount entity to a SCIM Resource
func directoryAccountToSCIMResource(da *generated.DirectoryAccount) scim.Resource {
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

	active := da.Status == enums.DirectoryAccountStatusActive

	attrs := scim.ResourceAttributes{
		scimschema.CommonAttributeID: da.ID,
		"userName":                   canonicalEmail,
		"displayName":                da.DisplayName,
		"name": map[string]any{
			"givenName":  givenName,
			"familyName": familyName,
		},
		"emails": []map[string]any{
			{
				"value":   canonicalEmail,
				"primary": true,
			},
		},
		"active": active,
	}

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

// ensureScimSyncRun gets or creates a sentinel DirectorySyncRun for SCIM-sourced directory writes.
// It queries for an existing run with source_cursor="scim" for the given integration, creating one
// if none exists, and returns its ID.
func ensureScimSyncRun(ctx context.Context, client *generated.Tx, integrationID, orgID string) (string, error) {
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
