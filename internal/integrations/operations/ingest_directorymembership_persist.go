package operations

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// persistDirectoryMembershipInput upserts one DirectoryMembership record using the ingest lookup key fields
func persistDirectoryMembershipInput(ctx context.Context, db *ent.Client, integration *ent.Integration, createInput ent.CreateDirectoryMembershipInput) (string, bool, bool, error) {
	if createInput.DirectoryAccountID == "" || createInput.DirectoryGroupID == "" {
		logx.FromContext(ctx).Error().Err(ErrIngestUpsertKeyMissing).Msg("directory membership ingest missing account or group reference")

		return "", false, false, ErrIngestUpsertKeyMissing
	}

	createInput, err := stampDirectoryProvenance(createInput, integration)
	if err != nil {
		return "", false, false, wrapIngestPersistError(err)
	}

	resolvedInput, err := resolveDirectoryMembershipInput(ctx, db, integration, createInput)
	if err != nil {
		return "", false, false, err
	}

	now := time.Now()
	runID := directorySyncRunIDFromContext(ctx)

	existing, found, err := findDirectoryMembershipForIngest(ctx, db, integration.OwnerID, resolvedInput)
	if err != nil {
		return "", false, false, wrapIngestPersistError(err)
	}

	if !found {
		resolvedInput.FirstSeenAt = &now
		resolvedInput.LastSeenAt = &now

		if resolvedInput.DirectoryName == nil && integration.Name != "" {
			resolvedInput.DirectoryName = &integration.Name
		}

		if runID != "" {
			resolvedInput.LastConfirmedRunID = &runID
		}

		dm, createErr := db.DirectoryMembership.Create().SetInput(resolvedInput).Save(ctx)
		if createErr != nil {
			logx.FromContext(ctx).Error().Err(createErr).Msg("directory membership create failed")

			return "", false, false, wrapIngestPersistError(createErr)
		}

		if batch := directorySyncBatchFromContext(ctx); batch != nil {
			batch.addMembership(dm)
		}

		recordIngestChange(ctx)

		return dm.ID, true, true, nil
	}

	input, err := roundTripUpdateInput[ent.CreateDirectoryMembershipInput, ent.UpdateDirectoryMembershipInput](resolvedInput)
	if err != nil {
		return "", false, false, err
	}

	changes, err := directoryChangeSet(ctx, db, entityops.SchemaDirectoryMembership, existing, resolvedInput)
	if err != nil {
		return "", false, false, wrapIngestPersistError(err)
	}

	if changes.Empty() && directoryMembershipRunCanAdvance(existing.LastConfirmedRunID, runID) {
		if batch := directorySyncBatchFromContext(ctx); batch != nil {
			if existing.IntegrationID != resolvedInput.IntegrationID {
				if err := relinkIngestIntegration(ctx, db, entityops.SchemaDirectoryMembership.Snake, existing.ID, resolvedInput.IntegrationID); err != nil {
					return existing.ID, false, false, wrapIngestPersistError(err)
				}
			}

			batch.confirmedMembershipIDs = append(batch.confirmedMembershipIDs, existing.ID)

			return existing.ID, false, true, nil
		}
	}

	if err := updateDirectoryMembership(ctx, db, existing, input, runID, now); err != nil {
		return existing.ID, false, false, wrapIngestPersistError(err)
	}

	return existing.ID, true, true, nil
}

