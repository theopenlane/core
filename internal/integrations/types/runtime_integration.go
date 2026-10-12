package types //nolint:revive

import (
	"context"
	"encoding/json"
)

// RuntimeIntegrationRegistration declares a definition provisioned from a runtime config struct
type RuntimeIntegrationRegistration struct {
	// Schema is the reflected JSON schema of the runtime config struct
	Schema json.RawMessage `json:"schema,omitempty"`
	// Config is the marshaled runtime config, nil when not provisioned
	Config json.RawMessage `json:"-"`
	// Build constructs the client from the runtime config once; the registry caches the result
	Build func(ctx context.Context, config json.RawMessage) (any, error) `json:"-"`
}
