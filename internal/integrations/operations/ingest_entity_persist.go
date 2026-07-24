package operations

import (
	"context"
	"slices"

	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
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

	if createInput.EntitySourceTypeName == nil && integration.Kind != "" {
		createInput.EntitySourceTypeName = &integration.Kind
	}

	if integration.ID != "" && !slices.Contains(createInput.IntegrationIDs, integration.ID) {
		createInput.IntegrationIDs = append(createInput.IntegrationIDs, integration.ID)
	}

	return persistCatalogUpsert(ctx, db, entityops.SchemaEntity, ownerID, createInput)
}
