package types //nolint:revive

import (
	"context"

	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/pkg/jsonx"
)

// ExecutionMetadata is the typed durable context attached to integration execution paths
type ExecutionMetadata struct {
	// OwnerID identifies the owning organization for the execution
	OwnerID string `json:"ownerId,omitempty"`
	// IntegrationID identifies the target integration installation
	IntegrationID string `json:"installationId,omitempty"`
	// DefinitionID identifies the integration definition
	DefinitionID string `json:"definitionId,omitempty"`
	// Operation identifies the definition-local operation being run
	Operation string `json:"operation,omitempty"`
	// RunID identifies the persisted integration run when present
	RunID string `json:"runId,omitempty"`
	// RunType identifies the integration run classification
	RunType enums.IntegrationRunType `json:"runType,omitempty"`
	// Webhook identifies the triggering webhook when present
	Webhook string `json:"webhook,omitempty"`
	// Event identifies the triggering webhook event when present
	Event string `json:"event,omitempty"`
	// DeliveryID identifies the provider delivery/event id when present
	DeliveryID string `json:"deliveryId,omitempty"`
	// Workflow captures workflow linkage when present
	Workflow *WorkflowMeta `json:"workflow,omitempty"`
}

// ExecutionMetadataKey stores durable integration execution metadata on a context
var ExecutionMetadataKey = contextx.NewKey[ExecutionMetadata]()

// WithExecutionMetadata stores metadata on the supplied context
func WithExecutionMetadata(ctx context.Context, metadata ExecutionMetadata) context.Context {
	return ExecutionMetadataKey.Set(ctx, metadata)
}

// ExecutionMetadataFromContext returns integration execution metadata from context when present
func ExecutionMetadataFromContext(ctx context.Context) (ExecutionMetadata, bool) {
	return ExecutionMetadataKey.Get(ctx)
}

// Properties returns metadata as a string map suitable for gala header properties
func (m ExecutionMetadata) Properties() map[string]string {
	raw, _ := jsonx.ToMap(m)
	out := make(map[string]string, len(raw))

	for k, v := range raw {
		if s, ok := v.(string); ok && s != "" {
			out[k] = s
		}
	}

	return out
}
