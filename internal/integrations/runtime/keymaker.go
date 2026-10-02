package runtime

import (
	"context"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/keymaker"
)

// lookupKeymakerInstallation resolves one installation into keymaker's lookup contract
func (r *Runtime) lookupKeymakerInstallation(ctx context.Context, integrationID string) (keymaker.InstallationRecord, error) {
	if integrationID == "" {
		return keymaker.InstallationRecord{}, keymaker.ErrInstallationIDRequired
	}

	return keymakerRecord(r.ResolveIntegration(ctx, IntegrationLookup{IntegrationID: integrationID}))
}

// keymakerRecord maps a resolved installation and its lookup error onto keymaker's record and sentinels
func keymakerRecord(record *ent.Integration, err error) (keymaker.InstallationRecord, error) {
	switch {
	case ent.IsNotFound(err):
		return keymaker.InstallationRecord{}, keymaker.ErrInstallationNotFound
	case err != nil:
		return keymaker.InstallationRecord{}, err
	}

	return keymaker.InstallationRecord{
		ID:           record.ID,
		OwnerID:      record.OwnerID,
		DefinitionID: record.DefinitionID,
	}, nil
}
