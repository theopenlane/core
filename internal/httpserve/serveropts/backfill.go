package serveropts

import (
	"context"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	intobvs "github.com/theopenlane/core/internal/integrations/observability"
	"github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/pkg/logx"
)

// backfillBypassCaps lets the backfill write organizations and memberships without a request caller while
// skipping the org-filter, FGA, and managed-group guards the membership hooks would otherwise apply
const backfillBypassCaps = auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation | auth.CapBypassManagedGroup

// WithBackfill runs a one-time, non-blocking, config-gated, idempotent startup backfills
// use-cases for this are things a db migration can't easily handle, computed data or fields, or repairs
func WithBackfill(ctx context.Context, dbClient *ent.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if dbClient == nil || !s.Config.Settings.Backfill.Enabled {
			return
		}

		rt := s.Config.Handler.IntegrationsRuntime

		go func() {
			backfillCtx := privacy.DecisionContext(ctx, privacy.Allow)
			backfillCtx = auth.WithCaller(backfillCtx, &auth.Caller{Capabilities: backfillBypassCaps})

			BackfillDirectoryDuplicates(backfillCtx, dbClient)

			if rt != nil {
				backfillIntegrationConfiguration(backfillCtx, dbClient, rt)
				backfillReconcileLoops(backfillCtx, dbClient, rt)
			}
		}()
	})
}

// backfillIntegrationConfiguration flags connected installations whose stored user input no longer
// satisfies their definition, so the owner is told to reconnect rather than left with a cycle that
// fails on every run
func backfillIntegrationConfiguration(ctx context.Context, dbClient *ent.Client, rt *runtime.Runtime) {
	installations, err := dbClient.Integration.Query().
		Where(integration.StatusEQ(enums.IntegrationStatusConnected)).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to query connected integrations")

		return
	}

	var flagged int

	for _, installation := range installations {
		def, ok := rt.Registry().Definition(installation.DefinitionID)
		if !ok {
			continue
		}

		installCtx := intobvs.WithInstallation(ctx, installation)

		if err := rt.ValidateUserInput(installCtx, def, installation.Config.ClientConfig); err == nil {
			continue
		}

		if err := rt.MarkIntegrationUnhealthy(installCtx, installation, "it is missing required configuration"); err != nil {
			logx.FromContext(installCtx).Error().Err(err).Msg("backfill: failed flagging misconfigured integration")

			continue
		}

		flagged++
	}

	logx.FromContext(ctx).Info().Int("flagged_misconfigured", flagged).Int("reviewed", len(installations)).Msg("backfill: connected integrations reviewed")
}

// backfillReconcileLoops collapses each connected installation's recurring loops to exactly one
// per operation: every active reconcile job is cancelled and a single fresh loop is emitted with
// insert-time uniqueness, removing duplicate loops left by historical seeding races. Emitted
// loops are unique-keyed, so re-running the backfill against a healthy state is a reset, not a
// duplication
func backfillReconcileLoops(ctx context.Context, dbClient *ent.Client, rt *runtime.Runtime) {
	installations, err := dbClient.Integration.Query().
		Where(integration.StatusEQ(enums.IntegrationStatusConnected)).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to query connected integrations for loop reset")

		return
	}

	var reset int

	for _, installation := range installations {
		installCtx := intobvs.WithInstallation(ctx, installation)

		if err := rt.ResetReconcileLoops(installCtx, installation); err != nil {
			logx.FromContext(installCtx).Error().Err(err).Msg("backfill: failed resetting reconcile loops")

			continue
		}

		reset++
	}

	logx.FromContext(ctx).Info().Int("reset", reset).Int("reviewed", len(installations)).Msg("backfill: reconcile loop reset completed")
}

// directoryDupKey is one (integration_id, external_id) identity with its row count, scanned
// from the duplicate-detection group-by
type directoryDupKey struct {
	IntegrationID string `json:"integration_id"`
	ExternalID    string `json:"external_id"`
	Count         int    `json:"count"`
}

// minForkedRowCount is the row count at which an identity is considered forked
const minForkedRowCount = 2

// BackfillDirectoryDuplicates removes forked directory rows sharing an (integration_id,
// external_id) identity: concurrent syncs raced the same record into multiple rows, which the
// run-scoped unique index permitted, and the ingest lookup fails "not singular" on every
// subsequent sync until only one row remains. The earliest row survives; the extras and their
// memberships are deleted rather than repointed, because membership foreign keys are immutable
// and the next complete directory sync rebuilds active memberships against the survivor
func BackfillDirectoryDuplicates(ctx context.Context, dbClient *ent.Client) {
	var groupKeys []directoryDupKey
	if err := dbClient.DirectoryGroup.Query().
		GroupBy(directorygroup.FieldIntegrationID, directorygroup.FieldExternalID).
		Aggregate(ent.Count()).
		Scan(ctx, &groupKeys); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to scan directory groups for duplicate identities")

		return
	}

	groupsRemoved := 0

	for _, key := range groupKeys {
		if key.Count < minForkedRowCount {
			continue
		}

		rows, err := dbClient.DirectoryGroup.Query().
			Where(directorygroup.IntegrationID(key.IntegrationID), directorygroup.ExternalID(key.ExternalID)).
			Order(ent.Asc(directorygroup.FieldID)).
			All(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("integration_id", key.IntegrationID).Str("external_id", key.ExternalID).Msg("backfill: failed to load duplicate directory groups")

			continue
		}

		for _, loser := range rows[1:] {
			if _, err := dbClient.DirectoryMembership.Delete().
				Where(directorymembership.DirectoryGroupID(loser.ID)).
				Exec(ctx); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("directory_group_id", loser.ID).Msg("backfill: failed to delete memberships of duplicate directory group")

				continue
			}

			if err := dbClient.DirectoryGroup.DeleteOneID(loser.ID).Exec(ctx); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("directory_group_id", loser.ID).Msg("backfill: failed to delete duplicate directory group")

				continue
			}

			groupsRemoved++
		}
	}

	var accountKeys []directoryDupKey
	if err := dbClient.DirectoryAccount.Query().
		GroupBy(directoryaccount.FieldIntegrationID, directoryaccount.FieldExternalID).
		Aggregate(ent.Count()).
		Scan(ctx, &accountKeys); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to scan directory accounts for duplicate identities")

		return
	}

	accountsRemoved := 0

	for _, key := range accountKeys {
		if key.Count < minForkedRowCount {
			continue
		}

		rows, err := dbClient.DirectoryAccount.Query().
			Where(directoryaccount.IntegrationID(key.IntegrationID), directoryaccount.ExternalID(key.ExternalID)).
			Order(ent.Asc(directoryaccount.FieldID)).
			All(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("integration_id", key.IntegrationID).Str("external_id", key.ExternalID).Msg("backfill: failed to load duplicate directory accounts")

			continue
		}

		for _, loser := range rows[1:] {
			if _, err := dbClient.DirectoryMembership.Delete().
				Where(directorymembership.DirectoryAccountID(loser.ID)).
				Exec(ctx); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("directory_account_id", loser.ID).Msg("backfill: failed to delete memberships of duplicate directory account")

				continue
			}

			if err := dbClient.DirectoryAccount.DeleteOneID(loser.ID).Exec(ctx); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("directory_account_id", loser.ID).Msg("backfill: failed to delete duplicate directory account")

				continue
			}

			accountsRemoved++
		}
	}

	logx.FromContext(ctx).Info().Int("groups_removed", groupsRemoved).Int("accounts_removed", accountsRemoved).Msg("backfill: forked directory records removed")
}
