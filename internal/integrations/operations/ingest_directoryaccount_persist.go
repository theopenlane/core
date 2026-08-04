package operations

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
)

// persistDirectoryAccountInput upserts one DirectoryAccount record using the ingest lookup key fields
func persistDirectoryAccountInput(ctx context.Context, db *ent.Client, integrationDef *ent.Integration, createInput ent.CreateDirectoryAccountInput) (string, error) {
	if createInput.ExternalID == "" {
		return "", ErrIngestUpsertKeyMissing
	}

	createInput.PrimarySource = &integrationDef.PrimaryDirectory

	if createInput.DirectoryName == nil && integrationDef.Name != "" {
		createInput.DirectoryName = &integrationDef.Name
	}

	now := time.Now()

	// prefer directory instance so recreating integrations does
	// not create a new directory account record, fall back to integration id
	lookup := func(externalID string) []predicate.DirectoryAccount {
		where := []predicate.DirectoryAccount{
			directoryaccount.OwnerID(*createInput.OwnerID),
			directoryaccount.ExternalID(externalID),
		}

		if createInput.DirectoryInstanceID != nil && *createInput.DirectoryInstanceID != "" {
			return append(where, directoryaccount.DirectoryInstanceID(*createInput.DirectoryInstanceID))
		}

		return append(where, directoryaccount.IntegrationID(*createInput.IntegrationID))
	}

	return persistRoundTripUpsert(
		ctx,
		createInput,
		func(ctx context.Context) (*ent.DirectoryAccount, error) {
			return findWithLegacyKeyAdoption(ctx, createInput.ExternalID,
				func(ctx context.Context, externalID string) (*ent.DirectoryAccount, error) {
					return db.DirectoryAccount.Query().
						Where(lookup(externalID)...).
						Only(ctx)
				},
				// the row still carries the old scientific notation key, so fix it in place
				// before the update proceeds (Modify because external_id is immutable)
				func(ctx context.Context, account *ent.DirectoryAccount) error {
					return db.DirectoryAccount.UpdateOneID(account.ID).
						Modify(func(u *sql.UpdateBuilder) {
							u.Set(directoryaccount.FieldExternalID, createInput.ExternalID)
						}).
						Exec(ctx)
				},
			)
		},
		func(ctx context.Context, input ent.CreateDirectoryAccountInput) (string, error) {
			input.FirstSeenAt = &now
			da, err := db.DirectoryAccount.Create().SetInput(input).Save(ctx)
			if err != nil {
				return "", err
			}
			return da.ID, nil
		},
		func(ctx context.Context, existing *ent.DirectoryAccount, input ent.UpdateDirectoryAccountInput) error {
			input.LastSeenAt = &now
			return db.DirectoryAccount.UpdateOneID(existing.ID).SetInput(input).Exec(ctx)
		},
		func(da *ent.DirectoryAccount) string { return da.ID },
	)
}
