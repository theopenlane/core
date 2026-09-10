package operations

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/samber/lo"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
)

// persistDirectoryGroupInput upserts one DirectoryGroup record using the ingest lookup key fields
func persistDirectoryGroupInput(ctx context.Context, db *ent.Client, integration *ent.Integration, createInput ent.CreateDirectoryGroupInput) (string, error) {
	if createInput.ExternalID == "" {
		return "", ErrIngestUpsertKeyMissing
	}

	if createInput.DirectoryName == nil && integration.Name != "" {
		createInput.DirectoryName = &integration.Name
	}

	if hash := directoryProfileHash(createInput.Profile); hash != "" {
		createInput.ProfileHash = &hash
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
			if directoryGroupUnchanged(existing, input) {
				return nil
			}

			return db.DirectoryGroup.UpdateOneID(existing.ID).SetInput(input).Exec(ctx)
		},
		func(dg *ent.DirectoryGroup) string { return dg.ID },
	)
}

// directoryGroupUnchanged reports whether the incoming update carries no payload or
// integration-derived changes for the existing row
func directoryGroupUnchanged(existing *ent.DirectoryGroup, input ent.UpdateDirectoryGroupInput) bool {
	if input.ProfileHash == nil || *input.ProfileHash == "" || existing.ProfileHash != *input.ProfileHash {
		return false
	}

	return input.DirectoryName == nil || lo.FromPtr(existing.DirectoryName) == *input.DirectoryName
}
