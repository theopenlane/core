package serveropts

import (
	"context"
	"errors"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/samber/do/v2"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

// backfillBypassCaps are the capabilities backfill routines run with
const backfillBypassCaps = auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation | auth.CapBypassManagedGroup

// backfillRoutineTopicName is the topic suffix and key namespace for backfill routines
const backfillRoutineTopicName = "startup.backfill.routine"

// backfillRoutineTopic carries one scheduled routine, so every routine retries and fails independently of its siblings
var backfillRoutineTopic = gala.NamespacedTopic[backfillRoutineRequest](gala.SystemVersioned, backfillRoutineTopicName)

// routineUniqueKey is the routine's run-once key that will be unique unless a version of the routine is changed
func (r backfillRoutine) routineUniqueKey() string {
	return gala.SystemVersioned.Key(backfillRoutineTopicName, r.Name, r.Version)
}

// backfillRoutineRequest names the routine to run
type backfillRoutineRequest struct {
	Name string `json:"name"`
}

// backfillDeps are the dependencies handed to every routine
type backfillDeps struct {
	// Client is the ent client resolved from the gala injector
	Client *ent.Client
	// Runtime is the integrations runtime
	Runtime *runtime.Runtime
	// Gala is the runtime routines emit follow-up work on
	Gala *gala.Gala
}

// backfillRoutine is one registered backfill and its run semantics
type backfillRoutine struct {
	// Name identifies the routine
	Name string
	// Version is the version of the routine being run
	Version string
	// Enabled schedules the routine
	Enabled bool
	// Run executes the routine
	Run func(context.Context, backfillDeps) error
}

// backfillRoutines are the registered backfill routines
var backfillRoutines = []backfillRoutine{
	{
		Name:    "backfill-schema-responsibilities",
		Version: "v1",
		Enabled: true,
		Run: func(ctx context.Context, deps backfillDeps) error {
			return backfillSchemaResponsibilities(ctx, deps.Client)
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
			seedIntegrationLoops(ctx, deps.Runtime)

			return nil
		},
	},
}

// WithBackfill submits the config-gated backfill scheduling run
func WithBackfill(ctx context.Context, galaApp *gala.Gala) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if !s.Config.Settings.Backfill.Enabled {
			return
		}

		if err := StartBackfill(ctx, galaApp); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to start")
		}
	})
}

// StartBackfill registers the routine listener and submits every enabled routine
func StartBackfill(ctx context.Context, galaApp *gala.Gala) error {
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

	var errs []error

	for _, routine := range backfillRoutines {
		if !routine.Enabled {
			continue
		}

		if _, err := galaApp.EmitWithHeaders(ctx, backfillRoutineTopic.Name, backfillRoutineRequest{Name: routine.Name}, gala.Headers{
			UniqueKey:  routine.routineUniqueKey(),
			UniqueOnce: true,
		}); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// runBackfillRoutine executes the named routine with its dependencies resolved from the injector
func runBackfillRoutine(handlerCtx gala.HandlerContext, galaApp *gala.Gala, name string) error {
	routine, ok := lo.Find(backfillRoutines, func(r backfillRoutine) bool {
		return r.Name == name
	})
	if !ok {
		logx.FromContext(handlerCtx.Context).Warn().Str("routine", name).Msg("backfill: routine is no longer registered, skipping")

		return nil
	}

	return routine.Run(handlerCtx.Context, backfillDeps{
		Client:  do.MustInvoke[*ent.Client](handlerCtx.Injector),
		Runtime: do.MustInvoke[*runtime.Runtime](handlerCtx.Injector),
		Gala:    galaApp,
	})
}

// backfillIngestBatching cancels queued per-record ingest jobs and reconcile loops for every installation
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

// purgeInstallationReconcileLoops cancels every live reconcile loop on the installation
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

// backfillInstanceIDs resolves and persists the instance id for every operational installation
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

// BackfillFileBackups enqueues a backup for existing files that still need one
func BackfillFileBackups(ctx context.Context, dbClient *ent.Client, galaApp *gala.Gala) {
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
