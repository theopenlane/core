package serveropts

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/samber/do/v2"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/jobspec"

	"github.com/theopenlane/core/v2/config"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/customdomain"
	"github.com/theopenlane/core/v2/internal/ent/generated/dnsverification"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/mappabledomain"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/core/v2/pkg/version"
)

// backfillBypassCaps lets the backfill write organizations and memberships without a request caller while
// skipping the org-filter, FGA, and managed-group guards the membership hooks would otherwise apply
const backfillBypassCaps = auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation | auth.CapBypassManagedGroup

// backfillTopic is the gala topic the backfill scheduling run is submitted on
var backfillTopic = gala.NamespacedTopic[backfillRequest](gala.System, "startup.backfill")

// backfillRoutineTopic carries one scheduled routine, so every routine retries and fails independently of its siblings
var backfillRoutineTopic = gala.NamespacedTopic[backfillRoutineRequest](gala.System, "startup.backfill.routine")

// backfillCompletedTopic announces that the backfill has finished for this release, so startup work
// that must observe the backfilled state (integration loop seeding) can begin
var backfillCompletedTopic = gala.NamespacedTopic[backfillCompleted](gala.System, "startup.backfill.completed")

// schedulerKeyPrefix is the base of the scheduling run's uniqueness key
const schedulerKeyPrefix = "startup-backfill-v5"

// routineKeyPrefix seeds each routine's run-once key
const routineKeyPrefix = "startup-backfill-routine"

// backfillCompletedKeyPrefix is the base of the completed event's uniqueness key
const backfillCompletedKeyPrefix = "startup-backfill-completed"

// versionedKey appends the application version the binary was built from to a key prefix, so
// every pod in a rollout shares one key and the next release keys again. An unstamped build,
// e.g. local development, has no application version and falls back to the bare prefix
func versionedKey(prefix string) string {
	if version.Version == "" {
		return prefix
	}

	return prefix + "-" + version.Version
}

// schedulerKey is the scheduling run's uniqueness key
func schedulerKey() string {
	return versionedKey(schedulerKeyPrefix)
}

// backfillCompletedKey is the completed event's uniqueness key, so a restart of the same release does
// not re-run the startup work downstream of the backfill
func backfillCompletedKey() string {
	return versionedKey(backfillCompletedKeyPrefix)
}

// routineUniqueKey is the routine's run-once key that will be unique unless a version of the routine is changed
func (r backfillRoutine) routineUniqueKey() string {
	return routineKeyPrefix + "-" + r.Name + "-" + r.Version
}

// backfillRequest is the payload for a backfill scheduling run submission
type backfillRequest struct{}

// backfillRoutineRequest names the routine to run
type backfillRoutineRequest struct {
	Name string `json:"name"`
}

// backfillCompleted is the payload of the completed event
type backfillCompleted struct{}

// backfillDeps are the dependencies handed to every routine
type backfillDeps struct {
	// Client is the ent client resolved from the gala injector
	Client *ent.Client
	// Runtime is the integrations runtime
	Runtime *runtime.Runtime
	// Gala is the runtime routines emit follow-up work on
	Gala         *gala.Gala
	ServerConfig config.Server
}

// backfillRoutine is one registered backfill and the run semantics it declares. Routines are
// offered to the queue whenever the scheduling run executes and the routine's own key is what decides the outcome from there
// By default that key holds through terminal states, so the routine runs a single time and later scheduling runs skip
// it as a duplicate. Bumping Version is what runs it again. A routine declaring Repeat has no such hold, so it runs on every scheduling run
type backfillRoutine struct {
	// Name identifies the routine and seeds its run-once uniqueness key
	Name string
	// Version is the version of the routine being run
	Version string
	// Enabled schedules the routine; a disabled routine stays registered but is never emitted
	Enabled bool
	// Repeat runs the routine on every scheduling run, so once per startup-backfill-v2 run
	Repeat bool
	// Run executes the routine
	Run func(context.Context, backfillDeps) error
}

