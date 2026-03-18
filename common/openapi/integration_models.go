package openapi

import (
	"encoding/json"
	"errors"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/pkg/jsonx"
)

// OAuthPublicConfig holds the public OAuth configuration for a definition
type OAuthPublicConfig struct {
	// ClientID is the OAuth client ID
	ClientID string `json:"clientId,omitempty"`
	// AuthURL is the OAuth authorization endpoint URL
	AuthURL string `json:"authUrl,omitempty"`
	// TokenURL is the OAuth token endpoint URL
	TokenURL string `json:"tokenUrl,omitempty"`
	// RedirectURI is the OAuth redirect URI
	RedirectURI string `json:"redirectUri,omitempty"`
	// Scopes is the list of OAuth scopes
	Scopes []string `json:"scopes,omitempty"`
	// UsePKCE indicates if PKCE is used
	UsePKCE bool `json:"usePkce,omitempty"`
	// AuthParams are additional parameters for the auth request
	AuthParams map[string]string `json:"authParams,omitempty"`
	// TokenParams are additional parameters for the token request
	TokenParams map[string]string `json:"tokenParams,omitempty"`
}

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
	// OAuth holds the public OAuth configuration
	OAuth *OAuthPublicConfig `json:"oauth,omitempty"`
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
	// SCIMProvisionMode is the SCIM provision mode
	SCIMProvisionMode enums.SCIMProvisionMode `json:"scimProvisionMode,omitempty"`
}

// IntegrationInstallationMetadata stores stable, non-secret installation identity metadata
type IntegrationInstallationMetadata struct {
	// Attributes is the provider-defined installation metadata payload
	Attributes json.RawMessage `json:"attributes,omitempty"`
}

// IntegrationProviderState stores provider-specific integration state captured during auth and config
type IntegrationProviderState struct {
	// Providers is a map of provider keys to raw state JSON
	Providers map[string]json.RawMessage `json:"providers,omitempty"`
}

// ErrProviderStateDecode is returned when provider state cannot be decoded
var ErrProviderStateDecode = errors.New("failed to decode provider state")

// ProviderData returns the raw provider state for a provider key.
func (s IntegrationProviderState) ProviderData(provider string) json.RawMessage {
	if provider == "" || len(s.Providers) == 0 {
		return nil
	}

	return s.Providers[provider]
}

// MergeProviderData deep-merges provider state and reports whether state changed.
func (s *IntegrationProviderState) MergeProviderData(provider string, patch json.RawMessage) (bool, error) {
	if s == nil || provider == "" || len(patch) == 0 {
		return false, nil
	}

	if s.Providers == nil {
		s.Providers = map[string]json.RawMessage{}
	}

	merged, changed, err := jsonx.DeepMerge(s.Providers[provider], patch)
	if err != nil {
		return false, ErrProviderStateDecode
	}

	if !changed {
		return false, nil
	}

	s.Providers[provider] = merged

	return true, nil
}
