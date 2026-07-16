package gala

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/pkg/jsonx"
)

// OperationContext is the durable entity-object metadata attached to event
// dispatch and restored on the handling side. Source-specific provenance is
// carried as opaque JSON in Attributes and decoded into the relevant typed
// struct on demand via DecodeAttributes. Authentication context (organization,
// user) travels with auth.Caller and is intentionally not duplicated here.
type OperationContext struct {
	// OwnerID is the owning organization for the operation
	OwnerID string `json:"ownerId,omitempty" jsonschema:"description=Owning organization identifier"`
	// Operation is the mutation type when applicable: CREATE, UPDATE, DELETE
	Operation string `json:"operation,omitempty" jsonschema:"description=Mutation type when applicable: CREATE, UPDATE, DELETE"`
	// EntityID is the identifier of the entity the operation targets
	EntityID string `json:"entityId,omitempty" jsonschema:"description=Identifier of the target entity"`
	// EntityType is the schema type of the target entity
	EntityType string `json:"entityType,omitempty" jsonschema:"description=Schema type of the target entity"`
	// Attributes is the source-specific provenance payload, decoded on demand
	Attributes json.RawMessage `json:"attributes,omitempty" jsonschema:"description=Source-specific provenance payload"`
}

// SetAttributes marshals a source-specific provenance payload onto the context
func SetAttributes[T any](c *OperationContext, attributes T) error {
	raw, err := jsonx.ToRawMessage(attributes)
	if err != nil {
		return err
	}

	c.Attributes = raw

	return nil
}

// DecodeAttributes decodes the source-specific provenance payload into T
func DecodeAttributes[T any](c OperationContext) (T, error) {
	return jsonx.Decode[T](c.Attributes)
}

// Properties returns the context as a flat string map for gala header visibility
func (c OperationContext) Properties() map[string]string {
	raw, _ := jsonx.ToMap(c)
	out := make(map[string]string, len(raw))

	for k, v := range raw {
		if s, ok := v.(string); ok && s != "" {
			out[k] = s
		}
	}

	return out
}

// operationContextCodecID is the durable context codec identity for OperationContext
const operationContextCodecID = ContextKey("operation_context")

// OperationContextKey stores durable entity-object operation metadata on a context
var OperationContextKey = contextx.NewKey[OperationContext]()

// WithOperationContext stores operation metadata on the supplied context
func WithOperationContext(ctx context.Context, oc OperationContext) context.Context {
	return OperationContextKey.Set(ctx, oc)
}

// OperationContextFromContext returns operation metadata from context when present
func OperationContextFromContext(ctx context.Context) (OperationContext, bool) {
	return OperationContextKey.Get(ctx)
}

// OperationContextCodec returns the durable context codec for OperationContext propagation
func OperationContextCodec() ContextCodec {
	return NewKeyCodec(operationContextCodecID, OperationContextKey)
}
