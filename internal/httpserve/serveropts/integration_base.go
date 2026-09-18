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
	"github.com/theopenlane/core/v2/pkg/version"
)

// integrationSeedTopic is the gala topic the integration loop seed is submitted on
var integrationSeedTopic = gala.NamespacedTopic[integrationSeedRequest](gala.SystemVersioned, "startup.integrations.seed")

// integrationSeedRequest is the seed payload
type integrationSeedRequest struct{}

// WithIntegrationsRuntime builds the integration runtime and wires it into the handler
func WithIntegrationsRuntime(ctx context.Context, dbClient *ent.Client, galaInstance *gala.Gala) ServerOption {
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

		if err := galaInstance.Attach(gala.WithValue(rt)); err != nil {
			log.Panic().Err(err).Msg("failed to attach integration runtime to gala injector")
		}

		if _, err := gala.Register(galaInstance, gala.Definition[integrationSeedRequest]{
			Topic: integrationSeedTopic,
			Handle: func(handlerCtx gala.HandlerContext, _ integrationSeedRequest) error {
				seedIntegrationLoops(handlerCtx.Context, rt)

				return nil
			},
		}); err != nil {
			log.Panic().Err(err).Msg("failed to register integration loop seeding listener")
		}

		if _, err := galaInstance.EmitWithHeaders(ctx, integrationSeedTopic.Name, integrationSeedRequest{}, gala.Headers{
			UniqueKey:  gala.SystemVersioned.Key("startup.integrations.seed", version.Version),
			UniqueOnce: true,
		}); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to submit integration loop seed")
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

// seedIntegrationLoops ensures every reconcilable and scheduled operation has a live loop
func seedIntegrationLoops(ctx context.Context, rt *runtime.Runtime) {
	if err := rt.SeedReconcileJobs(ctx); err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("failed to seed one or more missing reconcile jobs")
	}

	if err := rt.SeedScheduledOperations(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to seed one or more scheduled operation listeners")
	}
}
