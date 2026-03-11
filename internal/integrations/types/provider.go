package types

import (
	"context"
	"encoding/json"
	"strings"
)

// ProviderType is a strongly typed identifier for an integration provider
type ProviderType string

const (
	// ProviderUnknown represents an unset provider identifier
	ProviderUnknown ProviderType = ""
	// ProviderSCIM represents a generic SCIM 2.0 push source; covers any external IdP or
	// directory that speaks SCIM 2.0 regardless of vendor
	ProviderSCIM ProviderType = "scim"
)

// AuthKind indicates how a provider authenticates
type AuthKind string

const (
	// AuthKindUnknown represents an unset authentication kind
	AuthKindUnknown AuthKind = ""
	// AuthKindOAuth2 represents OAuth2 authentication
	AuthKindOAuth2 AuthKind = "oauth2"
	// AuthKindOAuth2ClientCredentials represents OAuth2 client-credentials authentication
	// #nosec G101 -- this is an auth kind enum identifier, not a secret
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
	// AuthKindNone represents push-based providers (e.g. SCIM) where the external system calls us;
	// no outbound credentials are stored
	AuthKindNone AuthKind = "none"
)

// ProviderTypeFromString normalizes arbitrary user/config input into a stable ProviderType
func ProviderTypeFromString(value string) ProviderType {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" {
		return ProviderUnknown
	}

	return ProviderType(normalized)
}

// IsKnown reports whether the auth kind is one of the supported canonical values
func (k AuthKind) IsKnown() bool {
	switch k.Normalize() {
	case AuthKindOAuth2, AuthKindOAuth2ClientCredentials, AuthKindOIDC, AuthKindAPIKey, AuthKindGitHubApp, AuthKindWorkloadIdentity, AuthKindAWSFederation, AuthKindNone:
		return true
	default:
		return false
	}
}

// Normalize returns the auth kind normalized to its canonical form
func (k AuthKind) Normalize() AuthKind {
	trimmed := strings.TrimSpace(strings.ToLower(string(k)))
	if trimmed == "" {
		return AuthKindUnknown
	}

	return AuthKind(trimmed)
}

// SupportsInteractiveFlow reports whether the auth kind supports browser OAuth/OIDC redirects
func (k AuthKind) SupportsInteractiveFlow() bool {
	switch k.Normalize() {
	case AuthKindOAuth2, AuthKindOIDC:
		return true
	default:
		return false
	}
}

// ProviderCapabilities describes optional behaviours supported by a provider
type ProviderCapabilities struct {
	// SupportsRefreshTokens indicates the provider can refresh tokens via Mint
	SupportsRefreshTokens bool
	// SupportsClientPooling indicates the provider can participate in client pools
	SupportsClientPooling bool
	// SupportsMetadataForm indicates there is a credentials schema for declarative handlers
	SupportsMetadataForm bool
	// EnvironmentCredentials indicates the provider derives credentials from installation
	// context rather than persisted per-user OAuth tokens
	EnvironmentCredentials bool
}

// Provider defines the behaviour required to integrate a third-party system
type Provider interface {
	// Type returns the provider identifier
	Type() ProviderType
	// Capabilities returns the capabilities supported by the provider
	Capabilities() ProviderCapabilities
	// BeginAuth starts an authentication flow
	BeginAuth(ctx context.Context, input AuthContext) (AuthSession, error)
	// Mint refreshes or exchanges credentials
	Mint(ctx context.Context, request CredentialMintRequest) (CredentialSet, error)
}

// ClientProvider is implemented by providers that expose SDK clients for downstream services
type ClientProvider interface {
	Provider
	// ClientDescriptors returns the list of clients offered by the provider
	ClientDescriptors() []ClientDescriptor
}

// OperationProvider is implemented by providers that publish runtime operations
type OperationProvider interface {
	Provider
	// Operations returns the list of operations offered by the provider
	Operations() []OperationDescriptor
}

// MappingProvider is implemented by providers that supply built-in ingest mappings
type MappingProvider interface {
	Provider
	// DefaultMappings returns built-in mappings across one or more schemas
	DefaultMappings() []MappingRegistration
}

// AuthContext carries the state necessary to start an OAuth/OIDC transaction
type AuthContext struct {
	// OrgID identifies the organization initiating the flow
	OrgID string
	// IntegrationID identifies the integration record (optional for new flows)
	IntegrationID string
	// RedirectURI overrides the default redirect URI when necessary
	RedirectURI string
	// State carries the CSRF token/opaque state
	State string
	// Scopes optionally override provider defaults
	Scopes []string
	// Metadata carries additional activation metadata
	Metadata json.RawMessage
	// LabelOverrides customizes labels shown to users
	LabelOverrides map[string]string
}