// backfillRoutines are the registered backfill routines, each scheduled as its own job so they run
// independently and concurrently. Add a routine here to have the scheduling run pick it up; mark it
// enabled=false to keep it registered but unscheduled. The ingest-batching routine runs the ordered
// ingest migration (batching purge, then instance-id resolution, then provenance) as a single unit,
// since provenance must stamp the instance id that resolution wrote, and signals completion so
// integration loop seeding runs against the migrated state
var backfillRoutines = []backfillRoutine{
	{
		Name:    "reconcile-loops",
		Version: "v1",
		Enabled: false,
		Run: func(ctx context.Context, deps backfillDeps) error {
			backfillReconcileLoops(ctx, deps.Client, deps.Runtime)

			return nil
		},
	},
	{
		Name:    "file-backups",
		Version: "v1",
		Enabled: false,
		Run: func(ctx context.Context, deps backfillDeps) error {
			backfillFileBackups(ctx, deps.Client, deps.Gala)

			return nil
		},
	},
	{
		Name:    "integration-expiry",
		Version: "v1",
		Enabled: false,
		Run: func(ctx context.Context, deps backfillDeps) error {
			backfillIntegrationExpiry(ctx, deps.Client)

			return nil
		},
	},
	{
		Name:    "ingest-batching",
		Version: "v2",
		Enabled: true,
		Run: func(ctx context.Context, deps backfillDeps) error {
			backfillIngestBatching(ctx, deps.Client, deps.Runtime)
			backfillInstanceIDs(ctx, deps.Client, deps.Runtime)
			backfillProvenance(ctx, deps.Client, deps.Runtime)

			return emitBackfillCompleted(ctx, deps.Gala)
		},
	},
	{
		Name:    "recreate-preview-domains",
		Version: "v3",
		Enabled: true,
		Run: func(ctx context.Context, deps backfillDeps) error {
			return backfillPreviewDomains(ctx, deps.Client, deps.ServerConfig)
		},
	},
}

// WithBackfill submits the config-gated backfill scheduling run as a gala job: every pod submits
// the same unique key, so exactly one process schedules, and the run then fans out one job per
// registered routine carrying that routine's own run-once semantics. With backfills disabled the
// completed event is submitted directly, under the same per-release key, so the startup work
// downstream of the backfill still runs exactly once per release
func WithBackfill(ctx context.Context, galaApp *gala.Gala) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if !s.Config.Settings.Backfill.Enabled {
			if err := emitBackfillCompleted(ctx, galaApp); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to submit completed event with backfills disabled")
			}

			return
		}

		if err := StartBackfill(ctx, galaApp); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to start")
		}
	})
}

// StartBackfill registers the backfill listeners on galaApp and submits the scheduling run, which
// fans out one job per enabled routine, each carrying its own run-once semantics
func StartBackfill(ctx context.Context, galaApp *gala.Gala) error {
	if _, err := gala.Register(galaApp, gala.Definition[backfillRequest]{
		Topic: backfillTopic,
		Caller: func(*auth.Caller, backfillRequest) *auth.Caller {
			return &auth.Caller{Capabilities: backfillBypassCaps}
		},
		Handle: func(handlerCtx gala.HandlerContext, _ backfillRequest) error {
			return scheduleBackfillRoutines(handlerCtx.Context, galaApp)
		},
	}); err != nil {
		return err
	}

	if _, err := gala.Register(galaApp, gala.Definition[backfillRoutineRequest]{
		Topic: backfillRoutineTopic,
		Caller: func(*auth.Caller, backfillRoutineRequest) *auth.Caller {
			return &auth.Caller{Capabilities: backfillBypassCaps}
		},
		Handle: func(handlerCtx gala.HandlerContext, req backfillRoutineRequest) error {
			return runBackfillRoutine(handlerCtx, galaApp, req.Name)
		},
	}); err != nil {
		return err
	}

	_, err := galaApp.EmitWithHeaders(ctx, backfillTopic.Name, backfillRequest{}, gala.Headers{
		UniqueKey:  schedulerKey(),
		UniqueOnce: true,
	})

	return err
}

