package config

import (
	"time"

	"github.com/theopenlane/shared/integrations/types"
)

// ProviderSpec mirrors the declarative provider definition files rendered in the UI
type ProviderSpec struct {
	// Name is the provider identifier
	Name string `json:"name"`
	// DisplayName is the UI-facing name
	DisplayName string `json:"displayName"`
	// Category groups providers (code, collab, etc)
	Category string `json:"category"`
	// AuthType describes the authentication kind
	AuthType types.AuthKind `json:"authType"`
	// Active toggles provider availability
	Active bool `json:"active"`
	// LogoURL references the logo asset
	LogoURL string `json:"logoUrl"`
	// DocsURL links to provider documentation
	DocsURL string `json:"docsUrl"`
	// SchemaVersion identifies the spec schema version
	SchemaVersion string `json:"schemaVersion,omitempty"`
	// OAuth contains OAuth configuration when applicable
	OAuth *OAuthSpec `json:"oauth,omitempty"`
	// APIKey contains API key configuration when applicable
	APIKey *APIKeySpec `json:"apiKey,omitempty"`
	// UserInfo describes optional user info lookups
	UserInfo *UserInfoSpec `json:"userInfo,omitempty"`
	// WorkloadIdentity contains Google WIF defaults
	GoogleWorkloadIdentity *GoogleWorkloadIdentitySpec `json:"googleWorkloadIdentity,omitempty"`
	// GitHubApp configures GitHub App providers
	GitHubApp *GitHubAppSpec `json:"githubApp,omitempty"`
	// AWSSTS configures AWS federation defaults
	AWSSTS *AWSFederationSpec `json:"awsSts,omitempty"`
	// CredentialsSchema drives declarative credential forms
	CredentialsSchema map[string]any `json:"credentialsSchema,omitempty"`
	// Persistence configures storage policies
	Persistence *PersistenceSpec `json:"persistence,omitempty"`
	// Labels carries optional metadata labels
	Labels map[string]string `json:"labels,omitempty"`
	// Metadata stores additional provider metadata
	Metadata map[string]any `json:"metadata,omitempty"`
	// Defaults stores provider-specific defaults
	Defaults map[string]interface{} `json:"defaults,omitempty"`
}

// ProviderType returns the normalized provider identifier
func (s ProviderSpec) ProviderType() types.ProviderType {
	return types.ProviderTypeFromString(s.Name)
}

// ToProviderConfig exposes the subset of spec fields used by registries and handlers
func (s ProviderSpec) ToProviderConfig() types.ProviderConfig {
	return types.ProviderConfig{
		Type:        s.ProviderType(),
		Auth:        s.AuthType,
		DisplayName: s.DisplayName,
		Description: "",
		Category:    s.Category,
		DocsURL:     s.DocsURL,
		LogoURL:     s.LogoURL,
		Schema:      s.CredentialsSchema,
		Metadata:    s.Metadata,
	}
}

// ToProviderConfigs converts a spec map into provider configs for handler consumption.
func ToProviderConfigs(specs map[types.ProviderType]ProviderSpec) map[types.ProviderType]types.ProviderConfig {
	if len(specs) == 0 {
		return nil
	}

	out := make(map[types.ProviderType]types.ProviderConfig, len(specs))
	for provider, spec := range specs {
		out[provider] = spec.ToProviderConfig()
	}

	return out
}

// PersistenceSpec controls how secrets are stored
type PersistenceSpec struct {
	// StoreRefreshToken indicates refresh tokens should be persisted
	StoreRefreshToken bool `json:"storeRefreshToken"`
}

