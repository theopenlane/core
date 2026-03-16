package serveropts

import (
	"reflect"

	"github.com/rs/zerolog/log"

	ent "github.com/theopenlane/core/internal/ent/generated"
	integrationregistry "github.com/theopenlane/core/internal/integrations/registry"
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
			DB:                    dbClient,
			Gala:                  galaInstance,
			Keystore:              credStore,
			RedisClient:           s.Config.Handler.RedisClient,
			CatalogConfig:         s.Config.Settings.Integrations,
			SkipExecutorListeners: wf != nil,
		})
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize integration runtime")
		}

		s.Config.Handler.IntegrationsRuntime = rt
		setIntegrationRegistry(dbClient, rt.Registry())

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

func setIntegrationRegistry(dbClient *ent.Client, reg *integrationregistry.Registry) {
	if dbClient == nil || reg == nil {
		return
	}

	field := reflect.ValueOf(dbClient).Elem().FieldByName("IntegrationRegistry")
	if !field.IsValid() || !field.CanSet() {
		return
	}

	value := reflect.ValueOf(reg)
	if value.Type().AssignableTo(field.Type()) {
		field.Set(value)
	}
}
