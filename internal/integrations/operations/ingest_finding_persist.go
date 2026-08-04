package operations

import (
	"context"
	"slices"

	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
)

// persistFindingInput upserts one Finding record through the catalog-driven entityops upsert
func persistFindingInput(ctx context.Context, db *ent.Client, integration *ent.Integration, createInput ent.CreateFindingInput) (string, error) {
	if createInput.Source == nil && integration.Name != "" {
		createInput.Source = &integration.Name
	}

	if createInput.Description != nil && *createInput.Description != "" {
		normalized := normalizeDescription(*createInput.Description)
		createInput.Description = &normalized
	}

	if !slices.Contains(createInput.IntegrationIDs, integration.ID) {
		createInput.IntegrationIDs = append(createInput.IntegrationIDs, integration.ID)
	}

	return persistCatalogUpsert(ctx, db, entityops.SchemaFinding, integration.OwnerID, createInput)
}
