package observability

import (
	"context"

	"github.com/rs/zerolog"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	// FieldIntegrationID is the log field key for the integration identifier
	FieldIntegrationID = "integration_id"
	// FieldDefinitionID is the log field key for the integration definition identifier
	FieldDefinitionID = "definition_id"
	// FieldOperation is the log field key for the operation name
	FieldOperation = "operation"
	// FieldRunID is the log field key for the run identifier
	FieldRunID = "run_id"
)

// IntegrationRef carries integration identity for structured log embedding
type IntegrationRef struct {
	// IntegrationID is the integration instance identifier
	IntegrationID string
	// DefinitionID is the integration definition (provider type) identifier
	DefinitionID string
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler
func (r IntegrationRef) MarshalZerologObject(e *zerolog.Event) {
	e.Str(FieldIntegrationID, r.IntegrationID)
	e.Str(FieldDefinitionID, r.DefinitionID)
}

// FromIntegration builds an IntegrationRef from an ent Integration
func FromIntegration(integration *ent.Integration) IntegrationRef {
	return IntegrationRef{
		IntegrationID: integration.ID,
		DefinitionID:  integration.DefinitionID,
	}
}

// WithIntegration enriches the context logger with integration identity fields so
// that all downstream log calls via logx.FromContext automatically include them
func WithIntegration(ctx context.Context, integrationID, definitionID string) context.Context {
	return logx.WithFields(ctx, map[string]any{
		FieldIntegrationID: integrationID,
		FieldDefinitionID:  definitionID,
	})
}

// WithContext stores execution metadata on context and enriches the logger
// with all non-empty execution fields
func WithContext(ctx context.Context, metadata types.ExecutionMetadata) context.Context {
	ctx = types.WithExecutionMetadata(ctx, metadata)
	ctx = WithIntegration(ctx, metadata.IntegrationID, metadata.DefinitionID)

	if metadata.Operation != "" {
		ctx = WithOperation(ctx, metadata.Operation)
	}

	if metadata.RunID != "" {
		ctx = WithRunID(ctx, metadata.RunID)
	}

	return ctx
}

// WithOperation adds the operation name to the context logger
func WithOperation(ctx context.Context, operation string) context.Context {
	return logx.WithField(ctx, FieldOperation, operation)
}

// WithRunID adds the run identifier to the context logger
func WithRunID(ctx context.Context, runID string) context.Context {
	return logx.WithField(ctx, FieldRunID, runID)
}

