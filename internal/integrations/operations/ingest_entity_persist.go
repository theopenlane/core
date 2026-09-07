package operations

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// persistEntityInput upserts one Entity record through the catalog-driven entityops upsert. The
// payload's owner takes priority over the integration owner so direct callers (questionnaire
// transform) can target the organization they resolved
func persistEntityInput(ctx context.Context, db *ent.Client, installation *ent.Integration, createInput ent.CreateEntityInput) (string, error) {
	ownerID := installation.OwnerID
	if createInput.OwnerID != nil && *createInput.OwnerID != "" {
		ownerID = *createInput.OwnerID
	}

	if ownerID == "" {
		return "", ErrIngestUpsertKeyMissing
	}

	id, _, err := persistCatalogUpsert(ctx, db, entityops.SchemaEntity, ownerID, createInput)
	if err != nil {
		return "", err
	}

	if err := ensureIngestIntegrationLink(ctx, entityops.SchemaEntity.Snake, id, installation.ID,
		func(ctx context.Context) ([]string, error) {
			return db.Entity.Query().Where(entity.HasIntegrationsWith(integration.ID(installation.ID))).IDs(ctx)
		},
		func(ctx context.Context) (bool, error) {
			return db.Entity.Query().Where(entity.ID(id), entity.HasIntegrationsWith(integration.ID(installation.ID))).Exist(ctx)
		},
		func(ctx context.Context) error {
			return db.Entity.UpdateOneID(id).AddIntegrationIDs(installation.ID).Exec(ctx)
		},
	); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("entity integration link failed")

		return id, wrapIngestPersistError(err)
	}

	return id, nil
}