// AuthSession encapsulates an authorization transaction (state, nonce, URL)
type AuthSession interface {
	// ProviderType returns the provider identifier for this session
	ProviderType() ProviderType
	// State returns the CSRF state value
	State() string
	// AuthURL returns the URL where the user should be redirected
	AuthURL() string
	// Finish exchanges the authorization code for credentials
	Finish(ctx context.Context, code string) (CredentialSet, error)
}

// CredentialMintRequest is passed to Provider.Mint when the broker needs to refresh
// or exchange long-lived credentials for short-lived ones
type CredentialMintRequest struct {
	// Provider identifies the provider whose credentials are being refreshed
	Provider ProviderType
	// OrgID identifies the organization requesting minting
	OrgID string
	// IntegrationID references the integration record containing the credential
	IntegrationID string
	// Credential contains the previously stored credential fields
	Credential CredentialSet
	// ProviderState carries optional provider state from the integration record.
	// This will become *state.IntegrationProviderState once the state subpackage is created
	ProviderState json.RawMessage
	// Attributes carries additional provider-specific attributes
	Attributes map[string]string
	// Scopes optionally override scopes for the mint call
	Scopes []string
}

// IntegrationProviderMetadata describes the data required for rendering integration forms
type IntegrationProviderMetadata struct {
	// Name is the provider's canonical name
	Name string `json:"name"`
	// DisplayName is the UI-friendly name
	DisplayName string `json:"displayName"`
	// Category groups providers (code, collab, etc.)
	Category string `json:"category"`
	// Description is the UI-friendly description
	Description string `json:"description,omitempty"`
	// AuthType indicates the authentication kind
	AuthType AuthKind `json:"authType"`
	// AuthStartPath is the integration API path to initiate provider authentication
	AuthStartPath string `json:"authStartPath,omitempty"`
	// AuthCallbackPath is the integration API callback path used to complete provider authentication
	AuthCallbackPath string `json:"authCallbackPath,omitempty"`
	// Active indicates whether the provider is active
	Active bool `json:"active"`
	// Visible indicates whether the provider is visible in the UI
	Visible bool `json:"visible"`
	// Tags is an optional list of categorization tags
	Tags []string `json:"tags,omitempty"`
	// LogoURL references a logo asset
	LogoURL string `json:"logoUrl,omitempty"`
	// DocsURL links to provider documentation
	DocsURL string `json:"docsUrl,omitempty"`
	// OAuth holds the public OAuth configuration for this provider
	OAuth *OAuthPublicConfig `json:"oauth,omitempty"`
	// EnvironmentCredentials carries operator-configured credential attributes
	EnvironmentCredentials json.RawMessage `json:"environmentCredentials,omitempty"`
	// CredentialsSchema is the JSON schema for the credentials form
	CredentialsSchema json.RawMessage `json:"credentialsSchema,omitempty"`
	// Persistence controls credential storage behaviour
	Persistence *PersistenceConfig `json:"persistence,omitempty"`
	// Labels carries arbitrary key-value metadata
	Labels map[string]string `json:"labels,omitempty"`
	// Operations lists the operations published by this provider
	Operations []OperationMetadata `json:"operations,omitempty"`
}

// OAuthPublicConfig holds non-secret OAuth fields safe for API responses
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

// PersistenceConfig controls credential storage behaviour for a provider
type PersistenceConfig struct {
	// StoreRefreshToken indicates whether to persist the refresh token
	StoreRefreshToken bool `json:"storeRefreshToken"`
}

// OperationMetadata describes an operation published by a provider
type OperationMetadata struct {
	// Name is the operation identifier
	Name string `json:"name"`
	// Kind categorizes the operation
	Kind string `json:"kind"`
	// Description explains what the operation does
	Description string `json:"description,omitempty"`
	// Client specifies which client type is required
	Client string `json:"client,omitempty"`
	// ConfigSchema defines the JSON schema for operation configuration
	ConfigSchema json.RawMessage `json:"configSchema,omitempty"`
}

// IntegrationExecutionContext captures durable integration runtime metadata;
// it intentionally carries identifiers only (no credential secrets)
type IntegrationExecutionContext struct {
	// OrgID identifies the organization executing the integration operation
	OrgID string `json:"org_id,omitempty"`
	// IntegrationID identifies the integration record used for the run
	IntegrationID string `json:"integration_id,omitempty"`
	// Provider identifies the integration provider
	Provider ProviderType `json:"provider,omitempty"`
	// AuthKind identifies the credential auth kind when known
	AuthKind AuthKind `json:"auth_kind,omitempty"`
	// RunID identifies the integration run record when present
	RunID string `json:"run_id,omitempty"`
	// Operation identifies the operation being executed when present
	Operation OperationName `json:"operation,omitempty"`
}
