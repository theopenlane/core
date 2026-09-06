package operations

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// persistEntityInput upserts one Entity record through the catalog-driven entityops upsert. The
// payload's owner takes priority over the integration owner so direct callers (questionnaire
// transform) can target the organization they resolved
func persistEntityInput(ctx context.Context, db *ent.Client, integration *ent.Integration, createInput ent.CreateEntityInput) (string, error) {
	ownerID := integration.OwnerID
	if createInput.OwnerID != nil && *createInput.OwnerID != "" {
		ownerID = *createInput.OwnerID
	}

	if ownerID == "" {
		return "", ErrIngestUpsertKeyMissing
	}

	id, changed, err := persistCatalogUpsert(ctx, db, entityops.SchemaEntity, ownerID, createInput)
	if err != nil {
		return "", err
	}

	if changed && integration.ID != "" {
		if err := db.Entity.UpdateOneID(id).AddIntegrationIDs(integration.ID).Exec(entityops.WithEmissionVetoed(ctx)); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("entity integration link failed")

			return id, wrapIngestPersistError(err)
		}
	}

	return id, nil
}
