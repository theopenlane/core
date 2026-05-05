package types //nolint:revive

import (
	"context"
	"encoding/json"
)

// RuntimeRefID is the stable identifier for a runtime integration config, derived from the Go type name
type RuntimeRefID struct {
	// name is the stable string identifier derived from the reflected schema
	name string
}

// NewRuntimeRefID creates a runtime integration identity handle with a stable name
func NewRuntimeRefID(name string) RuntimeRefID {
	return RuntimeRefID{name: name}
}

// Name returns the stable identifier
func (id RuntimeRefID) Name() string {
	return id.name
}

// String returns the stable identifier
func (id RuntimeRefID) String() string {
	return id.name
}

// Valid reports whether the ID was initialized
func (id RuntimeRefID) Valid() bool {
	return id.name != ""
}

// RuntimeRef is a typed reference to a runtime integration config.
// When populated, the definition operates entirely in memory with no Integration
// DB record, no keystore credentials, and no connection lifecycle
type RuntimeRef[T any] struct {
	// id is the stable identifier derived from the reflected schema
	id RuntimeRefID
	// schema is the reflected JSON schema of the config type
	schema json.RawMessage
	// config is the runtime config, nil when not provisioned
	config *T
}

// NewRuntimeRef creates a typed runtime integration ref with the given name and schema
func NewRuntimeRef[T any](name string, schema json.RawMessage) RuntimeRef[T] {
	return RuntimeRef[T]{
		id:     NewRuntimeRefID(name),
		schema: schema,
	}
}

// ID returns the stable identifier
func (r RuntimeRef[T]) ID() RuntimeRefID {
	return r.id
}

// Schema returns the reflected JSON schema
func (r RuntimeRef[T]) Schema() json.RawMessage {
	return r.schema
}

// SetConfig sets the runtime config. When set, the registry will call Build at registration time
func (r *RuntimeRef[T]) SetConfig(cfg *T) {
	r.config = cfg
}

// Config returns the runtime config, if set
func (r RuntimeRef[T]) Config() *T {
	return r.config
}

// Provisioned reports whether runtime config has been provided
func (r RuntimeRef[T]) Provisioned() bool {
	return r.config != nil
}

// MarshalConfig marshals the config to JSON for passing to the Build function
func (r RuntimeRef[T]) MarshalConfig() (json.RawMessage, error) {
	if r.config == nil {
		return nil, nil
	}

	return json.Marshal(r.config)
}

// RuntimeIntegrationRegistration is the non-generic registration stored on a Definition
type RuntimeIntegrationRegistration struct {
	// Ref is the stable identifier for the runtime config type
	Ref RuntimeRefID `json:"ref"`
	// Schema is the reflected JSON schema of the runtime config struct
	Schema json.RawMessage `json:"schema,omitempty"`
	// Config is the marshaled runtime config, nil when not provisioned
	Config json.RawMessage `json:"config,omitempty"`
	// Build constructs the client from the runtime config.
	// Called once at startup when Config is non-nil. The returned client
	// is cached for the lifetime of the process
	Build func(ctx context.Context, config json.RawMessage) (any, error) `json:"-"`
}
