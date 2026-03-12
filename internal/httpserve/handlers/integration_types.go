package handlers

import (
	"encoding/json"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/theopenlane/utils/rout"

	integrationsv2types "github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// IntegrationConfigBody is a raw JSON object body for non-OAuth provider configuration.
// It accepts arbitrary key-value pairs dictated by each provider's credentials schema.
type IntegrationConfigBody json.RawMessage

// ToMap converts the body to a map[string]any for schema validation and attribute access.
func (b IntegrationConfigBody) ToMap() map[string]any {
	m, _ := jsonx.ToMap(json.RawMessage(b))
	if m == nil {
		return map[string]any{}
	}

	return m
}

// MarshalJSON implements json.Marshaler.
func (b IntegrationConfigBody) MarshalJSON() ([]byte, error) {
	return json.RawMessage(b).MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *IntegrationConfigBody) UnmarshalJSON(data []byte) error {
	return (*json.RawMessage)(b).UnmarshalJSON(data)
}


// IntegrationConfigPayload is the request type for configuring a non-OAuth provider.
type IntegrationConfigPayload struct {
	// Provider is the provider key.
	Provider string `param:"provider" description:"Integration provider key" example:"github"`
	// Body holds the provider-specific credential fields as a raw JSON object.
	Body IntegrationConfigBody `json:"body"`
}

// IntegrationOperationBody is the request body for triggering a provider operation.
type IntegrationOperationBody struct {
	// Operation is the operation identifier.
	Operation string `json:"operation"`
	// Config holds optional operation-specific configuration as a raw JSON object.
	Config json.RawMessage `json:"config,omitempty"`
	// Force bypasses idempotency guards when true.
	Force bool `json:"force,omitempty"`
}

// IntegrationOperationPayload is the request type for running an integration operation.
type IntegrationOperationPayload struct {
	// Provider is the provider key.
	Provider string `param:"provider" description:"Integration provider key" example:"github"`
	// IntegrationID scopes the operation to a specific integration record.
	IntegrationID string `query:"integration_id,omitempty" description:"Optional integration ID" example:"01J4HMNDSZCCQBTY93BF9CBF5D"`
	// Body holds the operation name and optional configuration.
	Body IntegrationOperationBody `json:"body"`
}

// IntegrationConfigResponse is the response after successfully configuring a provider.
type IntegrationConfigResponse struct {
	rout.Reply
	// Provider is the configured provider key.
	Provider string `json:"provider"`
}

// IntegrationTokenResponse is the response containing a refreshed integration access token.
// Token fields are flattened directly onto the response.
type IntegrationTokenResponse struct {
	rout.Reply
	// Provider is the integration provider key.
	Provider string `json:"provider"`
	// AccessToken is the OAuth access token.
	AccessToken string `json:"accessToken"`
	// ExpiresAt is the token expiry timestamp when available.
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// IntegrationOperationResponse is the response after executing or queuing a provider operation.
type IntegrationOperationResponse struct {
	rout.Reply
	// Provider is the integration provider key.
	Provider string `json:"provider"`
	// Operation is the operation identifier that was executed.
	Operation string `json:"operation"`
	// Status is the operation result status (e.g. ok, queued, error).
	Status string `json:"status"`
	// Summary is a human-readable description of the result.
	Summary string `json:"summary,omitempty"`
	// Details holds structured result data specific to the operation.
	Details json.RawMessage `json:"details,omitempty"`
}

// IntegrationProvidersResponse is the response listing available integration providers.
type IntegrationProvidersResponse struct {
	rout.Reply
	// Schema is the JSON schema for provider specification objects; used by the UI to
	// construct available provider cards, surface connect/configure buttons, and build
	// documentation links.
	Schema *jsonschema.Schema `json:"schema,omitempty"`
	// Providers is the list of available integration providers.
	Providers []integrationsv2types.IntegrationProviderMetadata `json:"providers"`
}

// ExampleIntegrationConfigPayload is an example configuration payload for OpenAPI documentation.
var ExampleIntegrationConfigPayload = IntegrationConfigPayload{
	Provider: "google",
	Body:     IntegrationConfigBody(`{"serviceAccountKey":"{\"type\":\"service_account\",\"project_id\":\"my-project\"}"}`),
}

// ExampleIntegrationOperationPayload is an example operation payload for OpenAPI documentation.
var ExampleIntegrationOperationPayload = IntegrationOperationPayload{
	Provider: "github",
	Body: IntegrationOperationBody{
		Operation: "health",
	},
}
