package operations

import (
	"context"
	"slices"

	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
)

// persistInternalPolicyInput upserts one InternalPolicy record through the catalog-driven entityops upsert
func persistInternalPolicyInput(ctx context.Context, db *ent.Client, integration *ent.Integration, createInput ent.CreateInternalPolicyInput) (string, error) {
	if integration.ID != "" && !slices.Contains(createInput.IntegrationIDs, integration.ID) {
		createInput.IntegrationIDs = append(createInput.IntegrationIDs, integration.ID)
	}

	return persistCatalogUpsert(ctx, db, entityops.SchemaInternalPolicy, integration.OwnerID, createInput)
}
