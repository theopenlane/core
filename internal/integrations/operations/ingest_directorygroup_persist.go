package operations

import (
	"context"
	"errors"

	"entgo.io/ent/dialect/sql"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// persistDirectoryGroupInput upserts one DirectoryGroup record using the ingest lookup key fields
func persistDirectoryGroupInput(ctx context.Context, db *ent.Client, integration *ent.Integration, createInput ent.CreateDirectoryGroupInput) (string, bool, bool, error) {
	if createInput.ExternalID == "" {
		logx.FromContext(ctx).Error().Err(ErrIngestUpsertKeyMissing).Msg("directory group ingest missing external id")

		return "", false, false, ErrIngestUpsertKeyMissing
	}

	createInput, err := stampDirectoryProvenance(createInput, integration)
	if err != nil {
		return "", false, false, wrapIngestPersistError(err)
	}

	if hash := directoryProfileHash(createInput.Profile); hash != "" {
		createInput.ProfileHash = &hash
	}

	existing, found, err := findDirectoryGroupForIngest(ctx, db, createInput)
	if err != nil {
		if errors.Is(err, ErrIngestUpsertConflict) {
			return "", false, false, err
		}

		return "", false, false, wrapIngestPersistError(err)
	}

	if !found {
		if createInput.DirectoryName == nil && integration.Name != "" {
			createInput.DirectoryName = &integration.Name
		}

		dg, createErr := db.DirectoryGroup.Create().SetInput(createInput).Save(ctx)
		if createErr != nil {
			logx.FromContext(ctx).Error().Err(createErr).Msg("directory group create failed")

			return "", false, false, wrapIngestPersistError(createErr)
		}

		if batch := directorySyncBatchFromContext(ctx); batch != nil {
			batch.addGroup(dg, lookupScopeKey(lo.FromPtr(createInput.OwnerID), lo.FromPtr(createInput.SourceInstanceID), createInput.IntegrationID))
		}

		recordIngestChange(ctx)

		return dg.ID, true, true, nil
	}

	changes, err := directoryChangeSet(ctx, db, entityops.SchemaDirectoryGroup, existing, createInput)
	if err != nil {
		return "", false, false, wrapIngestPersistError(err)
	}

	if changes.Empty() {
		if existing.IntegrationID != createInput.IntegrationID {
			if err := relinkIngestIntegration(ctx, db, entityops.SchemaDirectoryGroup.Snake, existing.ID, createInput.IntegrationID); err != nil {
				return existing.ID, false, false, wrapIngestPersistError(err)
			}
		}

		return existing.ID, false, true, nil
	}

	updateInput, err := roundTripUpdateInput[ent.CreateDirectoryGroupInput, ent.UpdateDirectoryGroupInput](createInput)
	if err != nil {
		return "", false, false, err
	}

	if err := db.DirectoryGroup.UpdateOneID(existing.ID).SetInput(updateInput).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory group update failed")

		return existing.ID, false, false, wrapIngestPersistError(err)
	}

	recordIngestChange(ctx)

	return existing.ID, true, true, nil
}

// directoryGroupLookupPredicates prefers the directory instance, keeping the integration's not-yet-stamped rows findable
func directoryGroupLookupPredicates(createInput ent.CreateDirectoryGroupInput, externalID string) []predicate.DirectoryGroup {
	where := []predicate.DirectoryGroup{
		directorygroup.OwnerID(*createInput.OwnerID),
		directorygroup.ExternalID(externalID),
	}

	if createInput.SourceInstanceID != nil && *createInput.SourceInstanceID != "" {
		return append(where, directorygroup.Or(
			directorygroup.SourceInstanceID(*createInput.SourceInstanceID),
			directorygroup.And(directorygroup.IntegrationID(createInput.IntegrationID), directorygroup.SourceInstanceIDIsNil()),
		))
	}

	return append(where, directorygroup.IntegrationID(createInput.IntegrationID))
}

