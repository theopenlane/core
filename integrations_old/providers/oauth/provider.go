package oauth

import (
	"context"
	"maps"

	"golang.org/x/oauth2"

	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	iamauth "github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	// stateLength is the byte length used when generating a random OAuth state value
	stateLength = 32
)

// Provider implements types.Provider for OAuth2 and OIDC flows using Zitadel's relying party helpers
type Provider struct {
	// BaseProvider holds shared provider metadata
	providers.BaseProvider
	// spec is the full provider spec used to construct this instance
	spec spec.ProviderSpec
	// oauthConfig is the golang.org/x/oauth2 config built from the spec
	oauthConfig *oauth2.Config
	// relyingParty is the Zitadel RP instance used for auth URL generation and code exchange
	relyingParty rp.RelyingParty
	// authParams holds extra parameters appended to authorization requests
	authParams map[string]string
	// tokenParams holds extra parameters appended to token exchange requests
	tokenParams map[string]string
}

// Option customizes OAuth provider construction
type Option func(*Provider)

// WithClientDescriptors registers client descriptors on the provider for downstream pooling
func WithClientDescriptors(descriptors []types.ClientDescriptor) Option {
	return func(p *Provider) {
		p.Clients = providerkit.SanitizeClientDescriptors(p.Type(), descriptors)
	}
}

// WithOperations registers operation descriptors on the provider
func WithOperations(descriptors []types.OperationDescriptor) Option {
	return func(p *Provider) {
		p.Ops = providerkit.SanitizeOperationDescriptors(p.Type(), descriptors)
	}
}

// New constructs a Provider from the supplied spec, applying any options
func New(s spec.ProviderSpec, options ...Option) (*Provider, error) {
	if s.OAuth == nil {
		return nil, ErrSpecOAuthRequired
	}

	if !s.AuthType.Normalize().SupportsInteractiveFlow() {
		return nil, ErrAuthTypeMismatch
	}

	cfg := &oauth2.Config{
		ClientID:     s.OAuth.ClientID,
		ClientSecret: s.OAuth.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  s.OAuth.AuthURL,
			TokenURL: s.OAuth.TokenURL,
		},
		RedirectURL: s.OAuth.RedirectURI,
		Scopes:      s.OAuth.Scopes,
	}

	var rpOpts []rp.Option
	if s.OAuth.UsePKCE {
		rpOpts = append(rpOpts, rp.WithPKCE(nil))
	}

	rparty, err := rp.NewRelyingPartyOAuth(cfg, rpOpts...)
	if err != nil {
		return nil, ErrRelyingPartyInit
	}

	caps := types.ProviderCapabilities{
		SupportsRefreshTokens: true,
		SupportsClientPooling: true,
		SupportsMetadataForm:  len(s.CredentialsSchema) > 0,
	}

	p := &Provider{
		BaseProvider: providers.NewBaseProvider(s.ProviderType(), caps, nil, nil),
		spec:         s,
		oauthConfig:  cfg,
		relyingParty: rparty,
		authParams:   maps.Clone(s.OAuth.AuthParams),
		tokenParams:  maps.Clone(s.OAuth.TokenParams),
	}

	for _, opt := range options {
		if opt != nil {
			opt(p)
		}
	}

	return p, nil
}

// BeginAuth starts an OAuth authorization flow, returning a Session the caller can redirect to
func (p *Provider) BeginAuth(_ context.Context, input types.AuthContext) (types.AuthSession, error) {
	scopes := p.spec.OAuth.Scopes
	if len(input.Scopes) > 0 {
		scopes = input.Scopes
	}

	state := input.State
	if state == "" {
		generated, err := iamauth.GenerateOAuthState(stateLength)
		if err != nil {
			return nil, ErrStateGeneration
		}

		state = generated
	}

	authURL := rp.AuthURL(state, p.relyingParty, buildAuthURLOpts(scopes, p.authParams)...)

	return &Session{
		provider: p,
		state:    state,
		authURL:  authURL,
	}, nil
}

// Mint refreshes an access token using the stored credential
func (p *Provider) Mint(ctx context.Context, req types.CredentialMintRequest) (types.CredentialSet, error) {
	if req.Credential.OAuthAccessToken == "" && req.Credential.OAuthRefreshToken == "" {
		return types.CredentialSet{}, ErrTokenUnavailable
	}

	token := &oauth2.Token{
		AccessToken:  req.Credential.OAuthAccessToken,
		RefreshToken: req.Credential.OAuthRefreshToken,
		TokenType:    req.Credential.OAuthTokenType,
	}

	if req.Credential.OAuthExpiry != nil {
		token.Expiry = req.Credential.OAuthExpiry.UTC()
	}

	fresh, err := p.oauthConfig.TokenSource(ctx, token).Token()
	if err != nil {
		return types.CredentialSet{}, ErrTokenRefresh
	}

	var claims *oidc.IDTokenClaims
	if len(req.Credential.Claims) > 0 {
		var decoded oidc.IDTokenClaims
		if err := jsonx.RoundTrip(req.Credential.Claims, &decoded); err == nil {
			claims = &decoded
		}
	}

	return buildCredentialSet(fresh, claims)
}

// buildCredentialSet constructs a types.CredentialSet from an oauth2 token and optional OIDC claims
func buildCredentialSet(token *oauth2.Token, claims *oidc.IDTokenClaims) (types.CredentialSet, error) {
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
			return types.CredentialSet{}, ErrClaimsEncode
		}

		cs.Claims = claimsMap
	}

	return cs, nil
}
