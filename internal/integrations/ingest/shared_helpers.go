package ingest

import (
	"context"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	integrationstate "github.com/theopenlane/core/internal/integrations/state"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// schemaIngestContext captures shared mapping state used by schema-specific ingest handlers.
type schemaIngestContext struct {
	mapper               *MappingEvaluator
	schema               integrationgenerated.IntegrationMappingSchema
	overrideIndex        mappingOverrideIndex
	integrationConfigMap map[string]any
	providerStateMap     map[string]any
	mappingIndex         integrationtypes.MappingIndex
}

// envelopeMappingRequest captures shared request fields used during envelope mapping.
type envelopeMappingRequest struct {
	Provider        integrationtypes.ProviderType
	Operation       integrationtypes.OperationName
	OrgID           string
	IntegrationID   string
	OperationConfig map[string]any
}

// envelopeProcessFunc processes one envelope and reports whether it was ingested and whether it created a new record.
type envelopeProcessFunc func(envelope integrationtypes.AlertEnvelope) (ingested bool, created bool, err error)

// newSchemaIngestContext builds shared mapping context for one schema ingest execution.
func newSchemaIngestContext(integrationConfig openapi.IntegrationConfig, providerState integrationstate.IntegrationProviderState, mappingIndex integrationtypes.MappingIndex, schemaName string) (schemaIngestContext, error) {
	mapper, err := NewMappingEvaluator()
	if err != nil {
		return schemaIngestContext{}, err
	}

	schema, ok := integrationgenerated.IntegrationMappingSchemas[schemaName]
	if !ok {
		return schemaIngestContext{}, ErrMappingSchemaNotFound
	}

	integrationConfigMap, err := jsonx.ToMap(integrationConfig)
	if err != nil {
		return schemaIngestContext{}, err
	}
	providerStateMap, err := jsonx.ToMap(providerState)
	if err != nil {
		return schemaIngestContext{}, err
	}

	return schemaIngestContext{
		mapper:               mapper,
		schema:               schema,
		overrideIndex:        newMappingOverrideIndex(integrationConfig),
		integrationConfigMap: integrationConfigMap,
		providerStateMap:     providerStateMap,
		mappingIndex:         mappingIndex,
	}, nil
}

// mapIngestEnvelope evaluates mapping rules for one envelope using shared ingest context.
func mapIngestEnvelope(ctx context.Context, ingestCtx schemaIngestContext, req envelopeMappingRequest, schemaName string, envelope integrationtypes.AlertEnvelope) (map[string]any, bool, error) {
	payloadMap, err := decodeAlertPayload(envelope.Payload)
	if err != nil {
		return nil, false, err
	}

	spec, ok := resolveMappingSpecWithIndex(ingestCtx.overrideIndex, ingestCtx.mappingIndex, req.Provider, schemaName, envelope.AlertType)
	if !ok {
		return nil, false, ErrMappingNotFound
	}

	vars := MappingVars{
		Payload:           payloadMap,
		Resource:          envelope.Resource,
		AlertType:         envelope.AlertType,
		Provider:          req.Provider,
		Operation:         req.Operation,
		OrgID:             req.OrgID,
		IntegrationID:     req.IntegrationID,
		Config:            req.OperationConfig,
		IntegrationConfig: ingestCtx.integrationConfigMap,
		ProviderState:     ingestCtx.providerStateMap,
	}.Map()

	allowed, err := ingestCtx.mapper.EvaluateFilter(ctx, spec.FilterExpr, vars)
	if err != nil {
		return nil, false, err
	}
	if !allowed {
		return nil, false, nil
	}

	mapped, err := ingestCtx.mapper.EvaluateMap(ctx, spec.MapExpr, vars)
	if err != nil {
		return nil, false, err
	}

	mapped = filterMappingOutput(ingestCtx.schema, mapped)
	if err := validateMappingOutput(ingestCtx.schema, mapped); err != nil {
		return nil, false, err
	}

	return mapped, true, nil
}

// processIngestEnvelopes runs ingest processing for each envelope and accumulates summary counters and errors.
func processIngestEnvelopes(envelopes []integrationtypes.AlertEnvelope, process envelopeProcessFunc) (IngestSummary, []string) {
	summary := IngestSummary{}
	var errors []string

	for _, envelope := range envelopes {
		summary.Total++

		ingested, created, err := process(envelope)
		if err != nil {
			summary.Failed++
			errors = append(errors, err.Error())
			continue
		}
		if !ingested {
			summary.Skipped++
			continue
		}

		summary.Mapped++
		summary.Persisted++
		if created {
			summary.Created++
		} else {
			summary.Updated++
		}
	}

	return summary, errors
}
