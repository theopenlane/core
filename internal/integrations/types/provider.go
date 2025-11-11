package types

import (
	"context"
	"strings"
)

// ProviderType is a strongly typed identifier for an integration provider
// (github, slack, okta, etc). Use ProviderTypeFromString to normalize user input
// rather than relying on direct string comparisons in call sites.
type ProviderType string

const (
	// ProviderUnknown represents an unset provider identifier.
	ProviderUnknown ProviderType = ""
)

// ProviderTypeFromString normalizes arbitrary user/config input into a stable
// ProviderType by trimming whitespace and lowercasing.
func ProviderTypeFromString(value string) ProviderType {
	return ProviderType(strings.TrimSpace(strings.ToLower(value)))
}

// AuthKind indicates how a provider authenticates (oauth2, oidc, workload identity, etc.).
type AuthKind string

const (
	AuthKindOAuth2           AuthKind = "oauth2"
	AuthKindOIDC             AuthKind = "oidc"
	AuthKindAPIKey           AuthKind = "apikey"
	AuthKindWorkloadIdentity AuthKind = "workload_identity"
	AuthKindGitHubApp        AuthKind = "github_app"
	AuthKindAWSFederation    AuthKind = "aws_sts"
	AuthKindAzureFederated   AuthKind = "azure_federated"
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

// ProviderConfig mirrors the declarative provider specification (JSON/YAML)
// used by the HTTP handlers to render forms
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

// Provider defines the behaviour required to integrate a third-party system.
// Implementations typically wrap Zitadel's relying-party helpers and provider-
// specific grant workflows.
type Provider interface {
	Type() ProviderType
	Capabilities() ProviderCapabilities
	BeginAuth(ctx context.Context, input AuthContext) (AuthSession, error)
	Mint(ctx context.Context, subject CredentialSubject) (CredentialPayload, error)
}

// AuthContext carries the state necessary to start an OAuth/OIDC transaction.
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

// AuthSession encapsulates an authorization transaction (state, nonce, URL).
// Keymaker implementations store arbitrary data inside the session, then call
// Finish once the provider redirected with an authorization code.
type AuthSession interface {
	ProviderType() ProviderType
	State() string
	AuthURL() string
	Finish(ctx context.Context, code string) (CredentialPayload, error)
}

// CredentialSubject is passed to Provider.Mint when the broker needs to refresh
// or exchange long-lived credentials for short-lived ones (e.g., STS, PKCE).
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
