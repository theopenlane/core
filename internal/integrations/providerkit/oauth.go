package providerkit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"maps"
	"slices"

	"golang.org/x/oauth2"

	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const oauthStateBytes = 32

// OAuthFlowConfig describes OAuth2 or OIDC endpoint configuration for a definition's auth flow
type OAuthFlowConfig struct {
	// ClientID is the OAuth application client identifier
	ClientID string
	// ClientSecret is the OAuth application client secret
	ClientSecret string
	// AuthURL is the authorization endpoint URL; leave empty when DiscoveryURL is set
	AuthURL string
	// TokenURL is the token endpoint URL; leave empty when DiscoveryURL is set
	TokenURL string
	// DiscoveryURL is the OIDC issuer URL used for endpoint discovery; when non-empty,
	// AuthURL and TokenURL are ignored and resolved via OIDC discovery
	DiscoveryURL string
	// RedirectURL is the callback URL registered with the OAuth provider
	RedirectURL string
	// Scopes lists the OAuth scopes to request
	Scopes []string
	// AuthParams holds extra query parameters appended to the authorization URL
	AuthParams map[string]string
	// TokenParams holds extra parameters sent during code exchange
	TokenParams map[string]string
}

// oauthStartState carries the CSRF state value stored between start and complete
type oauthStartState struct {
	// State is the random CSRF token embedded in the OAuth authorization URL
	State string `json:"state"`
}

// OAuthCallbackInput carries the standard OAuth2 callback parameters from the provider redirect
type OAuthCallbackInput struct {
	// Code is the authorization code returned by the provider
	Code string `json:"code"`
	// State is the CSRF state value echoed back by the provider for verification
	State string `json:"state"`
}

// StartOAuthFlow builds an authorization URL for the given config and returns an AuthStartResult.
// For OIDC providers, set DiscoveryURL; the context is used for the discovery HTTP request.
func StartOAuthFlow(ctx context.Context, cfg OAuthFlowConfig) (types.AuthStartResult, error) {
	rparty, err := buildRelyingParty(ctx, cfg)
	if err != nil {
		return types.AuthStartResult{}, err
	}

	csrfState, err := generateOAuthState()
	if err != nil {
		return types.AuthStartResult{}, err
	}

	authURL := rp.AuthURL(csrfState, rparty, buildAuthURLOpts(cfg.AuthParams)...)

	stateData, err := jsonx.ToRawMessage(oauthStartState{State: csrfState})
	if err != nil {
		return types.AuthStartResult{}, err
	}

	return types.AuthStartResult{
		URL:   authURL,
		State: stateData,
	}, nil
}

// CompleteOAuthFlow exchanges an authorization code for a credential set.
// state is the AuthStartResult.State returned by StartOAuthFlow.
// input should be a JSON-encoded OAuthCallbackInput from the provider's redirect.
func CompleteOAuthFlow(ctx context.Context, cfg OAuthFlowConfig, state json.RawMessage, input json.RawMessage) (types.AuthCompleteResult, error) {
	var startState oauthStartState
	if err := jsonx.UnmarshalIfPresent(state, &startState); err != nil {
		return types.AuthCompleteResult{}, ErrOAuthStateInvalid
	}

	var callback OAuthCallbackInput
	if err := jsonx.UnmarshalIfPresent(input, &callback); err != nil {
		return types.AuthCompleteResult{}, ErrOAuthCallbackInputInvalid
	}

	if callback.Code == "" {
		return types.AuthCompleteResult{}, ErrOAuthCodeMissing
	}

	if startState.State != "" && callback.State != "" && startState.State != callback.State {
		return types.AuthCompleteResult{}, ErrOAuthStateMismatch
	}

	rparty, err := buildRelyingParty(ctx, cfg)
	if err != nil {
		return types.AuthCompleteResult{}, err
	}

	tokens, err := rp.CodeExchange[*oidc.IDTokenClaims](ctx, callback.Code, rparty, buildCodeExchangeOpts(cfg.TokenParams)...)
	if err != nil {
		return types.AuthCompleteResult{}, ErrOAuthCodeExchange
	}

	credential, err := buildOAuthCredential(tokens.Token, tokens.IDTokenClaims)
	if err != nil {
		return types.AuthCompleteResult{}, err
	}

	return types.AuthCompleteResult{Credential: credential}, nil
}

// buildRelyingParty constructs a Zitadel relying party from the given OAuthFlowConfig.
// When DiscoveryURL is set, OIDC endpoint discovery is used.
func buildRelyingParty(ctx context.Context, cfg OAuthFlowConfig) (rp.RelyingParty, error) {
	if cfg.DiscoveryURL != "" {
		rparty, err := rp.NewRelyingPartyOIDC(ctx, cfg.DiscoveryURL, cfg.ClientID, cfg.ClientSecret, cfg.RedirectURL, cfg.Scopes)
		if err != nil {
			return nil, ErrOAuthRelyingPartyInit
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
		return nil, ErrOAuthRelyingPartyInit
	}

	return rparty, nil
}

// buildOAuthCredential constructs a CredentialSet from an oauth2 token and optional OIDC claims
func buildOAuthCredential(token *oauth2.Token, claims *oidc.IDTokenClaims) (types.CredentialSet, error) {
	cs := types.CredentialSet{}

	if token != nil {
		cs.OAuthAccessToken = token.AccessToken
		cs.OAuthRefreshToken = token.RefreshToken
		cs.OAuthTokenType = token.TokenType

		if !token.Expiry.IsZero() {
			exp := token.Expiry.UTC()
			cs.OAuthExpiry = &exp
		}
	}

	if claims != nil {
		claimsMap, err := jsonx.ToMap(claims)
		if err != nil {
			return types.CredentialSet{}, ErrOAuthClaimsEncode
		}

		cs.Claims = claimsMap
	}

	return cs, nil
}

// generateOAuthState produces a cryptographically random hex CSRF state value
func generateOAuthState() (string, error) {
	b := make([]byte, oauthStateBytes)
	if _, err := rand.Read(b); err != nil {
		return "", ErrOAuthStateGeneration
	}

	return hex.EncodeToString(b), nil
}

// buildAuthURLOpts converts extra auth parameters into rp.AuthURLOpt values
func buildAuthURLOpts(params map[string]string) []rp.AuthURLOpt {
	return mapAuthCodeOptions[rp.AuthURLOpt](params)
}

// buildCodeExchangeOpts converts extra token parameters into rp.CodeExchangeOpt values
func buildCodeExchangeOpts(params map[string]string) []rp.CodeExchangeOpt {
	return mapAuthCodeOptions[rp.CodeExchangeOpt](params)
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