// scheduleBackfillRoutines emits one job per enabled routine
func scheduleBackfillRoutines(ctx context.Context, galaApp *gala.Gala) error {
	for _, routine := range backfillRoutines {
		if !routine.Enabled {
			continue
		}

		once := !routine.Repeat

		if _, err := galaApp.EmitWithHeaders(ctx, backfillRoutineTopic.Name, backfillRoutineRequest{Name: routine.Name}, gala.Headers{
			UniqueKey:  routine.routineUniqueKey(),
			UniqueOnce: once,
		}); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("routine", routine.Name).Msg("backfill: failed to schedule routine")
		}
	}

	return nil
}

// runBackfillRoutine executes the named routine with its dependencies resolved from the injector
func runBackfillRoutine(handlerCtx gala.HandlerContext, galaApp *gala.Gala, name string, serverConfig config.Server) error {
	routine, ok := lo.Find(backfillRoutines, func(r backfillRoutine) bool {
		return r.Name == name
	})
	if !ok {
		logx.FromContext(handlerCtx.Context).Warn().Str("routine", name).Msg("backfill: routine is no longer registered, skipping")

		return nil
	}

	return routine.Run(handlerCtx.Context, backfillDeps{
		Client:       do.MustInvoke[*ent.Client](handlerCtx.Injector),
		Runtime:      do.MustInvoke[*runtime.Runtime](handlerCtx.Injector),
		Gala:         galaApp,
		ServerConfig: serverConfig,
	})
}

// emitBackfillCompleted submits the completed event under its per-release run-once key
func emitBackfillCompleted(ctx context.Context, galaApp *gala.Gala) error {
	_, err := galaApp.EmitWithHeaders(ctx, backfillCompletedTopic.Name, backfillCompleted{}, gala.Headers{
		UniqueKey:  backfillCompletedKey(),
		UniqueOnce: true,
	})

	return err
}

// backfillReconcileLoops collapses each connected installation's recurring loops to exactly one
// per operation: every active reconcile job is cancelled and a single fresh loop is emitted with
// insert-time uniqueness, removing duplicate loops left by historical seeding races. Emitted
// loops are unique-keyed, so re-running the backfill against a healthy state is a reset, not a
// duplication
func backfillReconcileLoops(ctx context.Context, dbClient *ent.Client, rt *runtime.Runtime) {
	installations, err := dbClient.Integration.Query().
		Where(integration.StatusIn(enums.IntegrationOperationalStatuses...)).
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

// backfillIngestBatching transitions every installation from the fan-out ingest model to the
// batched model: queued per-record ingest jobs and every recurring reconcile loop are cancelled,
// since the next batched sync re-ingests their records with provenance and loop seeding runs
// fresh once the backfill completes. Loops are purged per operation rather than through the
// installation-wide purge, which would also cancel in-flight one-shot operation runs
func backfillIngestBatching(ctx context.Context, dbClient *ent.Client, rt *runtime.Runtime) {
	installations, err := dbClient.Integration.Query().All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to query integrations for ingest batching transition")

		return
	}

	var purgedIngest, purgedLoops int

	for _, installation := range installations {
		installCtx := intobvs.WithInstallation(ctx, installation)

		count, err := rt.PurgeInstallationIngestJobs(installCtx, installation.ID)
		if err != nil {
			logx.FromContext(installCtx).Error().Err(err).Msg("backfill: failed purging queued per-record ingest jobs")

			continue
		}

		purgedIngest += count

		count, err = purgeInstallationReconcileLoops(installCtx, rt, installation)
		if err != nil {
			logx.FromContext(installCtx).Error().Err(err).Msg("backfill: failed purging reconcile loops")

			continue
		}

		purgedLoops += count
	}

	logx.FromContext(ctx).Info().Int("purged_ingest", purgedIngest).Int("purged_loops", purgedLoops).Int("reviewed", len(installations)).Msg("backfill: queued per-record ingest jobs and reconcile loops purged")
}

