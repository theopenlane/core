package openapi

import (
	"encoding/json"
	"errors"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/pkg/jsonx"
)

// OAuthPublicConfig holds the public OAuth configuration for a definition.
type OAuthPublicConfig struct {
	ClientID    string            `json:"clientId,omitempty"`
	AuthURL     string            `json:"authUrl,omitempty"`
	TokenURL    string            `json:"tokenUrl,omitempty"`
	RedirectURI string            `json:"redirectUri,omitempty"`
	Scopes      []string          `json:"scopes,omitempty"`
	UsePKCE     bool              `json:"usePkce,omitempty"`
	AuthParams  map[string]string `json:"authParams,omitempty"`
	TokenParams map[string]string `json:"tokenParams,omitempty"`
}

// IntegrationProviderMetadata is a snapshot of definition metadata captured on installation.
type IntegrationProviderMetadata struct {
	Name                   string             `json:"name"`
	DisplayName            string             `json:"displayName"`
	Category               string             `json:"category"`
	Description            string             `json:"description,omitempty"`
	AuthType               string             `json:"authType"`
	AuthStartPath          string             `json:"authStartPath,omitempty"`
	AuthCallbackPath       string             `json:"authCallbackPath,omitempty"`
	Active                 bool               `json:"active"`
	Visible                bool               `json:"visible"`
	Tags                   []string           `json:"tags,omitempty"`
	LogoURL                string             `json:"logoUrl,omitempty"`
	DocsURL                string             `json:"docsUrl,omitempty"`
	OAuth                  *OAuthPublicConfig `json:"oauth,omitempty"`
	EnvironmentCredentials json.RawMessage    `json:"environmentCredentials,omitempty"`
	CredentialsSchema      json.RawMessage    `json:"credentialsSchema,omitempty"`
	Persistence            map[string]any     `json:"persistence,omitempty"`
	Labels                 map[string]string  `json:"labels,omitempty"`
}

// IntegrationConfig is the per-installation runtime configuration stored as a typed JSON field.
type IntegrationConfig struct {
	ClientConfig      json.RawMessage         `json:"clientConfig,omitempty"`
	SCIMProvisionMode enums.SCIMProvisionMode `json:"scimProvisionMode,omitempty"`
}

// IntegrationProviderState stores provider-specific integration state captured during auth and config.
type IntegrationProviderState struct {
	Providers map[string]json.RawMessage `json:"providers,omitempty"`
}

// ErrProviderStateDecode is returned when provider state cannot be decoded.
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
