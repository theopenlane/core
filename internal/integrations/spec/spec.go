package spec

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// DefaultSchemaVersion is the schema version assigned when specs omit the field
const DefaultSchemaVersion = "v1"

// supportedSchemaVersions enumerates the schema versions recognized by the loader
var supportedSchemaVersions = map[string]struct{}{
	DefaultSchemaVersion: {},
}

// ProviderSpec is the provider-agnostic declarative spec loaded from JSON files or Go structs.
// It carries only fields that are meaningful across all providers. Provider-specific operator
// configuration (e.g. GitHub App credentials, Google WIF audience) belongs in the provider's
// own package and is stored in ProviderConfig as raw JSON, decoded by the provider at runtime.
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
	// OAuth contains OAuth2/OIDC protocol configuration; present for any provider using
	// AuthKindOAuth2, AuthKindOAuth2ClientCredentials, or AuthKindOIDC
	OAuth *OAuthSpec `json:"oauth,omitempty" koanf:"oauth"`
	// UserInfo describes optional post-auth userinfo lookups for OAuth2/OIDC providers
	UserInfo *UserInfoSpec `json:"userInfo,omitempty"`
	// CredentialsSchema drives declarative credential forms
	CredentialsSchema json.RawMessage `json:"credentialsSchema,omitempty"`
	// Persistence configures storage policies
	Persistence *PersistenceSpec `json:"persistence,omitempty"`
	// Labels carries optional metadata labels
	Labels map[string]string `json:"labels,omitempty"`
	// Metadata stores additional provider metadata
	Metadata json.RawMessage `json:"metadata,omitempty"`
	// SuccessRedirectURL is the URL to redirect to after successful authentication for this provider
	SuccessRedirectURL string `json:"successRedirectUrl,omitempty" koanf:"successredirecturl"`
	// ProviderConfig holds provider-specific operator configuration as raw JSON.
	// Each provider package defines its own typed struct and decodes this field at startup.
	// Example: providers/github defines GitHubAppConfig and calls
	// jsonx.UnmarshalInto(spec.ProviderConfig, &cfg) in its builder.
	ProviderConfig json.RawMessage `json:"providerConfig,omitempty" koanf:"providerconfig"`
}

// ProviderType returns the normalized provider identifier
func (s ProviderSpec) ProviderType() types.ProviderType {
	return types.ProviderTypeFromString(s.Name)
}

// SupportsInteractiveAuthFlow reports whether the spec supports browser OAuth/OIDC callbacks
func (s ProviderSpec) SupportsInteractiveAuthFlow() bool {
	return s.OAuth != nil && s.AuthType.SupportsInteractiveFlow()
}

// ToProviderMetadata exposes the subset of spec fields used by registries and handlers
func (s ProviderSpec) ToProviderMetadata() types.IntegrationProviderMetadata {
	meta := types.IntegrationProviderMetadata{
		Name:              s.Name,
		DisplayName:       s.DisplayName,
		Category:          s.Category,
		Description:       s.Description,
		AuthType:          s.AuthType,
		AuthStartPath:     s.AuthStartPath,
		AuthCallbackPath:  s.AuthCallbackPath,
		Tags:              s.Tags,
		LogoURL:           s.LogoURL,
		DocsURL:           s.DocsURL,
		CredentialsSchema: jsonx.CloneRawMessage(s.CredentialsSchema),
		Labels:            s.Labels,
	}

	if s.Active != nil {
		meta.Active = *s.Active
	}

	if s.Visible != nil {
		meta.Visible = *s.Visible
	}

	if s.OAuth != nil {
		meta.OAuth = &types.OAuthPublicConfig{
			ClientID:    s.OAuth.ClientID,
			AuthURL:     s.OAuth.AuthURL,
			TokenURL:    s.OAuth.TokenURL,
			RedirectURI: s.OAuth.RedirectURI,
			Scopes:      s.OAuth.Scopes,
			UsePKCE:     s.OAuth.UsePKCE,
			AuthParams:  s.OAuth.AuthParams,
			TokenParams: s.OAuth.TokenParams,
		}
	}

	if s.Persistence != nil {
		meta.Persistence = &types.PersistenceConfig{
			StoreRefreshToken: s.Persistence.StoreRefreshToken,
		}
	}

	return meta
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

		next, err := jsonx.ApplyOverlay(current, override)
		if err != nil {
			logx.FromContext(ctx).Warn().Err(err).Str("provider", key).Msg("failed to apply provider spec override")
			continue
		}

		merged[provider] = next
	}

	return merged
}

// ProviderStubFromSpec returns a copy of spec where all pointer-to-struct fields with a
// koanf tag are guaranteed to be non-nil, initializing nil fields to zero-value structs.
// This allows schema generators to traverse the full type structure.
func ProviderStubFromSpec(spec ProviderSpec) ProviderSpec {
	result := spec
	rv := reflect.ValueOf(&result).Elem()
	rt := rv.Type()

	for i := range rt.NumField() {
		field := rt.Field(i)

		_, hasKoanf := field.Tag.Lookup("koanf")
		if !hasKoanf {
			continue
		}

		if field.Type.Kind() != reflect.Pointer {
			continue
		}

		if field.Type.Elem().Kind() != reflect.Struct {
			continue
		}

		fv := rv.Field(i)
		if fv.IsNil() {
			fv.Set(reflect.New(field.Type.Elem()))
		}
	}

	return result
}

// supportsSchemaVersion checks if the spec declares a recognized schema version
func (s *ProviderSpec) supportsSchemaVersion() bool {
	if s == nil {
		return false
	}

	_, ok := supportedSchemaVersions[lo.CoalesceOrEmpty(s.SchemaVersion, DefaultSchemaVersion)]

	return ok
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

// PersistenceSpec controls how secrets are stored
type PersistenceSpec struct {
	// StoreRefreshToken indicates refresh tokens should be persisted
	StoreRefreshToken bool `json:"storeRefreshToken"`
}
