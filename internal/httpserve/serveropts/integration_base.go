package serveropts

import (
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/keystore"
)

// IntegrationBrokerConfig configures the integration credential broker.
type IntegrationBrokerConfig struct {
	WorkloadIdentityIssuer keystore.WorkloadIdentityIssuer
}

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
func WithIntegrationBroker(cfg IntegrationBrokerConfig) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		store := s.Config.Handler.IntegrationStore
		if store == nil {
			return
		}

		registry := &s.Config.Handler.IntegrationRegistry

		broker := keystore.New(store, registry)
		issuer := cfg.WorkloadIdentityIssuer
		if issuer == nil {
			issuer = keystore.NewGCPWorkloadIdentityIssuer(store)
		}

		broker.WithWorkloadIdentityIssuer(issuer)

		s.Config.Handler.IntegrationBroker = broker
	})
}
