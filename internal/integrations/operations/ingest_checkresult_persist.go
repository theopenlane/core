package operations

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/checkresult"
)

// persistCheckResultInput upserts one CheckResult record using the ingest lookup key fields
func persistCheckResultInput(ctx context.Context, db *ent.Client, integration *ent.Integration, createInput ent.CreateCheckResultInput) (string, bool, bool, error) {
	if createInput.ParentExternalID == nil {
		return "", false, false, ErrIngestUpsertKeyMissing
	}

	if createInput.Source == "" && integration.Name != "" {
		createInput.Source = integration.Name
	}

	if createInput.IntegrationID == nil {
		createInput.IntegrationID = &integration.ID
	}

	q := db.CheckResult.Query().
		Where(
			checkresult.ParentExternalID(*createInput.ParentExternalID),
			checkresult.IntegrationID(*createInput.IntegrationID),
		)

	return persistLookupUpsert(ctx, db, entityops.SchemaCheckResult, integration, createInput, q.Only)
}
