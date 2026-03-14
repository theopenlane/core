package types

import (
	"context"
	"encoding/json"
)

// ClientName identifies a specific client type exposed by a provider (e.g. rest, graphql)
type ClientName string

// ClientBuilderFunc constructs provider-specific clients using persisted credentials and optional config
type ClientBuilderFunc func(ctx context.Context, credential CredentialSet, config json.RawMessage) (ClientInstance, error)

// ClientInstance wraps a provider client instance for operation execution
type ClientInstance struct {
	raw any
}

// NewClientInstance creates a client instance wrapper from a concrete client value
func NewClientInstance(raw any) ClientInstance {
	return ClientInstance{raw: raw}
}

// EmptyClientInstance returns a zero client wrapper
func EmptyClientInstance() ClientInstance {
	return ClientInstance{}
}

// ClientInstanceAs unwraps a wrapped client value as a concrete type
func ClientInstanceAs[T any](client ClientInstance) (T, bool) {
	value, ok := client.raw.(T)
	if ok {
		return value, true
	}

	var zero T

	return zero, false
}

// ClientDescriptor describes a provider-managed client that can be pooled and reused downstream
type ClientDescriptor struct {
	// Provider identifies which provider offers this client
	Provider ProviderType
	// Name is the unique client identifier
	Name ClientName
	// Description explains what the client does
	Description string
	// Build is the function that constructs the client
	Build ClientBuilderFunc
	// ConfigSchema defines the JSON schema for client configuration
	ConfigSchema json.RawMessage
}

// ClientRequest contains the parameters required to request a client instance
type ClientRequest struct {
	// OrgID identifies the organization requesting the client
	OrgID string
	// Provider identifies which provider to use
	Provider ProviderType
	// Client identifies which client type to build
	Client ClientName
	// Config contains client-specific configuration
	Config json.RawMessage
	// Force bypasses cached client instances
	Force bool
}