// updateDirectoryMembership applies one membership update, advancing bookkeeping only when this run can still confirm the row
func updateDirectoryMembership(ctx context.Context, db *ent.Client, existing *ent.DirectoryMembership, input ent.UpdateDirectoryMembershipInput, runID string, now time.Time) error {
	if runID == "" {
		input.LastSeenAt = &now

		if err := db.DirectoryMembership.UpdateOneID(existing.ID).SetInput(input).Exec(ctx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("directory membership update failed")

			return err
		}

		recordIngestChange(ctx)

		return nil
	}

	if !directoryMembershipRunCanAdvance(existing.LastConfirmedRunID, runID) {
		if err := db.DirectoryMembership.UpdateOneID(existing.ID).SetInput(input).Exec(ctx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("directory membership update failed")

			return err
		}

		recordIngestChange(ctx)

		return nil
	}

	input.LastSeenAt = &now
	input.LastConfirmedRunID = &runID
	update := db.DirectoryMembership.UpdateOneID(existing.ID).
		SetInput(input).
		Where(directorymembership.Or(
			directorymembership.LastConfirmedRunIDIsNil(),
			directorymembership.LastConfirmedRunIDLTE(runID),
		))
	err := update.Exec(ctx)

	switch {
	case err == nil:
		recordIngestChange(ctx)

		return nil
	case !ent.IsNotFound(err):
		logx.FromContext(ctx).Error().Err(err).Msg("directory membership update failed")

		return err
	}

	// A newer run won the guarded update. Preserve unrelated field changes,
	// but do not move confirmation or last-seen timestamps backward.
	input.LastSeenAt = nil
	input.LastConfirmedRunID = nil

	if err := db.DirectoryMembership.UpdateOneID(existing.ID).SetInput(input).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory membership update failed")

		return err
	}

	recordIngestChange(ctx)

	return nil
}

// findDirectoryMembershipForIngest finds the active row via the batch cache, falling back to the live lookup; found reports whether a row exists
func findDirectoryMembershipForIngest(ctx context.Context, db *ent.Client, ownerID string, input ent.CreateDirectoryMembershipInput) (*ent.DirectoryMembership, bool, error) {
	batch := directorySyncBatchFromContext(ctx)
	if batch == nil {
		return findDirectoryMembershipLive(ctx, db, input)
	}

	index, err := batch.membershipIndex(ctx, db, ownerID)
	if err != nil {
		return nil, false, err
	}

	rows := index[membershipKey{accountID: input.DirectoryAccountID, groupID: input.DirectoryGroupID}]

	switch len(rows) {
	case 0:
		return nil, false, nil
	case 1:
		return rows[0], true, nil
	default:
		return findDirectoryMembershipLive(ctx, db, input)
	}
}

// findDirectoryMembershipLive matches the active-pair uniqueness domain, excluding removed episode rows
func findDirectoryMembershipLive(ctx context.Context, db *ent.Client, input ent.CreateDirectoryMembershipInput) (*ent.DirectoryMembership, bool, error) {
	existing, err := db.DirectoryMembership.Query().
		Where(directorymembership.DirectoryAccountID(input.DirectoryAccountID)).
		Where(directorymembership.DirectoryGroupID(input.DirectoryGroupID)).
		Where(directorymembership.RemovedAtIsNil()).
		Only(ctx)

	switch {
	case err == nil:
		return existing, true, nil
	case ent.IsNotFound(err):
		return nil, false, nil
	default:
		logx.FromContext(ctx).Error().Err(err).Msg("directory membership lookup failed")

		return nil, false, err
	}
}

func directoryMembershipRunCanAdvance(currentRunID *string, incomingRunID string) bool {
	return incomingRunID != "" && (currentRunID == nil || incomingRunID >= *currentRunID)
}

// resolveDirectoryMembershipInput normalizes provider lookup values into internal record IDs before persistence
func resolveDirectoryMembershipInput(ctx context.Context, db *ent.Client, integration *ent.Integration, input ent.CreateDirectoryMembershipInput) (ent.CreateDirectoryMembershipInput, error) {
	ctx = logx.WithFields(ctx, map[string]any{"account_ref": input.DirectoryAccountID, "group_ref": input.DirectoryGroupID})

	accountID, err := resolveDirectoryAccountID(ctx, db, integration, lo.FromPtr(input.SourceInstanceID), input.DirectoryAccountID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unresolved directory account for membership")

		return input, err
	}

	groupID, err := resolveDirectoryGroupID(ctx, db, integration, lo.FromPtr(input.SourceInstanceID), input.DirectoryGroupID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unresolved directory group for membership")

		return input, err
	}

	input.DirectoryAccountID = accountID
	input.DirectoryGroupID = groupID

	return input, nil
}

