package keystore

import (
	"time"
)

type AuthType string

const (
	AuthTypeOAuth2           AuthType = "oauth2"
	AuthTypeOIDC             AuthType = "oidc"
	AuthTypeAPIKey           AuthType = "apikey"            // for non-OAuth connectors you may still want in the catalog
	AuthTypeWorkloadIdentity AuthType = "workload_identity" // GCP Workload Identity Federation + impersonation
	AuthTypeGitHubApp        AuthType = "github_app"        // GitHub App JWT + installation exchange
	AuthTypeAWSSTS           AuthType = "aws_sts"           // AWS AssumeRoleWithWebIdentity / STS-based federation
	AuthTypeAzureFederated   AuthType = "azure_federated"   // Azure AD federated credential exchange
)

type AuthHeaderStyle string

const (
	AuthHeaderStyleBearer AuthHeaderStyle = "Bearer"
	AuthHeaderStyleToken  AuthHeaderStyle = "token" // e.g., GitHub classic
	AuthHeaderStyleNone   AuthHeaderStyle = "none"
)

type HTTPMethod string

const (
	GET  HTTPMethod = "GET"
	POST HTTPMethod = "POST"
)

// JSONPath-like dot notation for mapping fields from userinfo JSON.
// Keep it simple (no full JSONPath engine unless you want to add one).
type JSONFieldPath string

// ProviderSpec represents one integration provider.
// Load a map[string]ProviderSpec from Koanf (keyed by provider name).
type ProviderSpec struct {
	Name        string   `json:"name"        koanf:"name"`
	DisplayName string   `json:"displayName" koanf:"displayName"`
	Category    string   `json:"category"    koanf:"category"` // "cloud","collab","code", etc.
	AuthType    AuthType `json:"authType"    koanf:"authType"` // oauth2|oidc|apikey
	Active      bool     `json:"active"      koanf:"active" default:"true"`
	LogoURL     string   `json:"logoUrl"     koanf:"logoUrl"`
	DocsURL     string   `json:"docsUrl"     koanf:"docsUrl"`

	// For OAuth2/OIDC
	OAuth *OAuthSpec `json:"oauth,omitempty" koanf:"oauth"`

	// For APIKey (optional, useful to drive forms)
	APIKey *APIKeySpec `json:"apiKey,omitempty" koanf:"apiKey"`

	// Optional: A generic userinfo fetcher you can use as a default.
	// You can override with a custom validator in code (see validators).
	UserInfo *UserInfoSpec `json:"userInfo,omitempty" koanf:"userInfo"`

	// For Workload Identity Federation / cloud impersonation flows
	WorkloadIdentity *WorkloadIdentitySpec `json:"workloadIdentity,omitempty" koanf:"workloadIdentity"`

	// For GitHub App installations
	GitHubApp *GitHubAppSpec `json:"githubApp,omitempty" koanf:"githubApp"`

	// For AWS STS federation
	AWSSTS *AWSFederationSpec `json:"awsSts,omitempty" koanf:"awsSts"`

	// Optional JSON schema describing user-supplied credentials/config fields.
	CredentialsSchema map[string]any `json:"credentialsSchema,omitempty" koanf:"credentialsSchema"`

	// Storage flags/policies
	Persistence *PersistenceSpec `json:"persistence,omitempty" koanf:"persistence"`

	// Optional: default labels/tags you might want to apply
	Labels map[string]string `json:"labels,omitempty" koanf:"labels"`
}

