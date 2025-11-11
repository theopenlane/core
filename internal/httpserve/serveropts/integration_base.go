package serveropts

import (
	"github.com/rs/zerolog/log"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/internal/keystore"
)

// WithIntegrationStore wires the integration persistence layer into the handlers.
func WithIntegrationStore(dbClient *ent.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if dbClient == nil {
			return
		}

		s.Config.Handler.IntegrationStore = keystore.NewStore(dbClient)
	})
}

// WithIntegrationBroker attaches a keystore broker that exchanges credentials for access tokens.
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

// WithKeymaker wires the keymaker service used by OAuth flows.
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
