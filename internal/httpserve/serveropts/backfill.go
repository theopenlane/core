package serveropts

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	intobvs "github.com/theopenlane/core/internal/integrations/observability"
	"github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// backfillBypassCaps lets the backfill write organizations and memberships without a request caller while
// skipping the org-filter, FGA, and managed-group guards the membership hooks would otherwise apply
const backfillBypassCaps = auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation | auth.CapBypassManagedGroup

// maxExactExternalID is the largest float64 that can still hold every integer exactly (2^53)
const maxExactExternalID = float64(1 << 53)

// integrationReconfigureObjectType is the notification object type for a misconfigured installation
const integrationReconfigureObjectType = "integration.reconfiguration.required"

// WithBackfill runs a one-time, non-blocking, config-gated, idempotent startup backfills
// use-cases for this are things a db migration can't easily handle, computed data or fields, or repairs
func WithBackfill(ctx context.Context, dbClient *ent.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if dbClient == nil || !s.Config.Settings.Backfill.Enabled {
			return
		}

		rt := s.Config.Handler.IntegrationsRuntime
		settings := s.Config.Settings.Backfill

		go func() {
			backfillCtx := privacy.DecisionContext(ctx, privacy.Allow)
			backfillCtx = auth.WithCaller(backfillCtx, &auth.Caller{Capabilities: backfillBypassCaps})

			if settings.DirectorySyncBackfill {
				backfillDirectoryExternalIDs(backfillCtx, dbClient)
			}

			if rt != nil {
				backfillIntegrationConfiguration(backfillCtx, dbClient, rt)

				if settings.FileBackups {
					backfillFileBackups(backfillCtx, dbClient, rt)
				}
			}

			if settings.FileRestores {
				backfillFileRestores(backfillCtx, dbClient)
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

		if err := markIntegrationErrored(installCtx, dbClient, installation, def.DisplayName); err != nil {
			logx.FromContext(installCtx).Error().Err(err).Msg("backfill: failed flagging misconfigured integration")

			continue
		}

		flagged++
	}

	logx.FromContext(ctx).Info().Int("flagged_misconfigured", flagged).Int("reviewed", len(installations)).Msg("backfill: connected integrations reviewed")
}

// markIntegrationErrored flags one installation as misconfigured and notifies the owning organization
func markIntegrationErrored(ctx context.Context, dbClient *ent.Client, installation *ent.Integration, displayName string) error {
	if err := dbClient.Integration.UpdateOneID(installation.ID).
		SetStatus(enums.IntegrationStatusErrored).
		Exec(ctx); err != nil {
		return err
	}

	logx.FromContext(ctx).Warn().Msg("backfill: integration is missing required configuration, marked errored")

	_, err := dbClient.Notification.Create().
		SetOwnerID(installation.OwnerID).
		SetNotificationType(enums.NotificationTypeOrganization).
		SetObjectType(integrationReconfigureObjectType).
		SetTitle(fmt.Sprintf("%s needs to be reconnected", displayName)).
		SetBody(fmt.Sprintf("The %s integration is missing required configuration and has stopped syncing. Reconnect it and supply the required settings to resume.", displayName)).
		SetData(map[string]any{
			"integration_id": installation.ID,
			"definition_id":  installation.DefinitionID,
		}).
		SetTopic(enums.NotificationTopicIntegration).
		Save(ctx)

	return err
}

// backfillDirectoryExternalIDs rewrites directory account and group external ids that the CEL double
// conversion stored in scientific notation (e.g. "1.47884153e+08" back to "147884153"); the "e+"
// contains query is just a cheap prefilter, the strict parse in decimalExternalID decides what
// actually gets touched, so values like emails that happen to contain "e+" are never rewritten
func backfillDirectoryExternalIDs(ctx context.Context, dbClient *ent.Client) {
	accounts, err := dbClient.DirectoryAccount.Query().
		Where(directoryaccount.ExternalIDContains("e+")).
		All(ctx)
	if err != nil {
		log.Error().Err(err).Msg("backfill: failed to query directory accounts with scientific notation external ids")
		return
	}

	accountsCorrected := 0

	for _, account := range accounts {
		corrected, ok := decimalExternalID(account.ExternalID)
		if !ok {
			continue
		}

		conflict, err := dbClient.DirectoryAccount.Query().
			Where(
				directoryaccount.OwnerID(account.OwnerID),
				directoryaccount.ExternalID(corrected),
				directoryaccount.IDNEQ(account.ID),
			).
			Exist(ctx)
		if err != nil {
			log.Error().Err(err).Str("directory_account_id", account.ID).Msg("backfill: failed to check directory account external id conflict")

			continue
		}

		if conflict {
			log.Warn().Str("directory_account_id", account.ID).Str("external_id", corrected).Msg("backfill: corrected external id already held by another directory account, skipping")

			continue
		}

		// external_id is immutable, so the fix has to go through the sql modifier
		if err := dbClient.DirectoryAccount.UpdateOneID(account.ID).
			Modify(func(u *sql.UpdateBuilder) {
				u.Set(directoryaccount.FieldExternalID, corrected)
			}).
			Exec(ctx); err != nil {
			log.Error().Err(err).Str("directory_account_id", account.ID).Msg("backfill: failed to correct directory account external id")

			continue
		}

		accountsCorrected++
	}

	groups, err := dbClient.DirectoryGroup.Query().
		Where(directorygroup.ExternalIDContains("e+")).
		All(ctx)
	if err != nil {
		log.Error().Err(err).Msg("backfill: failed to query directory groups with scientific notation external ids")
		return
	}

	groupsCorrected := 0

	for _, group := range groups {
		corrected, ok := decimalExternalID(group.ExternalID)
		if !ok {
			continue
		}

		conflict, err := dbClient.DirectoryGroup.Query().
			Where(
				directorygroup.OwnerID(group.OwnerID),
				directorygroup.ExternalID(corrected),
				directorygroup.IDNEQ(group.ID),
			).
			Exist(ctx)
		if err != nil {
			log.Error().Err(err).Str("directory_group_id", group.ID).Msg("backfill: failed to check directory group external id conflict")

			continue
		}

		if conflict {
			log.Warn().Str("directory_group_id", group.ID).Str("external_id", corrected).Msg("backfill: corrected external id already held by another directory group, skipping")

			continue
		}

		// external_id is immutable, so the fix has to go through the sql modifier
		if err := dbClient.DirectoryGroup.UpdateOneID(group.ID).
			Modify(func(u *sql.UpdateBuilder) {
				u.Set(directorygroup.FieldExternalID, corrected)
			}).
			Exec(ctx); err != nil {
			log.Error().Err(err).Str("directory_group_id", group.ID).Msg("backfill: failed to correct directory group external id")

			continue
		}

		groupsCorrected++
	}

	log.Info().Int("accounts_corrected", accountsCorrected).Int("groups_corrected", groupsCorrected).Msg("backfill: directory external id notation corrected")
}

// decimalExternalID converts a scientific notation external id back to plain digits, refusing
// anything that isn't a whole number float64 can represent exactly
func decimalExternalID(externalID string) (string, bool) {
	value, err := strconv.ParseFloat(externalID, 64)
	if err != nil || value != math.Trunc(value) || math.Abs(value) >= maxExactExternalID {
		return "", false
	}

	return strconv.FormatFloat(value, 'f', -1, 64), true
}

// backfillFileBackups enqueues a backup for existing files whose storage provider has a backup
// configured and whose backup is not already completed
func backfillFileBackups(ctx context.Context, dbClient *ent.Client, rt *runtime.Runtime) {
	if dbClient.ObjectManager == nil {
		log.Warn().Msg("backfill: object manager is nil, skipping backfill for file backups")
		return
	}

	if rt.Gala() == nil {
		log.Warn().Msg("backfill: gala runtime is unavailable, skipping backfill for file backups")
		return
	}

	sources := dbClient.ObjectManager.BackupSources()
	if len(sources) == 0 {
		return
	}

	sourceValues := lo.Map(sources, func(s storage.ProviderType, _ int) string {
		return string(s)
	})

	const batchSize = 10

	totalFiles := 0
	enqueuedCounter := 0
	failedCounter := 0
	lastKnownID := ""

	for {
		query := dbClient.File.Query().
			Where(
				file.StorageProviderIn(sourceValues...),
				// a file still needs a backup when it has never been attempted (backup_state is null) or it
				// failed and has not yet exhausted its retries; completed and exhausted files are skipped
				file.Or(
					file.BackupStateIsNil(),
					func(s *sql.Selector) {
						s.Where(sql.And(
							sql.Not(sqljson.ValueEQ(file.FieldBackupState, string(enums.FileBackupStatusCompleted), sqljson.Path("status"))),
							sql.Not(sqljson.ValueEQ(file.FieldBackupState, string(enums.FileBackupStatusExhausted), sqljson.Path("status"))),
						))
					},
				),
			).
			Order(file.ByID()).
			Limit(batchSize)

		if lastKnownID != "" {
			query = query.Where(file.IDGT(lastKnownID))
		}

		files, err := query.All(ctx)
		if err != nil {
			log.Error().Err(err).Msg("backfill: failed to query files missing a backup")
			return
		}

		if len(files) == 0 {
			break
		}

		totalFiles += len(files)

		for _, f := range files {
			lastKnownID = f.ID

			if err := hooks.EnqueueFileBackup(ctx, rt.Gala(), f.ID); err != nil {
				failedCounter++
				log.Error().Err(err).Str("file_id", f.ID).Msg("backfill: failed to enqueue file backup")

				continue
			}

			enqueuedCounter++
		}
	}

	log.Info().Int("enqueued_files", enqueuedCounter).
		Int("failed_files", failedCounter).
		Int("total_candidate_files", totalFiles).
		Msg("backfill: file backups enqueued")
}

// backfillFileRestores copies files back from their backup provider to their source provider, which
// is how a source recovers after its storage is lost and replaced. Files without a usable replica
// are skipped and reported, since there is nothing to restore them from
func backfillFileRestores(ctx context.Context, dbClient *ent.Client) {
	if dbClient.ObjectManager == nil {
		log.Warn().Msg("backfill: object manager is nil, skipping restore of file backups")
		return
	}

	sources := dbClient.ObjectManager.BackupSources()
	if len(sources) == 0 {
		return
	}

	sourceValues := lo.Map(sources, func(s storage.ProviderType, _ int) string {
		return string(s)
	})

	const batchSize = 10

	restoredCounter := 0
	failedCounter := 0
	skippedCounter := 0
	lastKnownID := ""

	for {
		query := dbClient.File.Query().
			Where(
				file.StorageProviderIn(sourceValues...),
				// only a completed replication has an object at the backup provider to restore from
				func(s *sql.Selector) {
					s.Where(sqljson.ValueEQ(file.FieldBackupState, string(enums.FileBackupStatusCompleted), sqljson.Path("status")))
				},
			).
			Order(file.ByID()).
			Limit(batchSize)

		if lastKnownID != "" {
			query = query.Where(file.IDGT(lastKnownID))
		}

		files, err := query.All(ctx)
		if err != nil {
			log.Error().Err(err).Msg("backfill: failed to query files with a backup replica")
			return
		}

		if len(files) == 0 {
			break
		}

		for _, f := range files {
			lastKnownID = f.ID

			storageFile := interceptors.StorageFileFromEnt(f)

			// a replication recorded without a key predates the key being persisted and cannot be
			// located at the destination, so it has to be replicated again rather than restored
			if storageFile == nil || storageFile.BackupLocation == nil {
				skippedCounter++

				log.Warn().Str("file_id", f.ID).Msg("backfill: file has no usable backup replica, skipping restore")

				continue
			}

			result, err := dbClient.ObjectManager.Restore(ctx, storageFile)
			if err != nil {
				failedCounter++

				log.Error().Err(err).Str("file_id", f.ID).Msg("backfill: failed to restore file from backup")

				continue
			}

			if result == nil {
				skippedCounter++

				continue
			}

			if err := persistRestoredLocation(ctx, dbClient, f, result); err != nil {
				failedCounter++

				log.Error().Err(err).Str("file_id", f.ID).Msg("backfill: failed to persist restored file location")

				continue
			}

			restoredCounter++
		}
	}

	log.Info().Int("restored_files", restoredCounter).
		Int("failed_files", failedCounter).
		Int("skipped_files", skippedCounter).
		Msg("backfill: file restores completed")
}

// persistRestoredLocation updates a file record when the replacement storage wrote the restored
// object to a different location than the file previously had
func persistRestoredLocation(ctx context.Context, dbClient *ent.Client, f *ent.File, result *objects.RestoreResult) error {
	if result.Key == f.StoragePath && result.Bucket == f.StorageVolume && result.Region == f.StorageRegion && result.URI == f.URI {
		return nil
	}

	update := dbClient.File.UpdateOneID(f.ID).
		SetStoragePath(result.Key).
		SetStorageVolume(result.Bucket).
		SetStorageRegion(result.Region)

	if result.URI != "" {
		update = update.SetURI(result.URI)
	}

	return update.Exec(ctx)
}
