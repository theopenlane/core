package runtime

import (
	"context"
	"errors"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keymaker"
)

func (r *Runtime) lookupKeymakerInstallation(ctx context.Context, installationID string) (keymaker.InstallationRecord, error) {
	return resolveKeymakerInstallation(ctx, installationID, r.ResolveInstallation)
}

func (r *Runtime) persistKeymakerAuthCompletion(ctx context.Context, installationID string, definition types.Definition, result types.AuthCompleteResult) error {
	return mapKeymakerPersistAuthError(r.PersistAuthCompletion(ctx, installationID, definition, result))
}

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

func mapKeymakerPersistAuthError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrInstallationIDRequired):
		return keymaker.ErrInstallationIDRequired
	case errors.Is(err, ErrInstallationNotFound):
		return keymaker.ErrInstallationNotFound
	default:
		return err
	}
}
