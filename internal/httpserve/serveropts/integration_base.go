package serveropts

import (
	"github.com/rs/zerolog/log"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/internal/keystore"
)

// WithIntegrationStore wires the integration persistence layer into the handlers
func WithIntegrationStore(dbClient *ent.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if dbClient == nil {
			return
		}

		s.Config.Handler.IntegrationStore = keystore.NewStore(dbClient)
	})
}

// WithIntegrationBroker attaches a keystore broker that exchanges credentials for access tokens
func WithIntegrationBroker() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		store := s.Config.Handler.IntegrationStore
		if store == nil {
			return
		}

		reg, ok := s.Config.Handler.IntegrationRegistry.(*registry.Registry)
		if !ok || reg == nil {
			return
		}

		s.Config.Handler.IntegrationBroker = keystore.NewBroker(store, reg)
	})
}

// WithIntegrationClients wires the client pool manager used by operations and handlers
func WithIntegrationClients() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		broker := s.Config.Handler.IntegrationBroker
		if broker == nil {
			return
		}

		reg, ok := s.Config.Handler.IntegrationRegistry.(*registry.Registry)
		if !ok || reg == nil {
			return
		}

		manager, err := keystore.NewClientPoolManager(broker, keystore.FlattenDescriptors(reg.ClientDescriptorCatalog()))
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize integration client pools")
		}

		s.Config.Handler.IntegrationClients = manager
	})
}

// WithIntegrationOperations wires the operation manager that executes provider logic
func WithIntegrationOperations() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		broker := s.Config.Handler.IntegrationBroker
		if broker == nil {
			return
		}

		reg, ok := s.Config.Handler.IntegrationRegistry.(*registry.Registry)
		if !ok || reg == nil {
			return
		}

		opts := []keystore.OperationManagerOption{}
		if s.Config.Handler.IntegrationClients != nil {
			opts = append(opts, keystore.WithOperationClients(s.Config.Handler.IntegrationClients))
		}

		manager, err := keystore.NewOperationManager(broker, keystore.FlattenOperationDescriptors(reg.OperationDescriptorCatalog()), opts...)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize integration operations manager")
		}

		s.Config.Handler.IntegrationOperations = manager
	})
}

// WithKeymaker wires the keymaker service used by OAuth flows
func WithKeymaker() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		store := s.Config.Handler.IntegrationStore
		reg, ok := s.Config.Handler.IntegrationRegistry.(*registry.Registry)
		if store == nil || !ok || reg == nil {
			return
		}

		sessions := keymaker.NewMemorySessionStore()
		svc, err := keymaker.NewService(reg, store, sessions, keymaker.ServiceOptions{})
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize keymaker service")
		}

		s.Config.Handler.KeymakerService = svc
	})
}
