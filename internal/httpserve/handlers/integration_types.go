package handlers

import (
	"encoding/json"

	"github.com/theopenlane/utils/rout"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// ConfigureIntegrationRequest is the request type for configuring a non-OAuth provider.
type ConfigureIntegrationRequest = openapi.ConfigureIntegrationRequest

// RunIntegrationOperationBody is the request body for triggering a provider operation.
type RunIntegrationOperationBody = openapi.RunIntegrationOperationBody

// RunIntegrationOperationRequest is the request type for running an integration operation.
type RunIntegrationOperationRequest = openapi.RunIntegrationOperationRequest

// ConfigureIntegrationResponse is the response after successfully configuring a provider.
type ConfigureIntegrationResponse = openapi.ConfigureIntegrationResponse

// RunIntegrationOperationResponse is the response after executing or queuing a provider operation.
type RunIntegrationOperationResponse = openapi.RunIntegrationOperationResponse

// IntegrationHealthRequest is the request type for running an installation health assessment.
type IntegrationHealthRequest = openapi.IntegrationHealthRequest

// IntegrationConnectionHealth is the connection-level health check outcome.
type IntegrationConnectionHealth = openapi.IntegrationConnectionHealth

// IntegrationOperationHealth is one operation's health outcome.
type IntegrationOperationHealth = openapi.IntegrationOperationHealth

// IntegrationHealthResponse is the response after running an installation health assessment.
type IntegrationHealthResponse = openapi.IntegrationHealthResponse

// IntegrationProvidersResponse is the response listing available integration definitions.
type IntegrationProvidersResponse struct {
	rout.Reply
	// Providers is the list of available integration definitions.
	Providers []IntegrationProvider `json:"providers"`
}

// IntegrationProvider is the provider listing projection of one definition
type IntegrationProvider struct {
	// Spec is the definition spec
	Spec types.DefinitionSpec `json:"spec"`
	// OperatorConfig is the operator config registration
	OperatorConfig *types.OperatorConfigRegistration `json:"operatorConfig,omitempty"`
	// UserInput is the installation user input registration
	UserInput *types.InputRegistration `json:"userInput,omitempty"`
	// CredentialRegistrations is the credential form of each connection
	CredentialRegistrations []IntegrationProviderCredential `json:"credentialRegistrations,omitempty"`
	// Connections is the list of connections
	Connections []IntegrationProviderConnection `json:"connections,omitempty"`
	// Operations is the list of customer selectable operations
	Operations []types.OperationRegistration `json:"operations,omitempty"`
	// Webhooks is the list of webhook registrations
	Webhooks []types.WebhookRegistration `json:"webhooks,omitempty"`
}

// IntegrationProviderCredential is one connection's credential form as the console reads it
type IntegrationProviderCredential struct {
	// Ref is the connection name the credential is stored under
	Ref string `json:"ref"`
	// Name is the user-facing connection name
	Name string `json:"name,omitempty"`
	// Description explains the connection
	Description string `json:"description,omitempty"`
	// Schema is the credential form schema
	Schema json.RawMessage `json:"schema,omitempty"`
	// Recommended marks the recommended connection
	Recommended bool `json:"recommended,omitempty"`
}

// IntegrationProviderConnection is one connection as the console reads it
type IntegrationProviderConnection struct {
	// CredentialRef is the connection name the credential is stored under
	CredentialRef string `json:"credentialRef"`
	// Name is the user-facing connection name
	Name string `json:"name,omitempty"`
	// Description explains the connection
	Description string `json:"description,omitempty"`
	// Meta is additional data needed to set up the connection
	Meta map[string]types.MetaInfo `json:"meta,omitempty"`
	// CredentialRefs lists the credential names the connection uses
	CredentialRefs []string `json:"credentialRefs,omitempty"`
	// Auth is set when an auth flow obtains the credential
	Auth *IntegrationProviderAuth `json:"auth,omitempty"`
	// Disconnect describes how the connection tears down an installation
	Disconnect *types.DisconnectRegistration `json:"disconnect,omitempty"`
}

// IntegrationProviderAuth marks a connection whose credential an auth flow obtains
type IntegrationProviderAuth struct{}

// IntegrationAuthStartRequest is the request type for starting an integration auth flow.
type IntegrationAuthStartRequest = openapi.IntegrationAuthStartRequest

// ExampleIntegrationAuthStartRequest is an example auth start request for OpenAPI documentation.
var ExampleIntegrationAuthStartRequest = openapi.ExampleIntegrationAuthStartRequest

// ExampleConfigureIntegrationRequest is an example configuration payload for OpenAPI documentation.
var ExampleConfigureIntegrationRequest = openapi.ExampleConfigureIntegrationRequest

// ExampleRunIntegrationOperationRequest is an example operation payload for OpenAPI documentation.
var ExampleRunIntegrationOperationRequest = openapi.ExampleRunIntegrationOperationRequest

// ExampleIntegrationHealthRequest is an example health assessment request for OpenAPI documentation.
var ExampleIntegrationHealthRequest = openapi.ExampleIntegrationHealthRequest