type OAuthSpec struct {
	// Standard OAuth2
	ClientID     string   `json:"clientId"      koanf:"clientId"`
	ClientSecret string   `json:"clientSecret"  koanf:"clientSecret"`
	AuthURL      string   `json:"authUrl"       koanf:"authUrl"`
	TokenURL     string   `json:"tokenUrl"      koanf:"tokenUrl"`
	Scopes       []string `json:"scopes"        koanf:"scopes"`

	// Optional OIDC discovery (if provided, take precedence to resolve AuthURL/TokenURL)
	// Useful when AuthType == oidc.
	OIDCDiscoveryURL string `json:"oidcDiscoveryUrl,omitempty" koanf:"oidcDiscoveryUrl"`

	// Redirect URI selection (if multiple frontends/environments)
	RedirectURI string `json:"redirectUri,omitempty" koanf:"redirectUri"`

	// PKCE enforcement
	UsePKCE bool `json:"usePkce,omitempty" koanf:"usePkce"`

	// Extra query params to add to the authorization request (e.g., access_type=offline)
	AuthParams map[string]string `json:"authParams,omitempty" koanf:"authParams"`

	// Extra token params if required by provider
	TokenParams map[string]string `json:"tokenParams,omitempty" koanf:"tokenParams"`
}

type APIKeySpec struct {
	// Simple API key or header name; drives form rendering for non-OAuth integrations.
	KeyLabel   string `json:"keyLabel"   koanf:"keyLabel"`   // "API Key"
	HeaderName string `json:"headerName" koanf:"headerName"` // "X-API-KEY"
	// In case some systems use query params
	QueryParam string `json:"queryParam,omitempty" koanf:"queryParam"`
}

type UserInfoSpec struct {
	URL        string          `json:"url"         koanf:"url"`
	Method     HTTPMethod      `json:"method"      koanf:"method"`
	AuthStyle  AuthHeaderStyle `json:"authStyle"   koanf:"authStyle"`
	AuthHeader string          `json:"authHeader,omitempty" koanf:"authHeader"` // override default header name ("Authorization")
	IDPath     JSONFieldPath   `json:"idPath"      koanf:"idPath"`              // e.g., "id" or "data.user.id"
	EmailPath  JSONFieldPath   `json:"emailPath"   koanf:"emailPath"`
	LoginPath  JSONFieldPath   `json:"loginPath"   koanf:"loginPath"` // username/display
	// Some providers (e.g., GitHub) require an additional email endpoint
	SecondaryEmailURL string `json:"secondaryEmailUrl,omitempty" koanf:"secondaryEmailUrl"`
}

type PersistenceSpec struct {
	// Where/how to store long-lived tokens (prefer none; see design).
	StoreRefreshToken bool          `json:"storeRefreshToken" koanf:"storeRefreshToken"`
}

// WorkloadIdentitySpec describes defaults used for GCP Workload Identity Federation integrations.
type WorkloadIdentitySpec struct {
	// Optional default audience for STS token exchange. Can be overridden per-tenant.
	Audience string `json:"audience,omitempty" koanf:"audience"`
	// Optional default service account email to impersonate.
	TargetServiceAccount string `json:"targetServiceAccount,omitempty" koanf:"targetServiceAccount"`
	// Optional default scopes for the generated access token.
	Scopes []string `json:"scopes,omitempty" koanf:"scopes"`
	// Optional requested access token lifetime (e.g., "30m").
	TokenLifetime time.Duration `json:"tokenLifetime,omitempty" koanf:"tokenLifetime"`
	// Subject token type when exchanging with STS (default id_token).
	SubjectTokenType string `json:"subjectTokenType,omitempty" koanf:"subjectTokenType"`
}

// GitHubAppSpec captures metadata needed to mint installation tokens.
type GitHubAppSpec struct {
	// Optional custom API base URL (for GitHub Enterprise).
	BaseURL string `json:"baseUrl,omitempty" koanf:"baseUrl"`
	// Optional default token lifetime hint.
	TokenTTL time.Duration `json:"tokenTtl,omitempty" koanf:"tokenTtl"`
}

// AWSFederationSpec captures defaults for AWS STS AssumeRoleWithWebIdentity flows.
type AWSFederationSpec struct {
	RoleARN     string        `json:"roleArn,omitempty" koanf:"roleArn"`
	SessionName string        `json:"sessionName,omitempty" koanf:"sessionName"`
	Duration    time.Duration `json:"duration,omitempty" koanf:"duration"`
	Region      string        `json:"region,omitempty" koanf:"region"`
	ExternalID  string        `json:"externalId,omitempty" koanf:"externalId"`
}