// findDirectoryGroupForIngest finds the existing row via the batch cache, falling back to the live lookup; found reports whether a row exists
func findDirectoryGroupForIngest(ctx context.Context, db *ent.Client, createInput ent.CreateDirectoryGroupInput) (*ent.DirectoryGroup, bool, error) {
	batch := directorySyncBatchFromContext(ctx)
	if batch == nil {
		return findDirectoryGroupLive(ctx, db, createInput)
	}

	scope, err := batch.groupScope(ctx, db, lookupScopeKey(lo.FromPtr(createInput.OwnerID), lo.FromPtr(createInput.SourceInstanceID), createInput.IntegrationID))
	if err != nil {
		return nil, false, err
	}

	if rows := scope.byExternalID[createInput.ExternalID]; len(rows) > 0 {
		return chooseDirectoryGroupRow(ctx, rows, createInput)
	}

	legacy, ok := legacyScientificKey(createInput.ExternalID)
	if !ok {
		return nil, false, nil
	}

	rows := scope.byExternalID[legacy]
	if len(rows) == 0 {
		return nil, false, nil
	}

	row, found, err := chooseDirectoryGroupRow(ctx, rows, createInput)
	if err != nil || !found {
		return nil, found, err
	}

	if err := repairDirectoryGroupExternalID(ctx, db, row.ID, createInput.ExternalID); err != nil {
		return nil, false, err
	}

	scope.byExternalID[createInput.ExternalID] = append(scope.byExternalID[createInput.ExternalID], row)

	return row, true, nil
}

// chooseDirectoryGroupRow selects the group row a scoped lookup should use
func chooseDirectoryGroupRow(ctx context.Context, rows []*ent.DirectoryGroup, createInput ent.CreateDirectoryGroupInput) (*ent.DirectoryGroup, bool, error) {
	row, err := chooseScopedDirectoryRow(rows, func(r *ent.DirectoryGroup) bool {
		return r.IntegrationID == createInput.IntegrationID
	}, newestDirectoryGroup)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory group lookup conflict")

		return nil, false, err
	}

	return row, true, nil
}

// findDirectoryGroupLive runs the per-record lookup with legacy key adoption; found reports
// whether a row exists
func findDirectoryGroupLive(ctx context.Context, db *ent.Client, createInput ent.CreateDirectoryGroupInput) (*ent.DirectoryGroup, bool, error) {
	rows, err := db.DirectoryGroup.Query().
		Where(directoryGroupLookupPredicates(createInput, createInput.ExternalID)...).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory group lookup failed")

		return nil, false, err
	}

	if len(rows) > 0 {
		return chooseDirectoryGroupRow(ctx, rows, createInput)
	}

	legacy, ok := legacyScientificKey(createInput.ExternalID)
	if !ok {
		return nil, false, nil
	}

	rows, err = db.DirectoryGroup.Query().
		Where(directoryGroupLookupPredicates(createInput, legacy)...).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory group lookup failed")

		return nil, false, err
	}

	if len(rows) == 0 {
		return nil, false, nil
	}

	row, found, err := chooseDirectoryGroupRow(ctx, rows, createInput)
	if err != nil || !found {
		return nil, found, err
	}

	if err := repairDirectoryGroupExternalID(ctx, db, row.ID, createInput.ExternalID); err != nil {
		return nil, false, err
	}

	return row, true, nil
}

// repairDirectoryGroupExternalID rewrites a legacy scientific notation key to the canonical form (Modify because external_id is immutable)
func repairDirectoryGroupExternalID(ctx context.Context, db *ent.Client, groupID string, externalID string) error {
	if err := db.DirectoryGroup.UpdateOneID(groupID).
		Modify(func(u *sql.UpdateBuilder) {
			u.Set(directorygroup.FieldExternalID, externalID)
		}).
		Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory group legacy key repair failed")

		return err
	}

	return nil
}
