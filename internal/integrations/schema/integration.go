package schema

import (
	"encoding/json"
	"time"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/pkg/jsonx"
)

// AuthKind identifies the authentication mechanism used by a definition
type AuthKind string

const (
	// AuthKindUnknown represents an unset authentication kind
	AuthKindUnknown AuthKind = ""
	// AuthKindOAuth2 represents OAuth2 authentication
	AuthKindOAuth2 AuthKind = "oauth2"
	// AuthKindOAuth2ClientCredentials represents OAuth2 client-credentials authentication
	AuthKindOAuth2ClientCredentials AuthKind = "oauth2_client_credentials"
	// AuthKindOIDC represents OpenID Connect authentication
	AuthKindOIDC AuthKind = "oidc"
	// AuthKindAPIKey represents API key authentication
	AuthKindAPIKey AuthKind = "apikey"
	// AuthKindGitHubApp represents GitHub App authentication
	AuthKindGitHubApp AuthKind = "githubapp"
	// AuthKindWorkloadIdentity represents workload identity authentication
	AuthKindWorkloadIdentity AuthKind = "workload_identity"
	// AuthKindAWSFederation represents AWS STS federation authentication
	AuthKindAWSFederation AuthKind = "aws_sts"
	// AuthKindNone represents push-based providers where the external system calls us
	AuthKindNone AuthKind = "none"
)

// OAuthPublicConfig holds the public OAuth configuration for a definition
type OAuthPublicConfig struct {
	// ClientID is the OAuth application client identifier
	ClientID string `json:"clientId,omitempty"`
	// AuthURL is the authorization endpoint URL
	AuthURL string `json:"authUrl,omitempty"`
	// TokenURL is the token endpoint URL
	TokenURL string `json:"tokenUrl,omitempty"`
	// RedirectURI is the OAuth redirect URI
	RedirectURI string `json:"redirectUri,omitempty"`
	// Scopes lists the OAuth scopes
	Scopes []string `json:"scopes,omitempty"`
	// UsePKCE indicates whether PKCE is used
	UsePKCE bool `json:"usePkce,omitempty"`
	// AuthParams carries additional authorization endpoint parameters
	AuthParams map[string]string `json:"authParams,omitempty"`
	// TokenParams carries additional token endpoint parameters
	TokenParams map[string]string `json:"tokenParams,omitempty"`
}

// PersistenceConfig controls credential storage behaviour for a definition
type PersistenceConfig struct {
	// StoreRefreshToken indicates whether to persist the refresh token
	StoreRefreshToken bool `json:"storeRefreshToken"`
}

// IntegrationProviderMetadata is a snapshot of definition metadata captured on installation
type IntegrationProviderMetadata struct {
	// Name is the definition's canonical identifier
	Name string `json:"name"`
	// DisplayName is the UI-friendly name
	DisplayName string `json:"displayName"`
	// Category groups definitions by domain
	Category string `json:"category"`
	// Description is the UI-friendly description
	Description string `json:"description,omitempty"`
	// AuthType identifies the authentication mechanism used by the definition
	AuthType AuthKind `json:"authType"`
	// AuthStartPath is the API path to initiate provider authentication
	AuthStartPath string `json:"authStartPath,omitempty"`
	// AuthCallbackPath is the API callback path used to complete provider authentication
	AuthCallbackPath string `json:"authCallbackPath,omitempty"`
	// Active indicates whether the definition is active
	Active bool `json:"active"`
	// Visible indicates whether the definition is visible in the UI
	Visible bool `json:"visible"`
	// Tags is an optional list of categorization tags
	Tags []string `json:"tags,omitempty"`
	// LogoURL references a logo asset
	LogoURL string `json:"logoUrl,omitempty"`
	// DocsURL links to definition documentation
	DocsURL string `json:"docsUrl,omitempty"`
	// OAuth holds the public OAuth configuration for this definition
	OAuth *OAuthPublicConfig `json:"oauth,omitempty"`
	// EnvironmentCredentials carries operator-configured credential attributes
	EnvironmentCredentials json.RawMessage `json:"environmentCredentials,omitempty"`
	// CredentialsSchema is the JSON schema for the credentials form
	CredentialsSchema json.RawMessage `json:"credentialsSchema,omitempty"`
	// Persistence controls credential storage behaviour
	Persistence *PersistenceConfig `json:"persistence,omitempty"`
	// Labels carries arbitrary key-value metadata
	Labels map[string]string `json:"labels,omitempty"`
}

// OperationTemplate holds a saved operation configuration template
type OperationTemplate struct {
	// Config holds the operation configuration
	Config json.RawMessage `json:"config,omitempty"`
	// AllowOverrides lists which config fields can be overridden at runtime
	AllowOverrides []string `json:"allowOverrides,omitempty"`
}

// MappingOverride holds a user-configurable CEL expression override for one mapping schema
type MappingOverride struct {
	// Version is the schema version for this override
	Version string `json:"version,omitempty"`
	// FilterExpr is a CEL expression evaluated against the raw provider payload
	FilterExpr string `json:"filterExpr,omitempty"`
	// MapExpr is a CEL expression that projects the raw payload into the normalized output schema
	MapExpr string `json:"mapExpr,omitempty"`
}

// RetentionPolicy defines storage settings for integration payloads
type RetentionPolicy struct {
	// StoreRawPayload indicates whether to store the raw payload
	StoreRawPayload bool `json:"storeRawPayload,omitempty"`
	// PayloadTTL defines how long raw payloads are retained
	PayloadTTL time.Duration `json:"payloadTtl,omitempty"`
}

// IntegrationConfig is the per-installation runtime configuration stored as a typed JSON field
type IntegrationConfig struct {
	// OperationTemplates holds saved operation templates keyed by operation name
	OperationTemplates map[string]OperationTemplate `json:"operationTemplates,omitempty"`
	// EnabledOperations lists which operations are enabled
	EnabledOperations []string `json:"enabledOperations,omitempty"`
	// ClientConfig holds provider-specific client configuration
	ClientConfig json.RawMessage `json:"clientConfig,omitempty"`
	// CollectionStrategy defines how data is collected from the provider
	CollectionStrategy string `json:"collectionStrategy,omitempty"`
	// Schedule defines the integration schedule
	Schedule string `json:"schedule,omitempty"`
	// PollInterval defines how often to poll the provider for new data
	PollInterval time.Duration `json:"pollInterval,omitempty"`
	// MappingOverrides holds user-configurable CEL expression overrides keyed by schema name
	MappingOverrides map[string]MappingOverride `json:"mappingOverrides,omitempty"`
	// RetentionPolicy defines the data retention policy
	RetentionPolicy *RetentionPolicy `json:"retentionPolicy,omitempty"`
	// SCIMProvisionMode controls how SCIM push events are persisted
	SCIMProvisionMode enums.SCIMProvisionMode `json:"scimProvisionMode,omitempty"`
}

// IntegrationProviderState stores provider-specific integration state captured during auth and config
type IntegrationProviderState struct {
	// Providers contains provider-specific state by provider key
	Providers map[string]json.RawMessage `json:"providers,omitempty"`
}

// ProviderData returns the raw provider state for a provider key
func (s IntegrationProviderState) ProviderData(provider string) json.RawMessage {
	if provider == "" || len(s.Providers) == 0 {
		return nil
	}

	return s.Providers[provider]
}

// MergeProviderData deep-merges provider state and reports whether state changed
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
