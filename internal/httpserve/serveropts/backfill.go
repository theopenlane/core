package serveropts

import (
	"context"

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

// schedulerKeyPrefix is the base of the scheduling run's uniqueness key
const schedulerKeyPrefix = "startup-backfill-v2"

// routineKeyPrefix seeds each routine's run-once key
const routineKeyPrefix = "startup-backfill-routine"

// schedulerKey is the scheduling run's uniqueness key. It carries the application version the
// binary was built from, so every pod in a rollout shares one key and the next release schedules
// again. An unstamped build, e.g. local development, has no application version and falls back to
// the bare prefix
func schedulerKey() string {
	if version.Version == "" {
		return schedulerKeyPrefix
	}

	return schedulerKeyPrefix + "-" + version.Version
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

// backfillDeps are the dependencies handed to every routine
type backfillDeps struct {
	// Client is the ent client resolved from the gala injector
	Client *ent.Client
	// Runtime is the integrations runtime
	Runtime *runtime.Runtime
	// Gala is the runtime routines emit follow-up work on
	Gala *gala.Gala
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

// backfillRoutines are the registered backfill routines. Add a routine here to have the scheduling run pick it up
// Mark as enabled=false to keep it here but do not schedule it
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
		Enabled: true,
		Run: func(ctx context.Context, deps backfillDeps) error {
			backfillFileBackups(ctx, deps.Client, deps.Gala)

			return nil
		},
	},
}

// WithBackfill submits the config-gated backfill scheduling run as a gala job: every pod submits
// the same unique key, so exactly one process schedules, and the run then fans out one job per
// registered routine carrying that routine's own run-once semantics
func WithBackfill(ctx context.Context, galaApp *gala.Gala) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if !s.Config.Settings.Backfill.Enabled {
			return
		}

		if _, err := gala.Register(galaApp, gala.Definition[backfillRequest]{
			Topic: backfillTopic,
			Caller: func(*auth.Caller, backfillRequest) *auth.Caller {
				return &auth.Caller{Capabilities: backfillBypassCaps}
			},
			Handle: func(handlerCtx gala.HandlerContext, _ backfillRequest) error {
				return scheduleBackfillRoutines(handlerCtx.Context, galaApp)
			},
		}); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to register scheduling listener")

			return
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
			logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to register routine listener")

			return
		}

		if _, err := galaApp.EmitWithHeaders(ctx, backfillTopic.Name, backfillRequest{}, gala.Headers{
			UniqueKey:  schedulerKey(),
			UniqueOnce: true,
		}); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to submit scheduling run")
		}
	})
}

// scheduleBackfillRoutines emits one job per registered routine
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
