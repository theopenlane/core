package serveropts

import (
	"github.com/rs/zerolog/log"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	integrationruntime "github.com/theopenlane/core/internal/integrations/runtime"
)

// WithIntegrationRuntime wires the full integration runtime dependency graph:
// Registry -> Store -> Broker -> Clients -> Operations -> Activation -> Ingest.
func WithIntegrationRuntime(dbClient *ent.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if dbClient == nil {
			return
		}

		reg, ok := s.Config.Handler.IntegrationRegistry.(*registry.Registry)
		if !ok || reg == nil {
			return
		}

		integrationDeps, err := integrationruntime.New(integrationruntime.Config{
			Registry:       reg,
			DB:             dbClient,
			WorkflowEngine: s.Config.Handler.WorkflowEngine,
		})
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize integration runtime")
		}

		s.Config.Handler.IntegrationStore = integrationDeps.Store
		s.Config.Handler.IntegrationBroker = integrationDeps.Broker
		s.Config.Handler.IntegrationClients = integrationDeps.Clients
		s.Config.Handler.IntegrationOperations = integrationDeps.Operations
		s.Config.Handler.IntegrationKeymaker = integrationDeps.Keymaker
		s.Config.Handler.IntegrationActivation = integrationDeps.Activation
		s.Config.Handler.IntegrationIngest = integrationDeps.Ingest
	})
}
