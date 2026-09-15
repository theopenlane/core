package serveropts

import (
	"context"

	"github.com/rs/zerolog/log"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	runtime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/internal/keystore"
	"github.com/theopenlane/core/v2/internal/workflows/engine"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// WithIntegrationsRuntime builds the integration runtime from server settings and wires it
// into the handler. When a workflow engine is present it also injects integration dependencies.
// Initialization is skipped if the database client or Gala runtime is nil.
func WithIntegrationsRuntime(dbClient *ent.Client, galaInstance *gala.Gala) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		s.Config.Handler.IntegrationsConfig = s.Config.Settings.Integrations

		if dbClient == nil {
			return
		}

		if galaInstance == nil {
			log.Warn().Msg("gala runtime not available; integration runtime will not be initialized")
			return
		}

		credStore, err := keystore.NewStore(dbClient)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize keystore for integrations")
		}

		wf := s.Config.Handler.WorkflowEngine
		rt, err := runtime.New(runtime.Config{
			DB:               dbClient,
			Gala:             galaInstance,
			Keystore:         credStore,
			RedisClient:      s.Config.Handler.RedisClient,
			CatalogConfig:    s.Config.Settings.Integrations,
			FederationIssuer: s.Config.Settings.Auth.Token.Issuer,
			DevMode:          s.Config.Settings.Server.Dev,
		})
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize integration runtime")
		}

		s.Config.Handler.IntegrationsRuntime = rt

		runtime.SetDefault(rt)

		// make the runtime resolvable from gala listener injectors
		if err := galaInstance.Attach(gala.WithValue(rt)); err != nil {
			log.Panic().Err(err).Msg("failed to attach integration runtime to gala injector")
		}

		if _, err := gala.Register(galaInstance, gala.Definition[backfillCompleted]{
			Topic: backfillCompletedTopic,
			Handle: func(handlerCtx gala.HandlerContext, _ backfillCompleted) error {
				seedIntegrationLoops(handlerCtx.Context, rt)

				return nil
			},
		}); err != nil {
			log.Panic().Err(err).Msg("failed to register integration loop seeding listener")
		}

		if wf == nil {
			return
		}

		if err := wf.SetIntegrationDeps(engine.IntegrationDeps{
			Runtime: rt,
		}); err != nil {
			log.Panic().Err(err).Msg("failed to wire integration deps into workflow engine")
		}
	})
}

// seedIntegrationLoops ensures every connected installation's reconcilable operations and every
// scheduled operation have a live loop, logging whatever could not be seeded
func seedIntegrationLoops(ctx context.Context, rt *runtime.Runtime) {
	if err := rt.SeedReconcileJobs(ctx); err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("failed to seed one or more missing reconcile jobs")
	}

	if err := rt.SeedScheduledOperations(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to seed one or more scheduled operation listeners")
	}
}
