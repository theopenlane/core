package runtime

import (
	"context"

	"github.com/samber/do/v2"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
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
			return nil, wrapInstallationQueryError(err)
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
		return nil, wrapInstallationQueryError(err)
	}

	return record, nil
}

// wrapInstallationQueryError maps ent Only(ctx) result errors to installation-specific sentinels
func wrapInstallationQueryError(err error) error {
	switch {
	case err == nil:
		return nil
	case ent.IsNotFound(err):
		return ErrInstallationNotFound
	case ent.IsNotSingular(err):
		return ErrInstallationAmbiguous
	default:
		return err
	}
}
