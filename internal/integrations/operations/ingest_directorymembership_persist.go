package operations

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/samber/lo"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/pkg/logx"
)

// persistDirectoryMembershipInput upserts one DirectoryMembership record using the ingest lookup key fields
func persistDirectoryMembershipInput(ctx context.Context, db *ent.Client, integration *ent.Integration, createInput ent.CreateDirectoryMembershipInput) (string, error) {
	if createInput.DirectoryAccountID == "" || createInput.DirectoryGroupID == "" {
		return "", ErrIngestUpsertKeyMissing
	}

	resolvedInput, err := resolveDirectoryMembershipInput(ctx, db, integration, createInput)
	if err != nil {
		return "", err
	}

	now := time.Now()
	runID := directorySyncRunIDFromContext(ctx)

	// removed memberships are excluded from the lookup so a re-added membership starts a
	// new episode row attributed to the sync run that observed the re-add
	return persistRoundTripUpsert(
		ctx,
		resolvedInput,
		func(ctx context.Context) (*ent.DirectoryMembership, error) {
			return db.DirectoryMembership.Query().
				Where(directorymembership.IntegrationID(integration.ID)).
				Where(directorymembership.DirectoryAccountID(resolvedInput.DirectoryAccountID)).
				Where(directorymembership.DirectoryGroupID(resolvedInput.DirectoryGroupID)).
				Where(directorymembership.RemovedAtIsNil()).
				Only(ctx)
		},
		func(ctx context.Context, input ent.CreateDirectoryMembershipInput) (string, error) {
			input.FirstSeenAt = &now
			input.LastSeenAt = &now
			if runID != "" {
				input.LastConfirmedRunID = &runID
			}

			dm, err := db.DirectoryMembership.Create().SetInput(input).Save(ctx)
			if err != nil {
				return "", err
			}
			return dm.ID, nil
		},
		func(ctx context.Context, existing *ent.DirectoryMembership, input ent.UpdateDirectoryMembershipInput) error {
			input.LastSeenAt = &now
			if runID != "" {
				input.LastConfirmedRunID = &runID
			}

			return db.DirectoryMembership.UpdateOneID(existing.ID).SetInput(input).Exec(ctx)
		},
		func(dm *ent.DirectoryMembership) string { return dm.ID },
	)
}

// resolveDirectoryMembershipInput normalizes provider lookup values into internal record IDs before persistence
func resolveDirectoryMembershipInput(ctx context.Context, db *ent.Client, integration *ent.Integration, input ent.CreateDirectoryMembershipInput) (ent.CreateDirectoryMembershipInput, error) {
	accountID, err := resolveDirectoryAccountID(ctx, db, integration, lo.FromPtr(input.DirectoryInstanceID), input.DirectoryAccountID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("account_ref", input.DirectoryAccountID).Str("group_ref", input.DirectoryGroupID).Msg("unresolved directory account for membership")

		return input, err
	}

	if input.DirectoryName == nil && integration.Name != "" {
		input.DirectoryName = &integration.Name
	}

	groupID, err := resolveDirectoryGroupID(ctx, db, integration, input.DirectoryGroupID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("account_ref", input.DirectoryAccountID).Str("group_ref", input.DirectoryGroupID).Msg("unresolved directory group for membership")

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

	scope := directoryaccount.IntegrationID(integration.ID)
	if instanceID != "" {
		scope = directoryaccount.DirectoryInstanceID(instanceID)
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
func resolveDirectoryGroupID(ctx context.Context, db *ent.Client, integration *ent.Integration, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrIngestUpsertKeyMissing
	}

	group, err := db.DirectoryGroup.Query().
		Where(directorygroup.ID(value), directorygroup.IntegrationID(integration.ID)).
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
		Where(directorygroup.IntegrationID(integration.ID)).
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
