package operations

import (
	"context"

	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
)

// persistActionPlanInput upserts one ActionPlan record through the catalog-driven entityops upsert
func persistActionPlanInput(ctx context.Context, db *ent.Client, integration *ent.Integration, createInput ent.CreateActionPlanInput) (string, error) {
	return persistCatalogUpsert(ctx, db, entityops.SchemaActionPlan, integration.OwnerID, createInput)
}
