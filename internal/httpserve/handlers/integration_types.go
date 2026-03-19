package handlers

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/internal/integrations/types"
)

// IntegrationConfigBody is a raw JSON object body for non-OAuth provider configuration.
// It accepts arbitrary key-value pairs dictated by each provider's credentials schema.
type IntegrationConfigBody json.RawMessage

// RawMessage returns the body as json.RawMessage.
func (b IntegrationConfigBody) RawMessage() json.RawMessage {
	return json.RawMessage(b)
}

// IsNullOrEmptyObject reports whether the body is absent, null, or an empty JSON object.
func (b IntegrationConfigBody) IsNullOrEmptyObject() bool {
	trimmed := bytes.TrimSpace(b)

	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) || bytes.Equal(trimmed, []byte("{}"))
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
	// DefinitionID is the canonical integration definition ID from the path.
	DefinitionID string `param:"definitionID" description:"Integration definition ID" example:"def_01K0GWKSP000000000000000001"`
	// InstallationID is the optional existing installation to update credentials on.
	// When omitted a new installation is created.
	InstallationID string `json:"installationId,omitempty"`
	// Body holds the provider-specific credential fields as a raw JSON object.
	Body IntegrationConfigBody `json:"body"`
	// UserInput holds optional installation-scoped provider configuration.
	UserInput IntegrationConfigBody `json:"userInput,omitempty"`
}

// IntegrationOperationBody is the request body for triggering a provider operation.
type IntegrationOperationBody struct {
	// Operation is the operation identifier.
	Operation string `json:"operation"`
	// Config holds optional operation-specific configuration as a raw JSON object.
	Config json.RawMessage `json:"config,omitempty"`
}

// IntegrationOperationPayload is the request type for running an integration operation.
type IntegrationOperationPayload struct {
	// DefinitionID is the canonical integration definition ID from the path.
	DefinitionID string `param:"definitionID" description:"Integration definition ID" example:"def_01K0GHAPP000000000000000001"`
	// IntegrationID scopes the operation to a specific installation record.
	IntegrationID string `query:"integration_id,omitempty" description:"Optional installation ID" example:"01J4HMNDSZCCQBTY93BF9CBF5D"`
	// Body holds the operation name and optional configuration.
	Body IntegrationOperationBody `json:"body"`
}

// IntegrationConfigResponse is the response after successfully configuring a provider.
type IntegrationConfigResponse struct {
	rout.Reply
	// Provider is the configured definition ID.
	Provider string `json:"provider"`
	// InstallationID is the installation record ID that was created or updated.
	InstallationID string `json:"installationId"`
}

