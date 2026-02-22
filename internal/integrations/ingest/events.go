package ingest

import (
	"context"
	"strings"

	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/gala"
)

const (
	// TopicIntegrationIngestRequested is emitted when webhook payloads should be ingested.
	TopicIntegrationIngestRequested = "integration.ingest.requested"
)

// RequestedPayload captures webhook alert envelopes for ingestion.
type RequestedPayload struct {
	// IntegrationID identifies the integration that owns the payload.
	IntegrationID string `json:"integration_id"`
	// Schema identifies the ingest mapping schema (vulnerability, asset, etc).
	Schema string `json:"schema"`
	// Envelopes holds provider alert payloads for ingestion.
	Envelopes []types.AlertEnvelope `json:"envelopes"`
}

// IntegrationIngestRequestedTopic is emitted when webhook payloads should be ingested.
var IntegrationIngestRequestedTopic = gala.Topic[RequestedPayload]{Name: gala.TopicName(TopicIntegrationIngestRequested)}

// RegisterIngestListeners registers ingest listeners on the supplied Gala registry.
func RegisterIngestListeners(registry *gala.Registry, db *ent.Client) ([]gala.ListenerID, error) {
	if registry == nil {
		return nil, ErrIngestEmitterRequired
	}
	if db == nil {
		return nil, ErrDBClientRequired
	}

	return gala.RegisterListeners(registry,
		gala.Definition[RequestedPayload]{
			Topic: IntegrationIngestRequestedTopic,
			Name:  "integration.ingest.requested",
			Handle: func(ctx gala.HandlerContext, payload RequestedPayload) error {
				return handleIngestRequested(ctx.Context, db, payload)
			},
		},
	)
}

func handleIngestRequested(ctx context.Context, db *ent.Client, payload RequestedPayload) error {
	if len(payload.Envelopes) == 0 {
		return nil
	}

	integrationID := strings.TrimSpace(payload.IntegrationID)
	if integrationID == "" {
		return ErrIngestIntegrationRequired
	}

	schema := strings.TrimSpace(payload.Schema)
	if schema == "" {
		return ErrIngestSchemaRequired
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	integrationRecord, err := db.Integration.Query().
		Where(integration.IDEQ(integrationID)).
		Only(allowCtx)
	if err != nil {
		return err
	}

	provider := types.ProviderTypeFromString(integrationRecord.Kind)
	if provider == types.ProviderUnknown {
		return ErrIngestProviderUnknown
	}

	switch normalizeMappingKey(schema) {
	case normalizeMappingKey(mappingSchemaVulnerability):
		if !SupportsVulnerabilityIngest(provider, integrationRecord.Config) {
			return ErrMappingNotFound
		}

		operationName := types.OperationVulnerabilitiesCollect
		operationConfig, err := operations.ResolveOperationConfig(&integrationRecord.Config, string(operationName), nil)
		if err != nil {
			return err
		}

		_, err = VulnerabilityAlerts(allowCtx, VulnerabilityIngestRequest{
			OrgID:             integrationRecord.OwnerID,
			IntegrationID:     integrationRecord.ID,
			Provider:          provider,
			Operation:         operationName,
			IntegrationConfig: integrationRecord.Config,
			ProviderState:     integrationRecord.ProviderState,
			OperationConfig:   operationConfig,
			Envelopes:         payload.Envelopes,
			DB:                db,
		})
		return err
	default:
		return ErrIngestSchemaUnsupported
	}
}
