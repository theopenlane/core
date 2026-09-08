package operations

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/internalpolicy"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// persistInternalPolicyInput upserts one InternalPolicy record through the catalog-driven entityops upsert
func persistInternalPolicyInput(ctx context.Context, db *ent.Client, installation *ent.Integration, createInput ent.CreateInternalPolicyInput) (string, bool, bool, error) {
	id, changed, managed, err := persistCatalogUpsert(ctx, db, entityops.SchemaInternalPolicy, installation.OwnerID, installation, createInput)
	if err != nil {
		return "", false, false, err
	}

	if !managed {
		return id, changed, managed, nil
	}

	if err := ensureIngestIntegrationLink(ctx, entityops.SchemaInternalPolicy.Snake, id, installation.ID,
		func(ctx context.Context) ([]string, error) {
			return db.InternalPolicy.Query().Where(internalpolicy.HasIntegrationsWith(integration.ID(installation.ID))).IDs(ctx)
		},
		func(ctx context.Context) (bool, error) {
			return db.InternalPolicy.Query().Where(internalpolicy.ID(id), internalpolicy.HasIntegrationsWith(integration.ID(installation.ID))).Exist(ctx)
		},
		func(ctx context.Context) error {
			return db.InternalPolicy.UpdateOneID(id).AddIntegrationIDs(installation.ID).Exec(ctx)
		},
	); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("internal policy integration link failed")

		return id, changed, managed, wrapIngestPersistError(err)
	}

	return id, changed, managed, nil
}
