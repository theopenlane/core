package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"time"

	"golang.org/x/oauth2"

	iamauth "github.com/theopenlane/iam/auth"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// OAuthConfig describes OAuth2 or OIDC endpoint configuration for an integration auth flow
// each individual provider will have their own respective fields (like ClientID, ClientSecret) so these aren't duplications
type OAuthConfig struct {
	// ClientID is the OAuth application client identifier
	ClientID string
	// ClientSecret is the OAuth application client secret
	ClientSecret string
	// AuthURL is the authorization endpoint URL; leave empty when DiscoveryURL is set
	AuthURL string
	// TokenURL is the token endpoint URL; leave empty when DiscoveryURL is set
	TokenURL string
	// DiscoveryURL is the OIDC issuer URL used for endpoint discovery
	DiscoveryURL string
	// RedirectURL is typically the callback URL registered with the OAuth provider
	RedirectURL string
	// Scopes lists the OAuth scopes to request
	Scopes []string
	// AuthParams holds extra query parameters appended to the authorization URL
	AuthParams map[string]string
	// TokenParams holds extra parameters sent during code exchange
	TokenParams map[string]string
}

// OAuthMaterial holds the credential material produced by a completed OAuth flow
type OAuthMaterial struct {
	// AccessToken is the OAuth2 access token
	AccessToken string
	// RefreshToken is the OAuth2 refresh token, if provided
	RefreshToken string
	// Expiry is the access token expiration time, if known
	Expiry *time.Time
	// Claims holds decoded OIDC ID token claims, if present
	Claims map[string]any
}

// oauthStartState carries the CSRF state value stored between start and complete
type oauthStartState struct {
	// State is the random CSRF token embedded in the OAuth authorization URL
	State string `json:"state"`
}

// StartOAuth builds an authorization URL for the given config and returns an AuthStartResult
func StartOAuth(ctx context.Context, cfg OAuthConfig) (types.AuthStartResult, error) {
	rparty, err := buildRelyingParty(ctx, cfg)
	if err != nil {
		return types.AuthStartResult{}, err
	}

	csrfState, err := iamauth.GenerateOAuthState(0)
	if err != nil {
		return types.AuthStartResult{}, ErrOAuthStateGeneration
	}

	authURL := rp.AuthURL(csrfState, rparty, mapAuthCodeOptions[rp.AuthURLOpt](cfg.AuthParams)...)

	stateData, err := jsonx.ToRawMessage(oauthStartState{State: csrfState})
	if err != nil {
		return types.AuthStartResult{}, err
	}

	return types.AuthStartResult{
		URL:   authURL,
		State: stateData,
	}, nil
}

// CompleteOAuth exchanges an authorization code for OAuth credential material
func CompleteOAuth(ctx context.Context, cfg OAuthConfig, state json.RawMessage, input types.AuthCallbackInput) (OAuthMaterial, error) {
	var startState oauthStartState
	if err := jsonx.UnmarshalIfPresent(state, &startState); err != nil {
		return OAuthMaterial{}, ErrOAuthStateInvalid
	}

	code := input.First("code")
	callbackState := input.First("state")
	if code == "" {
		return OAuthMaterial{}, ErrOAuthCodeMissing
	}

	if startState.State != "" && callbackState != startState.State {
		return OAuthMaterial{}, ErrOAuthStateMismatch
	}

	rparty, err := buildRelyingParty(ctx, cfg)
	if err != nil {
		return OAuthMaterial{}, err
	}

	tokens, err := rp.CodeExchange[*oidc.IDTokenClaims](ctx, code, rparty, mapAuthCodeOptions[rp.CodeExchangeOpt](cfg.TokenParams)...)
	if err != nil {
		return OAuthMaterial{}, fmt.Errorf("%w: %w", ErrOAuthCodeExchange, err)
	}

	return buildOAuthMaterial(tokens.Token, tokens.IDTokenClaims)
}

// buildRelyingParty constructs a Zitadel relying party from the given OAuthConfig
func buildRelyingParty(ctx context.Context, cfg OAuthConfig) (rp.RelyingParty, error) {
	if cfg.DiscoveryURL != "" {
		rparty, err := rp.NewRelyingPartyOIDC(ctx, cfg.DiscoveryURL, cfg.ClientID, cfg.ClientSecret, cfg.RedirectURL, cfg.Scopes)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrOAuthRelyingPartyInit, err)
		}

		return rparty, nil
	}

	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthURL,
			TokenURL: cfg.TokenURL,
		},
		RedirectURL: cfg.RedirectURL,
		Scopes:      cfg.Scopes,
	}

	rparty, err := rp.NewRelyingPartyOAuth(oauthCfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOAuthRelyingPartyInit, err)
	}

	return rparty, nil
}

// buildOAuthMaterial constructs an OAuthMaterial from an oauth2 token and optional OIDC claims
func buildOAuthMaterial(token *oauth2.Token, claims *oidc.IDTokenClaims) (OAuthMaterial, error) {
	var mat OAuthMaterial

	if token != nil {
		mat.AccessToken = token.AccessToken
		mat.RefreshToken = token.RefreshToken

		if !token.Expiry.IsZero() {
			exp := token.Expiry.UTC()
			mat.Expiry = &exp
		}
	}

	if claims != nil {
		claimsMap, err := jsonx.ToMap(claims)
		if err != nil {
			return OAuthMaterial{}, ErrOAuthClaimsEncode
		}

		mat.Claims = claimsMap
	}

	return mat, nil
}

// mapAuthCodeOptions converts a string map into a sorted slice of auth code options
func mapAuthCodeOptions[T ~func() []oauth2.AuthCodeOption](params map[string]string) []T {
	if len(params) == 0 {
		return nil
	}

	opts := make([]T, 0, len(params))

	for _, key := range slices.Sorted(maps.Keys(params)) {
		opts = append(opts, asAuthCodeOption[T](oauth2.SetAuthURLParam(key, params[key])))
	}

	return opts
}

// asAuthCodeOption wraps a single oauth2.AuthCodeOption into the target option function type
func asAuthCodeOption[T ~func() []oauth2.AuthCodeOption](option oauth2.AuthCodeOption) T {
	return T(func() []oauth2.AuthCodeOption {
		return []oauth2.AuthCodeOption{option}
	})
}
