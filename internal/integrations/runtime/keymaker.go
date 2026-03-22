package runtime

import (
	"context"
	"errors"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/keymaker"
)

// lookupKeymakerInstallation adapts runtime installation resolution to keymaker's lookup contract
func (r *Runtime) lookupKeymakerInstallation(ctx context.Context, installationID string) (keymaker.InstallationRecord, error) {
	return resolveKeymakerInstallation(ctx, installationID, r.ResolveInstallation)
}

// resolveKeymakerInstallation resolves one installation record for keymaker using an explicit installation ID
func resolveKeymakerInstallation(ctx context.Context, installationID string, resolve func(context.Context, string, string, string) (*ent.Integration, error)) (keymaker.InstallationRecord, error) {
	if installationID == "" {
		return keymaker.InstallationRecord{}, keymaker.ErrInstallationIDRequired
	}

	record, err := resolve(ctx, "", installationID, "")
	if err != nil {
		switch {
		case errors.Is(err, ErrInstallationNotFound):
			return keymaker.InstallationRecord{}, keymaker.ErrInstallationNotFound
		default:
			return keymaker.InstallationRecord{}, err
		}
	}

	return keymaker.InstallationRecord{
		ID:           record.ID,
		OwnerID:      record.OwnerID,
		DefinitionID: record.DefinitionID,
	}, nil
}
