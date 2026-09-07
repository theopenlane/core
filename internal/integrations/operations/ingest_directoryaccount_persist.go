package operations

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// persistDirectoryAccountInput upserts one DirectoryAccount record using the ingest lookup key fields
func persistDirectoryAccountInput(ctx context.Context, db *ent.Client, integrationDef *ent.Integration, createInput ent.CreateDirectoryAccountInput) (string, error) {
	if createInput.ExternalID == "" {
		logx.FromContext(ctx).Error().Err(ErrIngestUpsertKeyMissing).Msg("directory account ingest missing external id")

		return "", ErrIngestUpsertKeyMissing
	}

	createInput.PrimarySource = &integrationDef.PrimaryDirectory

	if hash := directoryProfileHash(createInput.Profile); hash != "" {
		createInput.ProfileHash = &hash
	}

	now := time.Now()

	existing, found, err := findDirectoryAccountForIngest(ctx, db, createInput)
	if err != nil {
		if errors.Is(err, ErrIngestUpsertConflict) {
			return "", err
		}

		return "", wrapIngestPersistError(err)
	}

	if !found {
		createInput.FirstSeenAt = &now

		if createInput.DirectoryName == nil && integrationDef.Name != "" {
			createInput.DirectoryName = &integrationDef.Name
		}

		da, createErr := db.DirectoryAccount.Create().SetInput(createInput).Save(ctx)
		if createErr != nil {
			logx.FromContext(ctx).Error().Err(createErr).Msg("directory account create failed")

			return "", wrapIngestPersistError(createErr)
		}

		if batch := directorySyncBatchFromContext(ctx); batch != nil {
			batch.addAccount(da, lookupScopeKey(lo.FromPtr(createInput.OwnerID), lo.FromPtr(createInput.SourceInstanceID), lo.FromPtr(createInput.IntegrationID)))
		}

		recordIngestChange(ctx)

		return da.ID, nil
	}

	updateInput, err := roundTripUpdateInput[ent.CreateDirectoryAccountInput, ent.UpdateDirectoryAccountInput](createInput)
	if err != nil {
		return "", err
	}

	if entityops.DirectoryAccountIngestUnchanged(existing, updateInput) {
		if existing.IntegrationID != lo.FromPtr(createInput.IntegrationID) {
			if err := relinkIngestIntegration(ctx, db, entityops.SchemaDirectoryAccount.Snake, existing.ID, lo.FromPtr(createInput.IntegrationID)); err != nil {
				return existing.ID, wrapIngestPersistError(err)
			}
		}

		if batch := directorySyncBatchFromContext(ctx); batch != nil {
			batch.seenAccountIDs = append(batch.seenAccountIDs, existing.ID)

			return existing.ID, nil
		}

		if err := db.DirectoryAccount.UpdateOneID(existing.ID).
			SetLastSeenAt(now).
			Exec(entityops.WithEmissionVetoed(ctx)); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("directory account last seen update failed")

			return existing.ID, wrapIngestPersistError(err)
		}

		return existing.ID, nil
	}

	updateInput.LastSeenAt = &now

	if err := db.DirectoryAccount.UpdateOneID(existing.ID).SetInput(updateInput).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory account update failed")

		return existing.ID, wrapIngestPersistError(err)
	}

	recordIngestChange(ctx)

	return existing.ID, nil
}

// directoryAccountLookupPredicates prefers the directory instance, keeping the integration's not-yet-stamped rows findable
func directoryAccountLookupPredicates(createInput ent.CreateDirectoryAccountInput, externalID string) []predicate.DirectoryAccount {
	where := []predicate.DirectoryAccount{
		directoryaccount.OwnerID(*createInput.OwnerID),
		directoryaccount.ExternalID(externalID),
	}

	if createInput.SourceInstanceID != nil && *createInput.SourceInstanceID != "" {
		return append(where, directoryaccount.Or(
			directoryaccount.SourceInstanceID(*createInput.SourceInstanceID),
			directoryaccount.And(directoryaccount.IntegrationID(*createInput.IntegrationID), directoryaccount.SourceInstanceIDIsNil()),
		))
	}

	return append(where, directoryaccount.IntegrationID(*createInput.IntegrationID))
}

