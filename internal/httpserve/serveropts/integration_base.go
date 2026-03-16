package serveropts

import (
	"github.com/rs/zerolog/log"

	ent "github.com/theopenlane/core/internal/ent/generated"
	IntegrationsRuntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/internal/workflows/engine"
)

// WithIntegrationsRuntime builds the integrationsv2 runtime from server settings and wires it
// into the handler. When a workflow engine is present it also injects v2 integration dependencies.
// Initialization is skipped if the database client or Gala runtime is nil.
func WithIntegrationsRuntime(dbClient *ent.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if dbClient == nil {
			return
		}

		galaInstance := s.Config.Handler.Gala
		if galaInstance == nil {
			log.Warn().Msg("gala runtime not available; integrationsv2 runtime will not be initialized")
			return
		}

		credStore, err := keystore.NewStore(dbClient)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize keystore for integrationsv2")
		}

		wf := s.Config.Handler.WorkflowEngine
		rt, err := IntegrationsRuntime.New(IntegrationsRuntime.Config{
			DB:                    dbClient,
			Gala:                  galaInstance,
			Keystore:              credStore,
			AuthStateStore:        keymaker.NewInMemoryAuthStateStore(),
			CatalogConfig:         s.Config.Settings.Integrations,
			SuccessRedirectURL:    s.Config.Settings.IntegrationSuccessRedirectURL,
			SkipExecutorListeners: wf != nil,
		})
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize integrationsv2 runtime")
		}

		s.Config.Handler.IntegrationsRuntime = rt

		if wf == nil {
			return
		}

		if err := wf.SetIntegrationDeps(engine.IntegrationDeps{
			Runtime: rt,
		}); err != nil {
			log.Panic().Err(err).Msg("failed to wire integrationsv2 deps into workflow engine")
		}
	})
}
