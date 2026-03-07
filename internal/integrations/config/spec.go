package config

import (
	"context"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// ProviderSpec mirrors the declarative provider definition files rendered in the UI
type ProviderSpec struct {
	// Name is the provider identifier
	Name string `json:"name"`
	// DisplayName is the UI-facing name
	DisplayName string `json:"displayName"`
	// Category groups providers (code, collab, etc)
	Category string `json:"category"`
	// Description optionally describes the provider in the UI
	Description string `json:"description,omitempty"`
	// AuthType describes the authentication kind
	AuthType types.AuthKind `json:"authType"`
	// AuthStartPath is the integration API path to initiate provider authentication
	AuthStartPath string `json:"authStartPath,omitempty"`
	// AuthCallbackPath is the integration API callback path used to complete provider authentication
	AuthCallbackPath string `json:"authCallbackPath,omitempty"`
	// Active toggles provider availability
	Active *bool `json:"active,omitempty"`
	// Visible toggles provider visibility in the UI
	Visible *bool `json:"visible,omitempty"`
	// Tags define UI labels/pills rendered for the provider card
	Tags []string `json:"tags,omitempty"`
	// LogoURL references the logo asset
	LogoURL string `json:"logoUrl"`
	// DocsURL links to provider documentation
	DocsURL string `json:"docsUrl"`
	// SchemaVersion identifies the spec schema version
	SchemaVersion string `json:"schemaVersion,omitempty"`
	// OAuth contains OAuth configuration when applicable
	OAuth *OAuthSpec `json:"oauth,omitempty" koanf:"oauth"`
	// APIKey contains API key configuration when applicable
	APIKey *APIKeySpec `json:"apiKey,omitempty"`
	// UserInfo describes optional user info lookups
	UserInfo *UserInfoSpec `json:"userInfo,omitempty"`
	// WorkloadIdentity contains Google WIF defaults
	GoogleWorkloadIdentity *GoogleWorkloadIdentitySpec `json:"googleWorkloadIdentity,omitempty" koanf:"workloadidentity"`
	// GitHubApp configures GitHub App providers
	GitHubApp *GitHubAppSpec `json:"githubApp,omitempty" koanf:"app"`
	// AWSSTS configures AWS federation defaults
	AWSSTS *AWSFederationSpec `json:"awsSts,omitempty" koanf:"sts"`
	// CredentialsSchema drives declarative credential forms
	CredentialsSchema map[string]any `json:"credentialsSchema,omitempty"`
	// Persistence configures storage policies
	Persistence *PersistenceSpec `json:"persistence,omitempty"`
	// Labels carries optional metadata labels
	Labels map[string]string `json:"labels,omitempty"`
	// Metadata stores additional provider metadata
	Metadata map[string]any `json:"metadata,omitempty"`
	// Defaults stores provider-specific defaults
	Defaults map[string]any `json:"defaults,omitempty"`
	// SuccessRedirectURL is the URL to redirect to after successful authentication for this provider.
	// When empty, handlers return JSON instead of redirecting.
	SuccessRedirectURL string `json:"successRedirectUrl,omitempty" koanf:"successredirecturl"`
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
		Description: s.Description,
		Category:    s.Category,
		DocsURL:     s.DocsURL,
		LogoURL:     s.LogoURL,
		Schema:      s.CredentialsSchema,
		Metadata:    s.Metadata,
	}
}

