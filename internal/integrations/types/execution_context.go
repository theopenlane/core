package types //nolint:revive

import (
	"context"

	"github.com/theopenlane/utils/contextx"
)

// IntegrationExecutionContext captures durable integration runtime metadata.
// It intentionally carries identifiers only (no credential secrets).
type IntegrationExecutionContext struct {
	// OrgID identifies the organization executing the integration operation.
	OrgID string `json:"org_id,omitempty"`
	// IntegrationID identifies the integration record used for the run.
	IntegrationID string `json:"integration_id,omitempty"`
	// Provider identifies the integration provider.
	Provider ProviderType `json:"provider,omitempty"`
	// AuthKind identifies the credential auth kind when known.
	AuthKind AuthKind `json:"auth_kind,omitempty"`
	// RunID identifies the integration run record when present.
	RunID string `json:"run_id,omitempty"`
	// Operation identifies the operation being executed when present.
	Operation OperationName `json:"operation,omitempty"`
}

var integrationExecutionContextKey = contextx.NewKey[IntegrationExecutionContext]()

// IntegrationExecutionContextKey returns the typed context key for integration execution metadata.
func IntegrationExecutionContextKey() contextx.Key[IntegrationExecutionContext] {
	return integrationExecutionContextKey
}

// WithIntegrationExecutionContext stores integration execution metadata on the context.
func WithIntegrationExecutionContext(ctx context.Context, value IntegrationExecutionContext) context.Context {
	return integrationExecutionContextKey.Set(ctx, value)
}

// IntegrationExecutionContextFromContext retrieves integration execution metadata from the context.
func IntegrationExecutionContextFromContext(ctx context.Context) (IntegrationExecutionContext, bool) {
	return integrationExecutionContextKey.Get(ctx)
}