// OAuthSpec captures OAuth2/OIDC metadata from the JSON files
type OAuthSpec struct {
	// ClientID is the OAuth client identifier
	ClientID string `json:"clientId"`
	// ClientSecret is the OAuth client secret
	ClientSecret string `json:"clientSecret"`
	// AuthURL is the authorization endpoint
	AuthURL string `json:"authUrl"`
	// TokenURL is the token endpoint
	TokenURL string `json:"tokenUrl"`
	// Scopes lists default scopes for the provider
	Scopes []string `json:"scopes"`
	// OIDCDiscovery optionally points to the discovery endpoint
	OIDCDiscovery string `json:"oidcDiscoveryUrl,omitempty"`
	// RedirectURI overrides the default redirect URI
	RedirectURI string `json:"redirectUri,omitempty"`
	// UsePKCE enables PKCE for this provider
	UsePKCE bool `json:"usePkce,omitempty"`
	// AuthParams contains extra authorization request parameters
	AuthParams map[string]string `json:"authParams,omitempty"`
	// TokenParams contains extra token request parameters
	TokenParams map[string]string `json:"tokenParams,omitempty"`
	// AdditionalHosts enumerates additional acceptable callback hosts
	AdditionalHosts []string `json:"additionalHosts,omitempty"`
}

// APIKeySpec represents non OAuth-based providers
type APIKeySpec struct {
	// KeyLabel is the label shown in the UI
	KeyLabel string `json:"keyLabel"`
	// HeaderName describes the HTTP header carrying the key
	HeaderName string `json:"headerName"`
	// QueryParam optionally describes the query parameter carrying the key
	QueryParam string `json:"queryParam,omitempty"`
}

// UserInfoSpec drives post-auth userinfo lookups
type UserInfoSpec struct {
	// URL is the user info endpoint
	URL string `json:"url"`
	// Method is the HTTP method used for the request
	Method string `json:"method"`
	// AuthStyle indicates how to present the token
	AuthStyle string `json:"authStyle"`
	// AuthHeader optionally overrides the header name
	AuthHeader string `json:"authHeader,omitempty"`
	// IDPath describes how to extract the user ID
	IDPath string `json:"idPath"`
	// EmailPath describes how to extract the email
	EmailPath string `json:"emailPath"`
	// LoginPath describes how to extract the username
	LoginPath string `json:"loginPath"`
	// SecondaryEmailURL optionally supplies a fallback email endpoint
	SecondaryEmailURL string `json:"secondaryEmailUrl,omitempty"`
}

// GoogleWorkloadIdentitySpec configures Google WIF defaults
type GoogleWorkloadIdentitySpec struct {
	// Audience is the default audience for STS exchanges
	Audience string `json:"audience,omitempty"`
	// TargetServiceAccount is the default service account to impersonate
	TargetServiceAccount string `json:"targetServiceAccount,omitempty"`
	// Scopes enumerates default scopes on generated tokens
	Scopes []string `json:"scopes,omitempty"`
	// TokenLifetime configures the default token lifetime
	TokenLifetime time.Duration `json:"tokenLifetime,omitempty"`
	// SubjectTokenType configures the subject token type for STS
	SubjectTokenType string `json:"subjectTokenType,omitempty"`
}

// GitHubAppSpec holds GitHub App metadata
type GitHubAppSpec struct {
	// BaseURL optionally sets a custom API base (GitHub Enterprise, etc)
	BaseURL string `json:"baseUrl,omitempty"`
	// TokenTTL optionally indicates desired installation token lifetime
	TokenTTL time.Duration `json:"tokenTtl,omitempty"`
}

// AWSFederationSpec captures AssumeRoleWithWebIdentity defaults
type AWSFederationSpec struct {
	// RoleARN is the default role to assume
	RoleARN string `json:"roleArn,omitempty"`
	// SessionName is the default AWS session name
	SessionName string `json:"sessionName,omitempty"`
	// Duration is the default session duration
	Duration time.Duration `json:"duration,omitempty"`
	// Region is the default AWS region
	Region string `json:"region,omitempty"`
	// ExternalID optionally configures the STS external ID
	ExternalID string `json:"externalId,omitempty"`
}
