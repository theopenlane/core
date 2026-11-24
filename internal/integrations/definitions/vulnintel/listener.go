package vulnintel

import (
	"slices"
	"time"

	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// enrichmentListener registers the listener enriching vulnerabilities after any provider's run persists them
func enrichmentListener(cfg Config) types.GalaListenerRegistration {
	return types.GalaListenerRegistration{
		Name: "vulnintel.enrich",
		Register: func(g *gala.Gala, services types.RuntimeServices) ([]gala.ListenerID, error) {
			e := &enricher{services: services, osv: newOSVClient(), maxPerRun: cfg.maxPerRun(), now: time.Now}

			return gala.Register(g, gala.Definition[operations.IngestCompletedEnvelope]{
				Topic:  operations.IngestCompletedTopic,
				Handle: e.handle,
			})
		},
	}
}

// handle enriches the installation's vulnerabilities when the completed run ingests them
func (e *enricher) handle(hc gala.HandlerContext, envelope operations.IngestCompletedEnvelope) error {
	if !slices.Contains(envelope.Schemas, entityops.SchemaVulnerability.Name) {
		return nil
	}

	result, err := e.enrich(hc.Context, envelope)
	if err != nil {
		logx.FromContext(hc.Context).Error().Err(err).Msg("vulnintel: vulnerability enrichment failed")

		return err
	}

	logx.FromContext(hc.Context).Info().Int("considered", result.Considered).Int("updated", result.Updated).Int("lookups", result.Lookups).Int("deferred", result.Deferred).Msg("vulnintel: vulnerability enrichment completed")

	return nil
}
