package operations

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// persistFindingInput upserts one Finding record through the catalog-driven entityops upsert
func persistFindingInput(ctx context.Context, db *ent.Client, integration *ent.Integration, createInput ent.CreateFindingInput) (string, error) {
	if createInput.Description != nil && *createInput.Description != "" {
		normalized := normalizeDescription(*createInput.Description)
		createInput.Description = &normalized
	}

	id, changed, err := persistCatalogUpsert(ctx, db, entityops.SchemaFinding, integration.OwnerID, createInput)
	if err != nil {
		return "", err
	}

	if changed && integration.ID != "" {
		if err := db.Finding.UpdateOneID(id).AddIntegrationIDs(integration.ID).Exec(entityops.WithEmissionVetoed(ctx)); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("finding integration link failed")

			return id, wrapIngestPersistError(err)
		}
	}

	return id, nil
}
