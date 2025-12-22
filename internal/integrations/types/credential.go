package types //nolint:revive

import (
	"strings"

	"github.com/samber/lo"
	"github.com/samber/mo"
	"golang.org/x/oauth2"

	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/theopenlane/core/internal/integrations/helpers"
	"github.com/theopenlane/core/pkg/models"
)

// CredentialKind distinguishes the various credential payloads managed by the broker
type CredentialKind string

const (
	// CredentialKindOAuthToken represents OAuth2 token credentials
	CredentialKindOAuthToken CredentialKind = "oauth_token" //nolint:gosec
	// CredentialKindMetadata represents integration metadata credentials
	CredentialKindMetadata CredentialKind = "integration_metadata"
	// CredentialKindAPIKey represents API key credentials
	CredentialKindAPIKey CredentialKind = "api_key"
	// CredentialKindWorkload represents workload identity credentials
	CredentialKindWorkload CredentialKind = "workload_identity"
)

// CredentialPayload is the canonical envelope exchanged between keymaker and keystore.
// It embeds the upstream token/claims types when applicable and stores the
// persisted credential set for non-OAuth flows.
type CredentialPayload struct {
	// Provider identifies the source provider for the credential
	Provider ProviderType `json:"provider"`
	// Kind indicates the credential kind (oauth token, metadata, etc)
	Kind CredentialKind `json:"kind"`
	// Token optionally embeds the upstream oauth2 token
	Token *oauth2.Token `json:"token,omitempty"`
	// Claims optionally carries upstream OIDC claims
	Claims *oidc.IDTokenClaims `json:"claims,omitempty"`
	// Data stores provider-specific credential data serialized via CredentialSet
	Data models.CredentialSet `json:"credential"`
}

// CredentialOption mutates the payload being built.
type CredentialOption func(*CredentialPayload)

// BuildCredentialPayload applies the provided options and enforces invariants.
func BuildCredentialPayload(provider ProviderType, opts ...CredentialOption) (CredentialPayload, error) {
	payload := CredentialPayload{
		Provider: provider,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&payload)
	}

	if payload.Provider == ProviderUnknown {
		return CredentialPayload{}, ErrProviderTypeRequired
	}

	if isCredentialSetEmpty(payload.Data) && payload.Token == nil {
		return CredentialPayload{}, ErrCredentialSetRequired
	}

	if payload.Kind == "" {
		payload.Kind = CredentialKindOAuthToken
	}

	return payload, nil
}

// CredentialBuilder offers a fluent API around BuildCredentialPayload.
type CredentialBuilder struct {
	provider ProviderType
	options  []CredentialOption
}

// NewCredentialBuilder initializes a builder for the given provider.
func NewCredentialBuilder(provider ProviderType) *CredentialBuilder {
	return &CredentialBuilder{provider: provider}
}

// With appends options to the builder.
func (b *CredentialBuilder) With(opts ...CredentialOption) *CredentialBuilder {
	b.options = append(b.options, opts...)

	return b
}

// Build returns the final payload.
func (b *CredentialBuilder) Build() (CredentialPayload, error) {
	return BuildCredentialPayload(b.provider, b.options...)
}

// WithCredentialKind overrides the automatically inferred kind.
func WithCredentialKind(kind CredentialKind) CredentialOption {
	return func(payload *CredentialPayload) {
		payload.Kind = kind
	}
}

// WithCredentialSet sets the stored credential data directly.
func WithCredentialSet(set models.CredentialSet) CredentialOption {
	return func(payload *CredentialPayload) {
		payload.Data = set
	}
}

// WithCredential allows providers to encode their specific credential structs via encoder.
func WithCredential[T any](value T, encoder func(T) models.CredentialSet) CredentialOption {
	return func(payload *CredentialPayload) {
		if encoder == nil {
			return
		}

		payload.Data = encoder(value)
	}
}

// WithOAuthToken embeds the upstream oauth2.Token.
func WithOAuthToken(token *oauth2.Token) CredentialOption {
	return func(payload *CredentialPayload) {
		payload.Token = helpers.CloneOAuthToken(token)
	}
}

// WithOIDCClaims embeds the upstream OIDC claims struct.
func WithOIDCClaims(claims *oidc.IDTokenClaims) CredentialOption {
	return func(payload *CredentialPayload) {
		payload.Claims = helpers.CloneOIDCClaims(claims)
	}
}

// MergeScopes merges scopes from an oauth2.Token into a plain slice (helpful for persistence).
func MergeScopes(dest []string, source ...string) []string {
	filtered := lo.Map(source, func(item string, _ int) string {
		return strings.TrimSpace(item)
	})

	nonEmpty := lo.Filter(filtered, func(item string, _ int) bool {
		return item != ""
	})

	return lo.Uniq(append(dest, nonEmpty...))
}

// Redacted returns a shallow copy of the payload with sensitive fields cleared
// Useful for logging and telemetry without leaking credentials
func (p CredentialPayload) Redacted() CredentialPayload {
	clone := CredentialPayload{
		Provider: p.Provider,
		Kind:     p.Kind,
	}

	if token := helpers.CloneOAuthToken(p.Token); token != nil {
		token.AccessToken = "[redacted]"
		token.RefreshToken = ""
		clone.Token = token
	}

	if claims := helpers.CloneOIDCClaims(p.Claims); claims != nil {
		clone.Claims = claims
	}
	// Intentionally drop Data; CredentialSet contains sensitive material
	return clone
}

// OAuthTokenOption returns a cloned token wrapped in an Option
func (p CredentialPayload) OAuthTokenOption() mo.Option[*oauth2.Token] {
	return optionFromPointer(helpers.CloneOAuthToken(p.Token))
}

// ClaimsOption returns cloned OIDC claims wrapped in an Option
func (p CredentialPayload) ClaimsOption() mo.Option[*oidc.IDTokenClaims] {
	return optionFromPointer(helpers.CloneOIDCClaims(p.Claims))
}

// optionFromPointer wraps a pointer in an Option, returning None if nil
func optionFromPointer[T any](value *T) mo.Option[*T] {
	if value == nil {
		return mo.None[*T]()
	}

	return mo.Some(value)
}

// isCredentialSetEmpty checks if all fields in a credential set are empty
func isCredentialSetEmpty(set models.CredentialSet) bool {
	fields := []string{
		set.AccessKeyID,
		set.SecretAccessKey,
		set.ProjectID,
		set.AccountID,
		set.APIToken,
		set.OAuthAccessToken,
		set.OAuthRefreshToken,
		set.OAuthTokenType,
	}

	for _, field := range fields {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}

	if len(set.ProviderData) > 0 {
		return false
	}

	if set.OAuthExpiry != nil {
		return false
	}

	if len(set.Claims) > 0 {
		return false
	}

	return true
}
