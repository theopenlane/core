//nolint:revive
package types

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/state"
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
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" {
		return ProviderUnknown
	}

	return ProviderType(normalized)
}

// AuthKind indicates how a provider authenticates (oauth2, oidc, workload identity, etc)
type AuthKind string

const (
	// AuthKindUnknown represents an unset authentication kind.
	AuthKindUnknown AuthKind = ""
	// AuthKindOAuth2 represents OAuth2 authentication
	AuthKindOAuth2 AuthKind = "oauth2"
	// AuthKindOAuth2ClientCredentials represents OAuth2 client-credentials authentication.
	// #nosec G101 -- this is an auth kind enum identifier, not a secret.
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
)

// AuthKindFromString normalizes auth kind values from config/user input.
func AuthKindFromString(value string) AuthKind {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" {
		return AuthKindUnknown
	}

	return AuthKind(normalized)
}

// Normalize returns a normalized auth kind value.
func (k AuthKind) Normalize() AuthKind {
	return AuthKindFromString(string(k))
}

// IsKnown reports whether the auth kind is one of the supported canonical values.
func (k AuthKind) IsKnown() bool {
	switch k.Normalize() {
	case AuthKindOAuth2, AuthKindOAuth2ClientCredentials, AuthKindOIDC, AuthKindAPIKey, AuthKindGitHubApp, AuthKindWorkloadIdentity, AuthKindAWSFederation:
		return true
	default:
		return false
	}
}

// SupportsInteractiveFlow reports whether the auth kind supports browser OAuth/OIDC redirects.
func (k AuthKind) SupportsInteractiveFlow() bool {
	switch k.Normalize() {
	case AuthKindOAuth2, AuthKindOIDC:
		return true
	default:
		return false
	}
}

// ProviderCapabilities describe optional behaviours supported by a provider
type ProviderCapabilities struct {
	// SupportsRefreshTokens indicates the provider can refresh tokens via Mint
	SupportsRefreshTokens bool
	// SupportsClientPooling indicates the provider can participate in client pools
	SupportsClientPooling bool
	// SupportsMetadataForm indicates there is a credentials schema for declarative handlers
	SupportsMetadataForm bool
	// EnvironmentCredentials indicates the provider derives credentials from installation
	// context (e.g. app keys, workload identity) rather than persisted per-user OAuth tokens.
	// Providers with this capability always mint rather than loading from the credential store,
	// and minted tokens are never written back to the store.
	EnvironmentCredentials bool
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
	// Schema exposes the JSON schema for declarative credential forms.
	Schema json.RawMessage
	// Metadata contains additional provider metadata.
	Metadata json.RawMessage
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
	Mint(ctx context.Context, request CredentialMintRequest) (models.CredentialSet, error)
}

// ClientName identifies a specific client type exposed by a provider (e.g., rest, graphql)
type ClientName string

// ClientBuilderFunc constructs provider-specific clients using persisted credentials and optional config
type ClientBuilderFunc func(ctx context.Context, credential models.CredentialSet, config json.RawMessage) (ClientInstance, error)

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
	// ConfigSchema defines the JSON schema for client configuration.
	ConfigSchema json.RawMessage
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
	Config json.RawMessage
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
	Finish(ctx context.Context, code string) (models.CredentialSet, error)
}

// CredentialMintRequest is passed to Provider.Mint when the broker needs to refresh
// or exchange long-lived credentials for short-lived ones (e.g., STS, PKCE).
type CredentialMintRequest struct {
	// Provider identifies the provider whose credentials are being refreshed
	Provider ProviderType
	// OrgID identifies the organization requesting minting
	OrgID string
	// IntegrationID references the integration record containing the credential
	IntegrationID string
	// Credential contains the previously stored credential fields
	Credential models.CredentialSet
	// ProviderState carries optional provider state from the integration record
	ProviderState *state.IntegrationProviderState
	// Attributes carries additional provider-specific attributes
	Attributes map[string]string
	// Scopes optionally override scopes for the mint call
	Scopes []string
}
