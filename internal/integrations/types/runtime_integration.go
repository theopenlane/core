package types //nolint:revive

import (
	"context"
	"encoding/json"
)

// RuntimeIntegrationID is the stable identifier for a runtime integration config, derived from the Go type name
type RuntimeIntegrationID struct {
	// name is the stable string identifier derived from the reflected schema
	name string
}

// NewRuntimeIntegrationID creates a runtime integration identity handle with a stable name
func NewRuntimeIntegrationID(name string) RuntimeIntegrationID {
	return RuntimeIntegrationID{name: name}
}

// Name returns the stable identifier
func (id RuntimeIntegrationID) Name() string {
	return id.name
}

// String returns the stable identifier
func (id RuntimeIntegrationID) String() string {
	return id.name
}

// Valid reports whether the ID was initialized
func (id RuntimeIntegrationID) Valid() bool {
	return id.name != ""
}

// RuntimeIntegrationRef is a typed reference to a runtime integration config.
// When populated, the definition operates entirely in memory with no Integration
// DB record, no keystore credentials, and no connection lifecycle
type RuntimeIntegrationRef[T any] struct {
	// id is the stable identifier derived from the reflected schema
	id RuntimeIntegrationID
	// schema is the reflected JSON schema of the config type
	schema json.RawMessage
	// config is the runtime config, nil when not provisioned
	config *T
}

// NewRuntimeIntegrationRef creates a typed runtime integration ref with the given name and schema
func NewRuntimeIntegrationRef[T any](name string, schema json.RawMessage) RuntimeIntegrationRef[T] {
	return RuntimeIntegrationRef[T]{
		id:     NewRuntimeIntegrationID(name),
		schema: schema,
	}
}

// ID returns the stable identifier
func (r RuntimeIntegrationRef[T]) ID() RuntimeIntegrationID {
	return r.id
}

// Schema returns the reflected JSON schema
func (r RuntimeIntegrationRef[T]) Schema() json.RawMessage {
	return r.schema
}

// SetConfig sets the runtime config. When set, the registry will call Build at registration time
func (r *RuntimeIntegrationRef[T]) SetConfig(cfg *T) {
	r.config = cfg
}

// Config returns the runtime config, if set
func (r RuntimeIntegrationRef[T]) Config() *T {
	return r.config
}

// Provisioned reports whether runtime config has been provided
func (r RuntimeIntegrationRef[T]) Provisioned() bool {
	return r.config != nil
}

// MarshalConfig marshals the config to JSON for passing to the Build function
func (r RuntimeIntegrationRef[T]) MarshalConfig() (json.RawMessage, error) {
	if r.config == nil {
		return nil, nil
	}

	return json.Marshal(r.config)
}

// RuntimeIntegrationRegistration is the non-generic registration stored on a Definition
type RuntimeIntegrationRegistration struct {
	// Ref is the stable identifier for the runtime config type
	Ref RuntimeIntegrationID `json:"ref"`
	// Schema is the reflected JSON schema of the runtime config struct
	Schema json.RawMessage `json:"schema,omitempty"`
	// Config is the marshaled runtime config, nil when not provisioned
	Config json.RawMessage `json:"config,omitempty"`
	// Build constructs the client from the runtime config.
	// Called once at startup when Config is non-nil. The returned client
	// is cached for the lifetime of the process
	Build func(ctx context.Context, config json.RawMessage) (any, error) `json:"-"`
}
