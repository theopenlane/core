//nolint:revive
package types

import (
	"context"
)

// ProviderType is a strongly typed identifier for an integration provider
// (github, slack, okta, etc). Use ProviderTypeFromString to normalize user input
// rather than relying on direct string comparisons in call sites.
type ProviderType string

const (
	// ProviderUnknown represents an unset provider identifier
	ProviderUnknown ProviderType = ""
)

// ProviderTypeFromString normalizes arbitrary user/config input into a stable ProviderType
func ProviderTypeFromString(value string) ProviderType {
	return ProviderType(value)
}

// AuthKind indicates how a provider authenticates (oauth2, oidc, workload identity, etc)
type AuthKind string

const (
	// AuthKindOAuth2 represents OAuth2 authentication
	AuthKindOAuth2 AuthKind = "oauth2"
	// AuthKindOIDC represents OpenID Connect authentication
	AuthKindOIDC AuthKind = "oidc"
	// AuthKindAPIKey represents API key authentication
	AuthKindAPIKey AuthKind = "apikey"
	// AuthKindGitHubApp represents GitHub App authentication
	AuthKindGitHubApp AuthKind = "github_app"
	// AuthKindWorkloadIdentity represents workload identity authentication
	AuthKindWorkloadIdentity AuthKind = "workload_identity"
	// AuthKindAWSFederation represents AWS STS federation authentication
	AuthKindAWSFederation AuthKind = "aws_sts"
)

// ProviderCapabilities describe optional behaviours supported by a provider
type ProviderCapabilities struct {
	// SupportsRefreshTokens indicates the provider can refresh tokens via Mint
	SupportsRefreshTokens bool
	// SupportsClientPooling indicates the provider can participate in client pools
	SupportsClientPooling bool
	// SupportsMetadataForm indicates there is a credentials schema for declarative handlers
	SupportsMetadataForm bool
}

// ProviderConfig mirrors the declarative provider specification (JSON/YAML) used by HTTP handlers to render forms
type ProviderConfig struct {
	// Type is the unique provider identifier
	Type ProviderType
	// Auth indicates the authentication kind (oauth2, oidc, etc)
	Auth AuthKind
	// DisplayName is the UI-friendly name
	DisplayName string
	// Description is the UI-friendly description
	Description string
	// Category groups providers (code, collab, etc)
	Category string
	// DocsURL links to provider documentation
	DocsURL string
	// LogoURL references a logo asset
	LogoURL string
	// Schema exposes the JSON schema for declarative credential forms
	Schema map[string]any
	// Metadata contains additional provider metadata
	Metadata map[string]any
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
	Mint(ctx context.Context, subject CredentialSubject) (CredentialPayload, error)
}

// ClientName identifies a specific client type exposed by a provider (e.g., rest, graphql)
type ClientName string

// ClientBuilderFunc constructs provider-specific clients using persisted credentials and optional config
type ClientBuilderFunc func(ctx context.Context, payload CredentialPayload, config map[string]any) (any, error)

// ClientDescriptor describes a provider-managed client that can be pooled/reused downstream
type ClientDescriptor struct {
	// Provider identifies which provider offers this client
	Provider ProviderType
	// Name is the unique client identifier
	Name ClientName
	// Description explains what the client does
	Description string
	// Build is the function that constructs the client
	Build ClientBuilderFunc
	// ConfigSchema defines the JSON schema for client configuration
	ConfigSchema map[string]any
}

// ClientProvider is implemented by providers that expose SDK clients for downstream services
type ClientProvider interface {
	Provider
	// ClientDescriptors returns the list of clients offered by the provider
	ClientDescriptors() []ClientDescriptor
}

// ClientRequest contains the parameters required to request a client instance
type ClientRequest struct {
	// OrgID identifies the organization requesting the client
	OrgID string
	// Provider identifies which provider to use
	Provider ProviderType
	// Client identifies which client type to build
	Client ClientName
	// Config contains client-specific configuration
	Config map[string]any
	// Force bypasses cached client instances
	Force bool
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
	Metadata map[string]any
	// LabelOverrides customizes labels shown to users
	LabelOverrides map[string]string
}

// AuthSession encapsulates an authorization transaction (state, nonce, URL)
// Keymaker implementations store arbitrary data inside the session, then call Finish once the provider redirected with an authorization code
type AuthSession interface {
	// ProviderType returns the provider identifier for this session
	ProviderType() ProviderType
	// State returns the CSRF state value
	State() string
	// AuthURL returns the URL where the user should be redirected
	AuthURL() string
	// Finish exchanges the authorization code for credentials
	Finish(ctx context.Context, code string) (CredentialPayload, error)
}

// CredentialSubject is passed to Provider.Mint when the broker needs to refresh or exchange long-lived credentials for short-lived ones (e.g., STS, PKCE)
type CredentialSubject struct {
	// Provider identifies the provider whose credentials are being refreshed
	Provider ProviderType
	// OrgID identifies the organization requesting minting
	OrgID string
	// IntegrationID references the integration record containing the credential
	IntegrationID string
	// Credential contains the previously stored credential payload
	Credential CredentialPayload
	// Attributes carries additional provider-specific attributes
	Attributes map[string]string
	// Scopes optionally override scopes for the mint call
	Scopes []string
}