// resolveDirectoryAccountID resolves a directory account reference to its internal ID by checking primary key, external ID, and canonical email
// Lookups are scoped the same way the account upsert is (owner + directory instance when the
// membership carries one, integration otherwise) so accounts that survived a reinstall still resolve
func resolveDirectoryAccountID(ctx context.Context, db *ent.Client, integration *ent.Integration, instanceID string, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrIngestUpsertKeyMissing
	}

	if batch := directorySyncBatchFromContext(ctx); batch != nil {
		id, ok, err := batch.resolveAccountFromCache(ctx, db, integration.OwnerID, instanceID, integration.ID, value)

		switch {
		case err != nil:
			return "", err
		case ok:
			return id, nil
		}
	}

	scope := directoryaccount.IntegrationID(integration.ID)
	if instanceID != "" {
		scope = directoryaccount.Or(
			directoryaccount.SourceInstanceID(instanceID),
			directoryaccount.And(directoryaccount.IntegrationID(integration.ID), directoryaccount.SourceInstanceIDIsNil()),
		)
	}

	account, err := db.DirectoryAccount.Query().
		Where(directoryaccount.ID(value), directoryaccount.OwnerID(integration.OwnerID)).
		Only(ctx)
	switch {
	case err == nil:
		return account.ID, nil
	case !ent.IsNotFound(err):
		return "", err
	}

	refs := []predicate.DirectoryAccount{
		directoryaccount.ExternalID(value),
		directoryaccount.CanonicalEmail(value),
	}

	// older rows may still hold the scientific notation form of the same key
	if legacy, ok := legacyScientificKey(value); ok {
		refs = append(refs, directoryaccount.ExternalID(legacy))
	}

	account, err = db.DirectoryAccount.Query().
		Where(directoryaccount.OwnerID(integration.OwnerID), scope).
		Where(directoryaccount.Or(refs...)).
		Order(directoryaccount.ByCreatedAt(sql.OrderDesc())).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", fmt.Errorf("%w: unresolved directory account reference %q", ErrIngestMappedDocumentInvalid, value)
		}

		return "", err
	}

	return account.ID, nil
}

// resolveDirectoryGroupID resolves a directory group reference to its internal ID by checking primary key, external ID, and email
func resolveDirectoryGroupID(ctx context.Context, db *ent.Client, integration *ent.Integration, instanceID string, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrIngestUpsertKeyMissing
	}

	if batch := directorySyncBatchFromContext(ctx); batch != nil {
		id, ok, err := batch.resolveGroupFromCache(ctx, db, integration.OwnerID, instanceID, integration.ID, value)

		switch {
		case err != nil:
			return "", err
		case ok:
			return id, nil
		}
	}

	scope := directorygroup.IntegrationID(integration.ID)
	if instanceID != "" {
		scope = directorygroup.Or(
			directorygroup.SourceInstanceID(instanceID),
			directorygroup.And(directorygroup.IntegrationID(integration.ID), directorygroup.SourceInstanceIDIsNil()),
		)
	}

	group, err := db.DirectoryGroup.Query().
		Where(directorygroup.ID(value), directorygroup.OwnerID(integration.OwnerID)).
		Only(ctx)
	switch {
	case err == nil:
		return group.ID, nil
	case !ent.IsNotFound(err):
		return "", err
	}

	refs := []predicate.DirectoryGroup{
		directorygroup.ExternalID(value),
		directorygroup.Email(value),
	}

	// older rows may still hold the scientific notation form of the same key
	if legacy, ok := legacyScientificKey(value); ok {
		refs = append(refs, directorygroup.ExternalID(legacy))
	}

	group, err = db.DirectoryGroup.Query().
		Where(directorygroup.OwnerID(integration.OwnerID), scope).
		Where(directorygroup.Or(refs...)).
		Order(directorygroup.ByCreatedAt(sql.OrderDesc())).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", fmt.Errorf("%w: unresolved directory group reference %q", ErrIngestMappedDocumentInvalid, value)
		}

		return "", err
	}

	return group.ID, nil
}
