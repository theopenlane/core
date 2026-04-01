package handlers

import (
	"encoding/json"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// IntegrationConfigPayload is the request type for configuring a non-OAuth provider
type IntegrationConfigPayload struct {
	// DefinitionID is the canonical integration definition ID from the path
	DefinitionID string `param:"definitionID" description:"Integration definition ID" example:"def_01K0GCPSCC00000000000000001"`
	// IntegrationID is the optional existing installation to update credentials on; when omitted we create a new integration
	IntegrationID string `json:"integrationId,omitempty"`
	// CredentialRef selects which credential slot is being configured
	CredentialRef types.CredentialSlotID `json:"credentialRef"`
	// Body holds the provider-specific credential fields as a raw JSON object
	Body json.RawMessage `json:"body"`
	// UserInput holds optional installation-scoped provider configuration
	UserInput json.RawMessage `json:"userInput,omitempty"`
}

// IntegrationOperationBody is the request body for triggering a provider operation
type IntegrationOperationBody struct {
	// Operation is the operation identifier
	Operation string `json:"operation"`
	// Config holds optional operation-specific configuration as a raw JSON object
	Config json.RawMessage `json:"config,omitempty"`
}

// IntegrationOperationPayload is the request type for running an integration operation
type IntegrationOperationPayload struct {
	// IntegrationID is the installation record to run the operation against
	IntegrationID string `param:"integrationID" description:"Integration installation ID" example:"01J4HMNDSZCCQBTY93BF9CBF5D"`
	// Body holds the operation name and optional configuration
	Body IntegrationOperationBody `json:"body"`
}

// IntegrationConfigResponse is the response after successfully configuring a provider
type IntegrationConfigResponse struct {
	rout.Reply
	// Provider is the configured definition ID
	Provider string `json:"provider"`
	// IntegrationID is the installation record ID that was created or updated
	IntegrationID string `json:"integrationId"`
	// HealthStatus is the result of the inline validation health check
	HealthStatus string `json:"healthStatus,omitempty"`
	// HealthSummary is a human-readable description of the health check result
	HealthSummary string `json:"healthSummary,omitempty"`
	// InstallationMetadata is the provider-specific installation identity metadata resolved during configuration (e.g. Slack team name, GitHub org, AWS account)
	InstallationMetadata json.RawMessage `json:"installationMetadata,omitempty"`
	// WebhookEndpointURL is the user-facing webhook or SCIM endpoint URL when the definition declares webhooks
	WebhookEndpointURL string `json:"webhookEndpointUrl,omitempty"`
	// WebhookSecret is the shared secret for authenticating inbound webhook or SCIM deliveries
	// This value is only populated on initial creation and should be captured by the caller
	WebhookSecret string `json:"webhookSecret,omitempty"`
}

// IntegrationOperationResponse is the response after executing or queuing a provider operation
type IntegrationOperationResponse struct {
	rout.Reply
	// Provider is the integration definition ID
	Provider string `json:"provider"`
	// Operation is the operation identifier that was executed
	Operation string `json:"operation"`
	// Status is the operation result status (e.g. ok, queued, error)
	Status string `json:"status"`
	// Summary is a human-readable description of the result
	Summary string `json:"summary,omitempty"`
	// Details holds structured result data specific to the operation
	Details json.RawMessage `json:"details,omitempty"`
}

// IntegrationProvidersResponse is the response listing available integration definitions
type IntegrationProvidersResponse struct {
	rout.Reply
	// Providers is the list of available integration definitions
	Providers []types.Definition `json:"providers"`
}

// IntegrationAuthStartRequest is the request type for starting an integration auth flow
type IntegrationAuthStartRequest struct {
	// DefinitionID is the canonical integration definition identifier
	DefinitionID string `json:"definitionId" description:"Integration definition ID" example:"def_01K0SLACK000000000000000001"`
	// IntegrationID is the existing installation to start the auth flow for; when omitted a new installation is created
	IntegrationID string `json:"integrationId,omitempty"`
	// CredentialRef selects which credential-schema-defined connection is being activated
	CredentialRef types.CredentialSlotID `json:"credentialRef"`
	// UserInput holds optional installation-scoped provider configuration
	UserInput json.RawMessage `json:"userInput,omitempty"`
}

// Validate validates the IntegrationConfigPayload
func (r *IntegrationConfigPayload) Validate() error {
	if r.DefinitionID == "" {
		return rout.NewMissingRequiredFieldError("definitionId")
	}

	if !jsonx.IsEmptyRawMessage(r.Body) && r.CredentialRef == (types.CredentialSlotID{}) {
		return rout.NewMissingRequiredFieldError("credentialRef")
	}

	return nil
}

// Validate validates the IntegrationAuthStartRequest
func (r *IntegrationAuthStartRequest) Validate() error {
	if r.DefinitionID == "" {
		return rout.NewMissingRequiredFieldError("definitionId")
	}
	if r.CredentialRef == (types.CredentialSlotID{}) {
		return rout.NewMissingRequiredFieldError("credentialRef")
	}

	return nil
}

// ExampleIntegrationAuthStartRequest is an example auth start request for OpenAPI documentation
var ExampleIntegrationAuthStartRequest = IntegrationAuthStartRequest{
	DefinitionID: "def_01K0SLACK000000000000000001",
}

// ExampleIntegrationConfigPayload is an example configuration payload for OpenAPI documentation
var ExampleIntegrationConfigPayload = IntegrationConfigPayload{
	DefinitionID: "def_01K0GCPSCC00000000000000001",
	Body:         json.RawMessage(`{"organizationId":"123456789","serviceAccountKey":"{\"type\":\"service_account\",\"project_id\":\"my-project\"}"}`),
}

// ExampleIntegrationOperationPayload is an example operation payload for OpenAPI documentation
var ExampleIntegrationOperationPayload = IntegrationOperationPayload{
	IntegrationID: "01J4HMNDSZCCQBTY93BF9CBF5D",
	Body: IntegrationOperationBody{
		Operation: "HealthCheck",
	},
}
