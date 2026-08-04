package operations

import (
	"context"

	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
)

// persistRiskInput upserts one Risk record through the catalog-driven entityops upsert
func persistRiskInput(ctx context.Context, db *ent.Client, integration *ent.Integration, createInput ent.CreateRiskInput) (string, error) {
	if createInput.IntegrationID == nil {
		createInput.IntegrationID = &integration.ID
	}

	return persistCatalogUpsert(ctx, db, entityops.SchemaRisk, integration.OwnerID, createInput)
}
