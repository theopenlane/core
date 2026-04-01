package openapi

import "encoding/json"

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
