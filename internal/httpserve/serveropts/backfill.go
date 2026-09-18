package serveropts

import (
	"context"
	"errors"

	"github.com/samber/do/v2"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
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
	// leaving for example structure
	//
	//	{
	//		Name:    "ingest-batching",
	//		Version: "v2",
	//		Enabled: true,
	//		Run: func(ctx context.Context, deps backfillDeps) error {
	//			seedIntegrationLoops(ctx, deps.Runtime)
	//
	//			return nil
	//		},
	//	},
}

// HasRoutines checks if there are any backfilling operation to run
func HasRoutines() bool {
	return len(backfillRoutines) > 0
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
