package types

import (
	"context"
	"encoding/json"

	generated "github.com/theopenlane/core/internal/ent/generated"
)

// ClientName identifies a named client within one definition
type ClientName string

// ClientBuildRequest bundles the inputs for building one installation-scoped client
type ClientBuildRequest struct {
	// Installation is the target installation record
	Installation *generated.Integration
	// Credential is the installation-scoped credential bundle
	Credential CredentialSet
	// Config is the client-specific configuration payload
	Config json.RawMessage
}

// ClientBuilderFunc builds a client for one installation
type ClientBuilderFunc func(ctx context.Context, req ClientBuildRequest) (any, error)

// ClientRegistration declares one buildable client for a definition
type ClientRegistration struct {
	// Name is the stable client identifier within the definition
	Name ClientName `json:"name"`
	// Description describes what the client is used for
	Description string `json:"description,omitempty"`
	// ConfigSchema is the JSON schema for client-specific configuration
	ConfigSchema json.RawMessage `json:"configSchema,omitempty"`
	// Build constructs the client
	Build ClientBuilderFunc `json:"-"`
}
