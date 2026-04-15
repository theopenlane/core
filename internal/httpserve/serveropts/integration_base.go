package serveropts

import (
	"github.com/rs/zerolog/log"

	ent "github.com/theopenlane/core/internal/ent/generated"
	runtime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/internal/workflows/engine"
)

// WithIntegrationsRuntime builds the integration runtime from server settings and wires it
// into the handler. When a workflow engine is present it also injects integration dependencies.
// Initialization is skipped if the database client or Gala runtime is nil.
func WithIntegrationsRuntime(dbClient *ent.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		s.Config.Handler.IntegrationsConfig = s.Config.Settings.Integrations

		if dbClient == nil {
			return
		}

		galaInstance := s.Config.Handler.Gala
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
			DB:            dbClient,
			Gala:          galaInstance,
			Keystore:      credStore,
			RedisClient:   s.Config.Handler.RedisClient,
			CatalogConfig: s.Config.Settings.Integrations,
		})
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize integration runtime")
		}

		s.Config.Handler.IntegrationsRuntime = rt

		// set the runtime on the ent client so hooks/mutations can access it
		dbClient.IntegrationsRuntime = rt

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

