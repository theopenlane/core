package runtime

import (
	"context"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/iam/auth"
)

type entInstallationResolver struct {
	db *ent.Client
}

func (r entInstallationResolver) ResolveInstallation(ctx context.Context, installationID string) (keymaker.InstallationRecord, error) {
	if installationID == "" {
		return keymaker.InstallationRecord{}, keymaker.ErrInstallationIDRequired
	}

	query := r.db.Integration.Query().Where(integration.IDEQ(installationID))

	caller, ok := auth.CallerFromContext(ctx)
	if ok && caller != nil && caller.OrganizationID != "" {
		query = query.Where(integration.OwnerIDEQ(caller.OrganizationID))
	}

	record, err := query.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return keymaker.InstallationRecord{}, keymaker.ErrInstallationNotFound
		}

		return keymaker.InstallationRecord{}, err
	}

	return keymaker.InstallationRecord{
		ID:           record.ID,
		OwnerID:      record.OwnerID,
		DefinitionID: types.DefinitionID(record.DefinitionID),
	}, nil
}
