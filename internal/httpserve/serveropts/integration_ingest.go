package serveropts

import (
	"github.com/rs/zerolog/log"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/ingest"
	"github.com/theopenlane/core/pkg/events/soiree"
)

const defaultIntegrationIngestWorkers = 100

// WithIntegrationIngestEvents wires a dedicated ingest event bus for webhook payloads.
func WithIntegrationIngestEvents(dbClient *ent.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if dbClient == nil {
			return
		}

		bus := soiree.New(
			soiree.Workers(defaultIntegrationIngestWorkers),
			soiree.Client(dbClient),
		)

		if err := ingest.RegisterIngestListeners(bus, dbClient); err != nil {
			log.Panic().Err(err).Msg("failed to register integration ingest listeners")
		}

		s.Config.Handler.IntegrationIngestEmitter = bus
	})
}