// MergeProviderSpecs overlays provider-specific overrides onto base specs using provider keys
func MergeProviderSpecs(ctx context.Context, base map[types.ProviderType]ProviderSpec, overrides map[string]ProviderSpec) map[types.ProviderType]ProviderSpec {
	merged := lo.Assign(map[types.ProviderType]ProviderSpec{}, base)

	for key, override := range overrides {
		provider := types.ProviderTypeFromString(key)
		current, ok := merged[provider]
		if !ok {
			continue
		}

		currentMap, err := jsonx.ToMap(current)
		if err != nil {
			logx.FromContext(ctx).Warn().Err(err).Str("provider", key).Msg("failed to serialize base provider spec for merge")
			continue
		}

		nextMap, err := JSONValue(currentMap, override, MapOptions{
			PruneZero: true,
			DeepMerge: true,
		})
		if err != nil {
			logx.FromContext(ctx).Warn().Err(err).Str("provider", key).Msg("failed to merge provider spec override")
			continue
		}

		var next ProviderSpec
		if err := jsonx.RoundTrip(nextMap, &next); err != nil {
			logx.FromContext(ctx).Warn().Err(err).Str("provider", key).Msg("failed to apply provider spec override")
			continue
		}

		merged[provider] = next
	}

	return merged
}

// PersistenceSpec controls how secrets are stored
type PersistenceSpec struct {
	// StoreRefreshToken indicates refresh tokens should be persisted
	StoreRefreshToken bool `json:"storeRefreshToken"`
}

// OAuthSpec captures OAuth2/OIDC metadata from the JSON files
type OAuthSpec struct {
	// ClientID is the OAuth client identifier
	ClientID string `json:"clientId" koanf:"clientid"`
	// ClientSecret is the OAuth client secret
	ClientSecret string `json:"clientSecret" koanf:"clientsecret" sensitive:"true"`
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
	// Audience is the Openlane WIF pool/provider audience for STS exchanges; operator-configured.
	Audience string `json:"audience,omitempty" koanf:"audience"`
	// TargetServiceAccount is the per-tenant service account provided by the user via credentialsSchema; not operator-configured.
	TargetServiceAccount string `json:"targetServiceAccount,omitempty"`
	// Scopes enumerates default GCP API scopes; operator-configured.
	Scopes []string `json:"scopes,omitempty"`
	// TokenLifetime configures the default token lifetime; operator-configured.
	TokenLifetime time.Duration `json:"tokenLifetime,omitempty" koanf:"tokenlifetime"`
	// SubjectTokenType configures the subject token type for STS; operator-configured.
	SubjectTokenType string `json:"subjectTokenType,omitempty" koanf:"subjecttokentype"`
}

// GitHubAppSpec holds GitHub App metadata
type GitHubAppSpec struct {
	// BaseURL optionally sets a custom API base (GitHub Enterprise, etc)
	BaseURL string `json:"baseUrl,omitempty"`
	// TokenTTL optionally indicates desired installation token lifetime
	TokenTTL time.Duration `json:"tokenTtl,omitempty"`
	// AppSlug optionally exposes the configured app slug for UI metadata
	AppSlug string `json:"appSlug,omitempty" koanf:"appslug"`
	// AppID carries the runtime GitHub App ID used for signing JWTs; this field is never serialized to JSON.
	AppID string `json:"-" koanf:"appid" sensitive:"true"`
	// PrivateKey carries the runtime GitHub App private key used for signing JWTs; this field is never serialized to JSON.
	PrivateKey string `json:"-" koanf:"privatekey" sensitive:"true"`
	// WebhookSecret is the shared secret used to validate incoming GitHub webhooks; never serialized to JSON.
	WebhookSecret string `json:"-" koanf:"webhooksecret" sensitive:"true"`
}

// AWSFederationSpec captures AssumeRoleWithWebIdentity defaults
type AWSFederationSpec struct {
	// RoleARN is the default role to assume
	RoleARN string `json:"roleArn,omitempty" koanf:"rolearn"`
	// SessionName is the default AWS session name
	SessionName string `json:"sessionName,omitempty" koanf:"sessionname"`
	// Duration is the default session duration
	Duration time.Duration `json:"duration,omitempty" koanf:"duration"`
	// Region is the default AWS region
	Region string `json:"region,omitempty" koanf:"region"`
	// ExternalID optionally configures the STS external ID
	ExternalID string `json:"externalId,omitempty" koanf:"externalid"`
}
