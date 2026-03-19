package runtime

import (
	"context"

	"github.com/samber/do/v2"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/integrations/types"
)

// ResolveInstallation resolves one installation by explicit ID or by owner plus definition
func (r *Runtime) ResolveInstallation(ctx context.Context, ownerID, installationID string, definitionID string) (*ent.Integration, error) {
	db := do.MustInvoke[*ent.Client](r.injector)
	if installationID != "" {
		query := db.Integration.Query().Where(integration.IDEQ(installationID))
		if ownerID != "" {
			query = query.Where(integration.OwnerIDEQ(ownerID))
		}

		record, err := query.Only(ctx)
		if err != nil {
			return nil, err
		}

		if definitionID != "" && record.DefinitionID != string(definitionID) {
			return nil, ErrInstallationDefinitionMismatch
		}

		return record, nil
	}

	if definitionID == "" {
		return nil, ErrDefinitionIDRequired
	}

	if ownerID == "" {
		return nil, ErrOwnerIDRequired
	}

	record, err := db.Integration.Query().
		Where(
			integration.OwnerIDEQ(ownerID),
			integration.DefinitionIDEQ(string(definitionID)),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// EnsureInstallation returns an existing installation for the owner and definition or creates a new
// Pending one when none exists. When an explicit installationID is given the record must already exist.
// The boolean return value indicates whether a new record was created
func (r *Runtime) EnsureInstallation(ctx context.Context, ownerID, installationID string, def types.Definition) (*ent.Integration, bool, error) {
	db := do.MustInvoke[*ent.Client](r.injector)

	if installationID != "" {
		record, err := r.ResolveInstallation(ctx, ownerID, installationID, def.ID)
		if err != nil {
			return nil, false, err
		}

		return record, false, nil
	}

	existing, err := r.ResolveInstallation(ctx, ownerID, "", def.ID)
	if err == nil {
		return existing, false, nil
	}

	if !ent.IsNotFound(err) {
		return nil, false, err
	}

	record, err := db.Integration.Create().
		SetOwnerID(ownerID).
		SetName(def.DisplayName).
		SetDefinitionID(def.ID).
		SetDefinitionSlug(def.Slug).
		SetFamily(def.Family).
		SetStatus(enums.IntegrationStatusPending).
		Save(ctx)
	if err != nil {
		return nil, false, err
	}

	return record, true, nil
}

// MarkConnected updates the installation's status to Connected in the database and on the in-memory record
func (r *Runtime) MarkConnected(ctx context.Context, installation *ent.Integration) error {
	db := do.MustInvoke[*ent.Client](r.injector)

	if err := db.Integration.UpdateOneID(installation.ID).
		SetStatus(enums.IntegrationStatusConnected).
		Exec(ctx); err != nil {
		return err
	}

	installation.Status = enums.IntegrationStatusConnected

	return nil
}

// DeleteInstallation permanently removes the installation record by ID
func (r *Runtime) DeleteInstallation(ctx context.Context, installationID string) error {
	db := do.MustInvoke[*ent.Client](r.injector)

	return db.Integration.DeleteOneID(installationID).Exec(ctx)
}
