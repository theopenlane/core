package operations

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// persistInternalPolicyInput upserts one InternalPolicy record through the catalog-driven entityops upsert
func persistInternalPolicyInput(ctx context.Context, db *ent.Client, integration *ent.Integration, createInput ent.CreateInternalPolicyInput) (string, error) {
	id, changed, err := persistCatalogUpsert(ctx, db, entityops.SchemaInternalPolicy, integration.OwnerID, createInput)
	if err != nil {
		return "", err
	}

	if changed && integration.ID != "" {
		if err := db.InternalPolicy.UpdateOneID(id).AddIntegrationIDs(integration.ID).Exec(entityops.WithEmissionVetoed(ctx)); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("internal policy integration link failed")

			return id, wrapIngestPersistError(err)
		}
	}

	return id, nil
}