// purgeInstallationReconcileLoops cancels every live reconcile loop for each reconcilable operation
// on the installation, matching the per-operation fragment the runtime's health handling purges with
func purgeInstallationReconcileLoops(ctx context.Context, rt *runtime.Runtime, installation *ent.Integration) (int, error) {
	def, ok := rt.Registry().Definition(installation.DefinitionID)
	if !ok {
		return 0, nil
	}

	var purged int

	for _, op := range def.Operations {
		if !op.Policy.Reconcile {
			continue
		}

		fragment, err := types.PropertiesFragment(map[string]string{
			"entityId":  installation.ID,
			"operation": op.Name,
			"runType":   enums.IntegrationRunTypeReconcile.String(),
		})
		if err != nil {
			return purged, err
		}

		count, err := rt.Gala().PurgeActiveJobsWithMetadata(intobvs.WithOperation(ctx, op.Name), fragment)
		if err != nil {
			return purged, err
		}

		purged += count
	}

	return purged, nil
}

// backfillInstanceIDs resolves and persists the current instance id for every operational installation
// before provenance stamping and batched ingest run, overwriting a stale id left by a definition whose
// instance-id derivation changed; never-connected installations resolve to nothing and are left for
// their next reconnect to populate. Errored installations are skipped here and self-heal on recovery
func backfillInstanceIDs(ctx context.Context, dbClient *ent.Client, rt *runtime.Runtime) {
	installations, err := dbClient.Integration.Query().
		Where(integration.StatusIn(enums.IntegrationOperationalStatuses...)).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to query integrations for instance id resolution")

		return
	}

	var resolved, unresolved int

	for _, installation := range installations {
		installCtx := intobvs.WithInstallation(ctx, installation)

		if err := rt.BackfillInstallationInstanceID(installCtx, installation); err != nil {
			logx.FromContext(installCtx).Error().Err(err).Str("integration_id", installation.ID).Msg("backfill: failed resolving installation instance id")

			unresolved++

			continue
		}

		if installation.InstallationMetadata.Display.ExternalID == "" {
			unresolved++

			continue
		}

		resolved++
	}

	logx.FromContext(ctx).Info().Int("resolved", resolved).Int("unresolved", unresolved).Int("reviewed", len(installations)).Msg("backfill: installation instance id resolution completed")
}

// backfillProvenance stamps every ingest schema's fill-only provenance columns for every installation's existing records
func backfillProvenance(ctx context.Context, dbClient *ent.Client, rt *runtime.Runtime) {
	installations, err := dbClient.Integration.Query().All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to query integrations for provenance stamping")

		return
	}

	var stamped int

	for _, installation := range installations {
		installCtx := intobvs.WithInstallation(ctx, installation)

		count, err := rt.BackfillInstallationProvenance(installCtx, installation)
		if err != nil {
			logx.FromContext(installCtx).Error().Err(err).Str("integration_id", installation.ID).Msg("backfill: failed stamping installation provenance")

			continue
		}

		stamped += count
	}

	logx.FromContext(ctx).Info().Int("stamped", stamped).Int("reviewed", len(installations)).Msg("backfill: installation provenance stamping completed")
}

// backfillIntegrationExpiry stamps expires_at on pending installations created before the
// column existed, deriving the expiry from the row's last update
func backfillIntegrationExpiry(ctx context.Context, dbClient *ent.Client) {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	rows, err := dbClient.Integration.Query().
		Where(integration.StatusEQ(enums.IntegrationStatusPending), integration.ExpiresAtIsNil()).
		All(allowCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to query pending integrations missing an expiry")

		return
	}

	var stamped int

	for _, row := range rows {
		if err := dbClient.Integration.UpdateOneID(row.ID).
			SetExpiresAt(row.UpdatedAt.Add(runtime.PendingInstallationTTL)).
			Exec(allowCtx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("integration_id", row.ID).Msg("backfill: failed stamping installation expiry")

			continue
		}

		stamped++
	}

	logx.FromContext(ctx).Info().Int("stamped", stamped).Int("reviewed", len(rows)).Msg("backfill: integration expiry stamping completed")
}

