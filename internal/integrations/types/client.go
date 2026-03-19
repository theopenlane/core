package types

import (
	"context"
	"encoding/json"

	generated "github.com/theopenlane/core/internal/ent/generated"
)

// ClientBuildRequest bundles the inputs for building one installation-scoped client
type ClientBuildRequest struct {
	// Installation is the target installation record
	Installation *generated.Integration
	// Credential is the primary credential bundle for convenience when the client uses one slot
	Credential CredentialSet
	// Credentials lists all resolved credential bundles for this client by slot ref
	Credentials CredentialBindings
	// Config is the client-specific configuration payload
	Config json.RawMessage
}

// ClientBuilderFunc builds a client for one installation
type ClientBuilderFunc func(ctx context.Context, req ClientBuildRequest) (any, error)

// ClientRegistration declares one buildable client for a definition
type ClientRegistration struct {
	// Ref is the internal client identity used for operation/runtime lookup
	Ref ClientID `json:"-"`
	// CredentialRefs identifies which durable credential slots this client may use
	CredentialRefs []CredentialRef `json:"credentialRefs,omitempty"`
	// Description describes what the client is used for
	Description string `json:"description,omitempty"`
	// ConfigSchema is the JSON schema for client-specific configuration
	ConfigSchema json.RawMessage `json:"configSchema,omitempty"`
	// Build constructs the client
	Build ClientBuilderFunc `json:"-"`
}
