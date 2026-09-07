package operations

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/finding"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// persistFindingInput upserts one Finding record through the catalog-driven entityops upsert
func persistFindingInput(ctx context.Context, db *ent.Client, installation *ent.Integration, createInput ent.CreateFindingInput) (string, error) {
	if createInput.Description != nil && *createInput.Description != "" {
		normalized := normalizeDescription(*createInput.Description)
		createInput.Description = &normalized
	}

	id, _, err := persistCatalogUpsert(ctx, db, entityops.SchemaFinding, installation.OwnerID, createInput)
	if err != nil {
		return "", err
	}

	if err := ensureIngestIntegrationLink(ctx, entityops.SchemaFinding.Snake, id, installation.ID,
		func(ctx context.Context) ([]string, error) {
			return db.Finding.Query().Where(finding.HasIntegrationsWith(integration.ID(installation.ID))).IDs(ctx)
		},
		func(ctx context.Context) (bool, error) {
			return db.Finding.Query().Where(finding.ID(id), finding.HasIntegrationsWith(integration.ID(installation.ID))).Exist(ctx)
		},
		func(ctx context.Context) error {
			return db.Finding.UpdateOneID(id).AddIntegrationIDs(installation.ID).Exec(ctx)
		},
	); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("finding integration link failed")

		return id, wrapIngestPersistError(err)
	}

	return id, nil
}