// backfillFileBackups enqueues a backup for existing files whose storage provider has a backup
// configured and whose backup is not already completed
func backfillFileBackups(ctx context.Context, dbClient *ent.Client, galaApp *gala.Gala) {
	if dbClient.ObjectManager == nil {
		logx.FromContext(ctx).Warn().Msg("backfill: object manager is nil, skipping backfill for file backups")
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
			logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to query files missing a backup")
			return
		}

		if len(files) == 0 {
			break
		}

		totalFiles += len(files)

		for _, f := range files {
			lastKnownID = f.ID

			if _, err := galaApp.EmitWithHeaders(ctx, hooks.FileBackupTopic.Name, hooks.FileBackupRequest{FileID: f.ID}, gala.Headers{}); err != nil {
				failedCounter++

				logx.FromContext(ctx).Error().Err(err).Str("file_id", f.ID).Msg("backfill: failed to enqueue file backup")

				continue
			}

			enqueuedCounter++
		}
	}

	logx.FromContext(ctx).Info().Int("enqueued_files", enqueuedCounter).Int("failed_files", failedCounter).Int("total_candidate_files", totalFiles).Msg("backfill: file backups enqueued")
}

func backfillPreviewDomains(ctx context.Context, client *ent.Client, cfg config.Server) error {
	if client.Job == nil {
		logx.FromContext(ctx).Warn().Msg("backfill: job client is nil, skipping preview domains backfill")

		return nil
	}

	// a preview domain is stale when cloudflare blocked its hostname or when it hangs off a mappable domain other than the configured target
	domains, err := client.CustomDomain.Query().
		Where(
			customdomain.DomainTypeEQ(enums.CustomDomainTypePreview),
			customdomain.TrustCenterIDNEQ(""), // must be linked to a trustcenter
			customdomain.Or(
				customdomain.HasDNSVerificationWith(
					dnsverification.DNSVerificationStatusEQ(enums.DNSVerificationStatusBlocked),
				),
				customdomain.HasMappableDomainWith(
					mappabledomain.NameNEQ(cfg.TrustCenterCnameTarget),
				),
			),
		).
		WithMappableDomain().
		// fine to do this as items <= 50
		All(ctx)
	if err != nil {
		return err
	}

	if len(domains) == 0 {
		return nil
	}

	var queuedCounter, deleteCounter int

	for _, domain := range domains {
		// make sure to clear out the preview domain. if not done, the resolver will always fail
		// when the job calls creation of a new trustcenter domain as we validate the presence of the domain
		// if a domain is set, you cannot create another one through the createTrustcenterDomain resolver
		cleared, err := client.TrustCenter.Update().
			Where(
				trustcenter.ID(domain.TrustCenterID),
				trustcenter.PreviewDomainID(domain.ID),
			).
			ClearPreviewDomainID().
			Save(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).
				Str("trust_center_id", domain.TrustCenterID).
				Msg("backfill: failed to clear blocked preview domain")

			return err
		}

		// only recreate when this row was still the trust center's preview domain, otherwise a newer one already replaced it
		if cleared > 0 {
			_, err = client.Job.Insert(ctx, jobspec.CreatePreviewDomainArgs{
				TrustCenterID:            domain.TrustCenterID,
				TrustCenterPreviewZoneID: cfg.TrustCenterPreviewZoneID,
				TrustCenterCnameTarget:   cfg.TrustCenterCnameTarget,
			}, nil)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).
					Str("custom_domain_id", domain.ID).
					Str("domain", domain.CnameRecord).
					Msg("backfill: failed to queue preview domain creation")

				return err
			}

			queuedCounter++
		}

		if domain.Edges.MappableDomain == nil {
			logx.FromContext(ctx).Warn().
				Str("custom_domain_id", domain.ID).
				Str("domain", domain.CnameRecord).
				Msg("backfill: stale preview domain has no mappable domain, skipping cleanup")

			continue
		}

		// the stale cname record was written into the zone of the mappable domain the row was linked to
		_, err = client.Job.Insert(ctx, jobspec.DeletePreviewDomainArgs{
			CustomDomainID:           domain.ID,
			TrustCenterPreviewZoneID: domain.Edges.MappableDomain.ZoneID,
		}, nil)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).
				Str("custom_domain_id", domain.ID).
				Str("domain", domain.CnameRecord).
				Msg("backfill: failed to queue stale preview domain deletion")

			return err
		}

		deleteCounter++
	}

	logx.FromContext(ctx).Info().Int("queued", queuedCounter).Int("deleting", deleteCounter).
		Msg("backfill: stale preview domains queued for recreation and cleanup")

	return nil
}
