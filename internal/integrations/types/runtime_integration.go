package types //nolint:revive

import (
	"context"
	"encoding/json"
)

// RuntimeIntegrationRegistration declares a definition provisioned in memory from one runtime config struct, with no Integration DB record, no keystore credentials, and no connection lifecycle
type RuntimeIntegrationRegistration struct {
	// Schema is the reflected JSON schema of the runtime config struct
	Schema json.RawMessage `json:"schema,omitempty"`
	// Config is the marshaled runtime config, nil when not provisioned
	Config json.RawMessage `json:"config,omitempty"`
	// Build constructs the client from the runtime config once at registration when Config is non-nil; the registry caches the result for the lifetime of the process
	Build func(ctx context.Context, config json.RawMessage) (any, error) `json:"-"`
}
