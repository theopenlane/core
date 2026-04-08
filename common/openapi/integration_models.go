package openapi

import (
	"bytes"
	"encoding/json"

	"github.com/theopenlane/utils/rout"
)

// IntegrationProviderMetadata is a snapshot of definition metadata captured on installation
type IntegrationProviderMetadata struct {
	// Name is the provider's unique name
	Name string `json:"name"`
	// DisplayName is the human-readable provider name
	DisplayName string `json:"displayName"`
	// Category is the provider category
	Category string `json:"category"`
	// Description is the provider description
	Description string `json:"description,omitempty"`
	// AuthType is the authentication type
	AuthType string `json:"authType"`
	// AuthStartPath is the path to start authentication
	AuthStartPath string `json:"authStartPath,omitempty"`
	// AuthCallbackPath is the path for authentication callback
	AuthCallbackPath string `json:"authCallbackPath,omitempty"`
	// Active indicates if the provider is active
	Active bool `json:"active"`
	// Visible indicates if the provider is visible
	Visible bool `json:"visible"`
	// Tags is a list of provider tags
	Tags []string `json:"tags,omitempty"`
	// LogoURL is the URL to the provider logo
	LogoURL string `json:"logoUrl,omitempty"`
	// DocsURL is the URL to the provider documentation
	DocsURL string `json:"docsUrl,omitempty"`
	// EnvironmentCredentials is the environment credentials JSON
	EnvironmentCredentials json.RawMessage `json:"environmentCredentials,omitempty"`
	// CredentialsSchema is the credentials schema JSON
	CredentialsSchema json.RawMessage `json:"credentialsSchema,omitempty"`
	// Persistence is the persistence configuration
	Persistence map[string]any `json:"persistence,omitempty"`
	// Labels is a set of provider labels
	Labels map[string]string `json:"labels,omitempty"`
}

// IntegrationConfig is the per-installation runtime configuration stored as a typed JSON field
type IntegrationConfig struct {
	// ClientConfig is the client configuration JSON
	ClientConfig json.RawMessage `json:"clientConfig,omitempty"`
}

// IntegrationInstallationIdentity is the normalized, provider-agnostic installation identity
// surfaced in the GraphQL metadata field for UI display
type IntegrationInstallationIdentity struct {
	// ExternalName is the human-readable name of the external entity
	// (e.g. Slack workspace name, Azure tenant display name, Google Workspace domain)
	ExternalName string `json:"externalName,omitempty"`
	// ExternalID is the machine identifier for the external entity
	// (e.g. Slack team ID, Azure tenant ID, AWS account ID, GitHub installation ID)
	ExternalID string `json:"externalId,omitempty"`
	// CredentialRef is the credential method used to set up the integration
	CredentialRef string `json:"credentialRef,omitempty"`
	// LastSuccessfulHealthCheck is the RFC3339 timestamp of the last successful validation operation
	LastSuccessfulHealthCheck string `json:"lastSuccessfulHealthCheck,omitempty"`
}

// IntegrationInstallationMetadata stores stable, non-secret installation identity metadata
type IntegrationInstallationMetadata struct {
	// Attributes is the provider-defined installation metadata payload
	Attributes json.RawMessage `json:"attributes,omitempty"`
	// Display is the normalized installation identity for UI rendering
	Display IntegrationInstallationIdentity `json:"display,omitzero"`
}

// IntegrationProviderState stores provider-specific integration state captured during auth and config
type IntegrationProviderState struct {
	// Providers is a map of provider keys to raw state JSON
	Providers map[string]json.RawMessage `json:"providers,omitempty"`
}

// ConfigureIntegrationRequest is the request type for configuring a non-OAuth provider.
type ConfigureIntegrationRequest struct {
	// DefinitionID is the canonical integration definition ID from the path.
	DefinitionID string `param:"definitionID" description:"Integration definition ID" example:"def_01K0GCPSCC00000000000000001"`
	// IntegrationID is the optional existing installation to update credentials on; when omitted we create a new integration.
	IntegrationID string `json:"integrationId,omitempty"`
	// CredentialRef selects which credential slot is being configured.
	CredentialRef string `json:"credentialRef"`
	// Body holds the provider-specific credential fields as a raw JSON object.
	Body json.RawMessage `json:"body"`
	// UserInput holds optional installation-scoped provider configuration.
	UserInput json.RawMessage `json:"userInput,omitempty"`
}

// RunIntegrationOperationBody is the request body for triggering a provider operation.
type RunIntegrationOperationBody struct {
	// Operation is the operation identifier.
	Operation string `json:"operation"`
	// Config holds optional operation-specific configuration as a raw JSON object.
	Config json.RawMessage `json:"config,omitempty"`
}

