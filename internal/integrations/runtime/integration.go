package runtime

import (
	"context"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/integrations/types"
)

// ResolveIntegration resolves one integration by explicit ID with an optional definition cross-check
func (r *Runtime) ResolveIntegration(ctx context.Context, ownerID, integrationID, definitionID string) (*ent.Integration, error) {
	if integrationID == "" {
		return nil, ErrIntegrationIDRequired
	}

	query := r.DB().Integration.Query().Where(integration.IDEQ(integrationID))
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

// EnsureInstallation returns an existing installation when integrationID is provided, or creates a new one
func (r *Runtime) EnsureInstallation(ctx context.Context, ownerID, integrationID string, def types.Definition) (*ent.Integration, bool, error) {
	if integrationID != "" {
		record, err := r.ResolveIntegration(ctx, ownerID, integrationID, def.ID)
		if err != nil {
			return nil, false, err
		}

		return record, false, nil
	}

	record, err := r.DB().Integration.Create().
		SetOwnerID(ownerID).
		SetName(def.DisplayName).
		SetDefinitionID(def.ID).
		SetFamily(def.Family).
		SetStatus(enums.IntegrationStatusPending).
		Save(ctx)
	if err != nil {
		return nil, false, err
	}

	return record, true, nil
}
