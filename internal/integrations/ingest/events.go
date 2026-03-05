package ingest

import (
	"cmp"
	"context"
	"slices"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// RequestedPayload captures webhook alert envelopes for ingestion.
type RequestedPayload struct {
	// IntegrationID identifies the integration that owns the payload.
	IntegrationID string `json:"integration_id"`
	// Schema identifies the ingest mapping schema (vulnerability, directory account, etc).
	Schema string `json:"schema"`
	// Operation optionally identifies the provider operation that produced the envelopes.
	Operation string `json:"operation,omitempty"`
	// Envelopes holds provider alert payloads for ingestion.
	Envelopes []types.AlertEnvelope `json:"envelopes"`
}

// requestedSchemaContract pairs a mapping schema with its Gala topic and listener name
type requestedSchemaContract struct {
	Schema   types.MappingSchema
	Topic    gala.TopicName
	Listener string
}

// requestedSchemaContracts maps schema names to their generated ingest topic contracts, built at init time
var requestedSchemaContracts = lo.Reduce(lo.Entries(integrationgenerated.IntegrationIngestSchemas), func(
	acc map[string]requestedSchemaContract,
	entry lo.Entry[string, integrationgenerated.IntegrationIngestSchema],
	_ int,
) map[string]requestedSchemaContract {
	schemaName := entry.Key
	if schemaName == "" {
		return acc
	}

	contract := entry.Value
	acc[schemaName] = requestedSchemaContract{
		Schema:   types.MappingSchema(entry.Key),
		Topic:    gala.TopicName(contract.Topic),
		Listener: contract.Listener,
	}

	return acc
}, map[string]requestedSchemaContract{})


// IngestRequestedTopicForSchema resolves the schema-scoped ingest topic
func IngestRequestedTopicForSchema(schema string) (gala.Topic[RequestedPayload], bool) {
	contract, ok := requestedSchemaContracts[schema]
	if !ok {
		return gala.Topic[RequestedPayload]{}, false
	}

	return gala.Topic[RequestedPayload]{Name: contract.Topic}, true
}

// RegisterIngestListeners registers ingest listeners on the supplied Gala registry
func RegisterIngestListeners(registry *gala.Registry, db *ent.Client, mappingIndex types.MappingIndex) ([]gala.ListenerID, error) {
	if registry == nil {
		return nil, ErrIngestEmitterRequired
	}
	if db == nil {
		return nil, ErrDBClientRequired
	}

	contracts := lo.Values(requestedSchemaContracts)
	slices.SortFunc(contracts, func(a, b requestedSchemaContract) int {
		return cmp.Compare(string(a.Topic), string(b.Topic))
	})

	definitions := lo.Map(contracts, func(contract requestedSchemaContract, _ int) gala.Definition[RequestedPayload] {
		return gala.Definition[RequestedPayload]{
			Topic: gala.Topic[RequestedPayload]{Name: contract.Topic},
			Name:  contract.Listener,
			Handle: func(ctx gala.HandlerContext, payload RequestedPayload) error {
				return handleIngestRequested(ctx.Context, db, payload, mappingIndex)
			},
		}
	})

	return gala.RegisterListeners(registry, definitions...)
}

// handleIngestRequested executes ingest materialization for a requested schema payload batch.
func handleIngestRequested(ctx context.Context, db *ent.Client, payload RequestedPayload, mappingIndex types.MappingIndex) error {
	if len(payload.Envelopes) == 0 {
		return nil
	}

	integrationID := payload.IntegrationID
	if integrationID == "" {
		return ErrIngestIntegrationRequired
	}

	schema, ok := schemaNameToMappingSchema(payload.Schema)
	if !ok {
		return ErrIngestSchemaUnsupported
	}

	// privacy.Allow is required here because ingest is a system-level operation executed from a
	// background Gala listener that carries no user authentication context
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
	if !supportsSchemaIngest(mappingIndex, provider, integrationRecord.Config, schema) {
		return ErrMappingNotFound
	}

	ingestHandler, ok := HandlerForSchema(schema)
	if !ok {
		return ErrIngestSchemaUnsupported
	}

	operationName := operationNameForRequestedPayload(payload, schema)
	operationConfig := map[string]any{}
	if operationName != "" {
		operationConfig, err = operations.ResolveOperationConfig(&integrationRecord.Config, string(operationName), nil)
		if err != nil {
			return err
		}
	}

	result, err := ingestHandler(allowCtx, IngestRequest{
		OrgID:             integrationRecord.OwnerID,
		IntegrationID:     integrationRecord.ID,
		Provider:          provider,
		Operation:         operationName,
		IntegrationConfig: integrationRecord.Config,
		ProviderState:     integrationRecord.ProviderState,
		OperationConfig:   operationConfig,
		MappingIndex:      mappingIndex,
		Envelopes:         payload.Envelopes,
		DB:                db,
	})
	if err != nil {
		return err
	}

	log.Ctx(allowCtx).Debug().Str("integration_id", integrationID).Str("schema", payload.Schema).Int("total", result.Summary.Total).Int("created", result.Summary.Created).Int("updated", result.Summary.Updated).Int("skipped", result.Summary.Skipped).Int("failed", result.Summary.Failed).Msg("ingest handler completed")

	return nil
}

// operationNameForRequestedPayload resolves the effective operation name for an ingest payload.
func operationNameForRequestedPayload(payload RequestedPayload, schema types.MappingSchema) types.OperationName {
	if payload.Operation != "" {
		return types.OperationName(payload.Operation)
	}

	contract, ok := integrationgenerated.IntegrationIngestSchemas[string(schema)]
	if !ok {
		return ""
	}

	return types.OperationName(contract.DefaultOperation)
}