// findDirectoryAccountForIngest finds the existing row via the batch cache, falling back to the live lookup; found reports whether a row exists
func findDirectoryAccountForIngest(ctx context.Context, db *ent.Client, createInput ent.CreateDirectoryAccountInput) (*ent.DirectoryAccount, bool, error) {
	batch := directorySyncBatchFromContext(ctx)
	if batch == nil {
		return findDirectoryAccountLive(ctx, db, createInput)
	}

	scope, err := batch.accountScope(ctx, db, lookupScopeKey(lo.FromPtr(createInput.OwnerID), lo.FromPtr(createInput.SourceInstanceID), lo.FromPtr(createInput.IntegrationID)))
	if err != nil {
		return nil, false, err
	}

	if rows := scope.byExternalID[createInput.ExternalID]; len(rows) > 0 {
		return chooseDirectoryAccountRow(ctx, rows, createInput)
	}

	legacy, ok := legacyScientificKey(createInput.ExternalID)
	if !ok {
		return nil, false, nil
	}

	rows := scope.byExternalID[legacy]
	if len(rows) == 0 {
		return nil, false, nil
	}

	row, found, err := chooseDirectoryAccountRow(ctx, rows, createInput)
	if err != nil || !found {
		return nil, found, err
	}

	if err := repairDirectoryAccountExternalID(ctx, db, row.ID, createInput.ExternalID); err != nil {
		return nil, false, err
	}

	scope.byExternalID[createInput.ExternalID] = append(scope.byExternalID[createInput.ExternalID], row)

	return row, true, nil
}

// chooseDirectoryAccountRow selects the account row a scoped lookup should use
func chooseDirectoryAccountRow(ctx context.Context, rows []*ent.DirectoryAccount, createInput ent.CreateDirectoryAccountInput) (*ent.DirectoryAccount, bool, error) {
	row, err := chooseScopedDirectoryRow(rows, func(r *ent.DirectoryAccount) bool {
		return r.IntegrationID == lo.FromPtr(createInput.IntegrationID)
	}, newestDirectoryAccount)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory account lookup conflict")

		return nil, false, err
	}

	return row, true, nil
}

// findDirectoryAccountLive runs the per-record lookup with legacy key adoption; found reports
// whether a row exists
func findDirectoryAccountLive(ctx context.Context, db *ent.Client, createInput ent.CreateDirectoryAccountInput) (*ent.DirectoryAccount, bool, error) {
	rows, err := db.DirectoryAccount.Query().
		Where(directoryAccountLookupPredicates(createInput, createInput.ExternalID)...).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory account lookup failed")

		return nil, false, err
	}

	if len(rows) > 0 {
		return chooseDirectoryAccountRow(ctx, rows, createInput)
	}

	legacy, ok := legacyScientificKey(createInput.ExternalID)
	if !ok {
		return nil, false, nil
	}

	rows, err = db.DirectoryAccount.Query().
		Where(directoryAccountLookupPredicates(createInput, legacy)...).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory account lookup failed")

		return nil, false, err
	}

	if len(rows) == 0 {
		return nil, false, nil
	}

	row, found, err := chooseDirectoryAccountRow(ctx, rows, createInput)
	if err != nil || !found {
		return nil, found, err
	}

	if err := repairDirectoryAccountExternalID(ctx, db, row.ID, createInput.ExternalID); err != nil {
		return nil, false, err
	}

	return row, true, nil
}

// repairDirectoryAccountExternalID rewrites a legacy scientific notation key to the canonical form (Modify because external_id is immutable)
func repairDirectoryAccountExternalID(ctx context.Context, db *ent.Client, accountID string, externalID string) error {
	if err := db.DirectoryAccount.UpdateOneID(accountID).
		Modify(func(u *sql.UpdateBuilder) {
			u.Set(directoryaccount.FieldExternalID, externalID)
		}).
		Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory account legacy key repair failed")

		return err
	}

	return nil
}

// directoryProfileHash returns the canonical hash of the normalized profile payload, or empty when
// the mapping carries no profile, which disables hash-based change detection for the record
func directoryProfileHash(profile map[string]any) string {
	if len(profile) == 0 {
		return ""
	}

	raw, err := json.Marshal(profile)
	if err != nil {
		return ""
	}

	sum := sha256.Sum256(raw)

	return hex.EncodeToString(sum[:])
}