// IntegrationTokenResponse is the response containing a refreshed integration access token.
// Token fields are flattened directly onto the response.
type IntegrationTokenResponse struct {
	rout.Reply
	// Provider is the integration definition ID.
	Provider string `json:"provider"`
	// AccessToken is the OAuth access token.
	AccessToken string `json:"accessToken"`
	// ExpiresAt is the token expiry timestamp when available.
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// IntegrationOperationResponse is the response after executing or queuing a provider operation.
type IntegrationOperationResponse struct {
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

// DefinitionCatalogEntry is the API response shape for one registered integration definition.
type DefinitionCatalogEntry struct {
	// ID is the canonical definition identifier
	ID string `json:"id"`
	// Slug is the human-readable definition alias
	Slug string `json:"slug"`
	// Family is the optional grouping label
	Family string `json:"family,omitempty"`
	// DisplayName is the UI-facing name
	DisplayName string `json:"displayName"`
	// Description is the user-facing description
	Description string `json:"description,omitempty"`
	// Category is the catalog category
	Category string `json:"category,omitempty"`
	// DocsURL links to documentation
	DocsURL string `json:"docsUrl,omitempty"`
	// LogoURL links to a catalog logo asset
	LogoURL string `json:"logoUrl,omitempty"`
	// Tags are optional catalog labels
	Tags []string `json:"tags,omitempty"`
	// Labels stores arbitrary metadata
	Labels map[string]string `json:"labels,omitempty"`
	// Active indicates whether the definition is enabled
	Active bool `json:"active"`
	// Visible indicates whether the definition is visible in catalog surfaces
	Visible bool `json:"visible"`
	// HasAuth indicates whether the definition exposes an auth flow
	HasAuth bool `json:"hasAuth"`
	// Auth describes the install or auth flow exposed by the definition
	Auth *types.AuthRegistration `json:"auth,omitempty"`
	// CredentialSchema is the JSON schema for credential fields
	CredentialSchema json.RawMessage `json:"credentialSchema,omitempty"`
	// OperatorConfig is the JSON schema for operator config
	OperatorConfig json.RawMessage `json:"operatorConfig,omitempty"`
	// UserInputSchema is the JSON schema for installation-scoped user input
	UserInputSchema json.RawMessage `json:"userInputSchema,omitempty"`
	// Operations lists the operations the definition exposes
	Operations []DefinitionOperationEntry `json:"operations,omitempty"`
}

// DefinitionOperationEntry describes one operation exposed by a definition.
type DefinitionOperationEntry struct {
	// Name is the operation identifier
	Name string `json:"name"`
	// Description is the operation description
	Description string `json:"description,omitempty"`
	// ConfigSchema is the JSON schema for operation config
	ConfigSchema json.RawMessage `json:"configSchema,omitempty"`
}

// IntegrationProvidersResponse is the response listing available integration definitions.
type IntegrationProvidersResponse struct {
	rout.Reply
	// Providers is the list of available integration definitions.
	Providers []DefinitionCatalogEntry `json:"providers"`
}

// OAuthFlowRequest is the request type for starting an OAuth auth flow.
type OAuthFlowRequest struct {
	// DefinitionID is the canonical integration definition identifier.
	DefinitionID string `json:"definitionId" description:"Integration definition ID" example:"def_01K0SLACK000000000000000001"`
	// InstallationID is the optional existing installation to start the auth flow for.
	// When omitted a new installation is created.
	InstallationID string `json:"installationId,omitempty"`
	// UserInput holds optional installation-scoped provider configuration.
	UserInput IntegrationConfigBody `json:"userInput,omitempty"`
}

// Validate validates the OAuthFlowRequest.
func (r *OAuthFlowRequest) Validate() error {
	if r.DefinitionID == "" {
		return rout.NewMissingRequiredFieldError("definitionId")
	}

	return nil
}

// ExampleOAuthFlowRequest is an example OAuth flow request for OpenAPI documentation.
var ExampleOAuthFlowRequest = OAuthFlowRequest{
	DefinitionID: "def_01K0SLACK000000000000000001",
}

// RefreshInstallationCredentialRequest is the request for refreshing an installation's OAuth tokens.
type RefreshInstallationCredentialRequest struct {
	// InstallationID is the installation to refresh credentials for.
	InstallationID string `param:"id" json:"installationId" description:"Installation ID" example:"01J4HMNDSZCCQBTY93BF9CBF5D"`
}

// Validate validates the RefreshInstallationCredentialRequest.
func (r *RefreshInstallationCredentialRequest) Validate() error {
	if r.InstallationID == "" {
		return rout.NewMissingRequiredFieldError("installationId")
	}

	return nil
}

// ExampleIntegrationConfigPayload is an example configuration payload for OpenAPI documentation.
var ExampleIntegrationConfigPayload = IntegrationConfigPayload{
	DefinitionID: "def_01K0GWKSP000000000000000001",
	Body:         IntegrationConfigBody(`{"serviceAccountKey":"{\"type\":\"service_account\",\"project_id\":\"my-project\"}"}`),
}

// ExampleIntegrationOperationPayload is an example operation payload for OpenAPI documentation.
var ExampleIntegrationOperationPayload = IntegrationOperationPayload{
	DefinitionID: "def_01K0GHAPP000000000000000001",
	Body: IntegrationOperationBody{
		Operation: "health.default",
	},
}
