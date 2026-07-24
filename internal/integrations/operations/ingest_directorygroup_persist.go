package operations

import (
	"context"

	"entgo.io/ent/dialect/sql"

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
			return findWithLegacyKeyAdoption(ctx, createInput.ExternalID,
				func(ctx context.Context, externalID string) (*ent.DirectoryGroup, error) {
					return db.DirectoryGroup.Query().
						Where(directorygroup.IntegrationID(integration.ID)).
						Where(directorygroup.ExternalID(externalID)).
						Only(ctx)
				},
				// the row still carries the old scientific notation key, so fix it in place
				// before the update proceeds (Modify because external_id is immutable)
				func(ctx context.Context, group *ent.DirectoryGroup) error {
					return db.DirectoryGroup.UpdateOneID(group.ID).
						Modify(func(u *sql.UpdateBuilder) {
							u.Set(directorygroup.FieldExternalID, createInput.ExternalID)
						}).
						Exec(ctx)
				},
			)
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
