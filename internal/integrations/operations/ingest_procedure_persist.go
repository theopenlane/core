package operations

import (
	"context"

	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
)

// persistProcedureInput upserts one Procedure record through the catalog-driven entityops upsert
func persistProcedureInput(ctx context.Context, db *ent.Client, integration *ent.Integration, createInput ent.CreateProcedureInput) (string, error) {
	return persistCatalogUpsert(ctx, db, entityops.SchemaProcedure, integration.OwnerID, createInput)
}