// RunIntegrationOperationRequest is the request type for running an integration operation.
type RunIntegrationOperationRequest struct {
	// IntegrationID is the installation record to run the operation against.
	IntegrationID string `param:"integrationID" description:"Integration installation ID" example:"01J4HMNDSZCCQBTY93BF9CBF5D"`
	// Body holds the operation name and optional configuration.
	Body RunIntegrationOperationBody `json:"body"`
}

// ConfigureIntegrationResponse is the response after successfully configuring a provider.
type ConfigureIntegrationResponse struct {
	rout.Reply
	// Provider is the configured definition ID.
	Provider string `json:"provider"`
	// IntegrationID is the installation record ID that was created or updated.
	IntegrationID string `json:"integrationId"`
	// HealthStatus is the result of the inline validation health check.
	HealthStatus string `json:"healthStatus,omitempty"`
	// HealthSummary is a human-readable description of the health check result.
	HealthSummary string `json:"healthSummary,omitempty"`
	// InstallationMetadata is the provider-specific installation identity metadata resolved during configuration (e.g. Slack team name, GitHub org, AWS account).
	InstallationMetadata json.RawMessage `json:"installationMetadata,omitempty"`
	// WebhookEndpointURL is the user-facing webhook or SCIM endpoint URL when the definition declares webhooks.
	WebhookEndpointURL string `json:"webhookEndpointUrl,omitempty"`
	// WebhookSecret is the shared secret for authenticating inbound webhook or SCIM deliveries.
	// This value is only populated on initial creation and should be captured by the caller.
	WebhookSecret string `json:"webhookSecret,omitempty"`
}

// RunIntegrationOperationResponse is the response after executing or queuing a provider operation.
type RunIntegrationOperationResponse struct {
	rout.Reply
	// Provider is the integration definition ID.
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

// IntegrationAuthStartRequest is the request type for starting an integration auth flow.
type IntegrationAuthStartRequest struct {
	// DefinitionID is the canonical integration definition identifier.
	DefinitionID string `json:"definitionId" description:"Integration definition ID" example:"def_01K0SLACK000000000000000001"`
	// IntegrationID is the existing installation to start the auth flow for; when omitted a new installation is created.
	IntegrationID string `json:"integrationId,omitempty"`
	// CredentialRef selects which credential-schema-defined connection is being activated.
	CredentialRef string `json:"credentialRef"`
	// UserInput holds optional installation-scoped provider configuration.
	UserInput json.RawMessage `json:"userInput,omitempty"`
}

// Validate validates the ConfigureIntegrationRequest.
func (r *ConfigureIntegrationRequest) Validate() error {
	if r.DefinitionID == "" {
		return rout.NewMissingRequiredFieldError("definitionId")
	}

	if r.HasCredentialBody() && r.CredentialRef == "" {
		return rout.NewMissingRequiredFieldError("credentialRef")
	}

	return nil
}

// HasCredentialBody reports whether the request includes a meaningful credential payload.
func (r ConfigureIntegrationRequest) HasCredentialBody() bool {
	return hasIntegrationCredentialBody(r.Body)
}

// Validate validates the IntegrationAuthStartRequest.
func (r *IntegrationAuthStartRequest) Validate() error {
	if r.DefinitionID == "" {
		return rout.NewMissingRequiredFieldError("definitionId")
	}

	if r.CredentialRef == "" {
		return rout.NewMissingRequiredFieldError("credentialRef")
	}

	return nil
}

// ExampleIntegrationAuthStartRequest is an example auth start request for OpenAPI documentation.
var ExampleIntegrationAuthStartRequest = IntegrationAuthStartRequest{
	DefinitionID: "def_01K0SLACK000000000000000001",
}

// ExampleConfigureIntegrationRequest is an example configuration payload for OpenAPI documentation.
var ExampleConfigureIntegrationRequest = ConfigureIntegrationRequest{
	DefinitionID: "def_01K0GCPSCC00000000000000001",
	Body:         json.RawMessage(`{"organizationId":"123456789","serviceAccountKey":"{\"type\":\"service_account\",\"project_id\":\"my-project\"}"}`),
}

// ExampleRunIntegrationOperationRequest is an example operation payload for OpenAPI documentation.
var ExampleRunIntegrationOperationRequest = RunIntegrationOperationRequest{
	IntegrationID: "01J4HMNDSZCCQBTY93BF9CBF5D",
	Body: RunIntegrationOperationBody{
		Operation: "HealthCheck",
	},
}

// hasIntegrationCredentialBody reports whether the config request contains a meaningful
// credential payload. Empty objects are treated as absent so user-input-only updates can
// send `body: {}` without triggering credential validation.
func hasIntegrationCredentialBody(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)

	return len(trimmed) > 0 && !bytes.Equal(trimmed, []byte("null")) && !bytes.Equal(trimmed, []byte("{}"))
}
