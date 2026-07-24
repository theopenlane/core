package operations

import (
	"context"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directorygroup"
)

// persistDirectoryGroupInput upserts one DirectoryGroup record using the ingest lookup key fields
func persistDirectoryGroupInput(ctx context.Context, db *ent.Client, integration *ent.Integration, createInput ent.CreateDirectoryGroupInput) (string, error) {
	if createInput.ExternalID == "" {
		return "", ErrIngestUpsertKeyMissing
	}

	if createInput.DirectoryName == nil && integration.Name != "" {
		createInput.DirectoryName = &integration.Name
	}

	return persistRoundTripUpsert(
		ctx,
		createInput,
		func(ctx context.Context) (*ent.DirectoryGroup, error) {
			return db.DirectoryGroup.Query().
				Where(directorygroup.IntegrationID(integration.ID)).
				Where(directorygroup.ExternalID(createInput.ExternalID)).
				Only(ctx)
		},
		func(ctx context.Context, input ent.CreateDirectoryGroupInput) (string, error) {
			dg, err := db.DirectoryGroup.Create().SetInput(input).Save(ctx)
			if err != nil {
				return "", err
			}
			return dg.ID, nil
		},
		func(ctx context.Context, existing *ent.DirectoryGroup, input ent.UpdateDirectoryGroupInput) error {
			return db.DirectoryGroup.UpdateOneID(existing.ID).SetInput(input).Exec(ctx)
		},
		func(dg *ent.DirectoryGroup) string { return dg.ID },
	)
}
