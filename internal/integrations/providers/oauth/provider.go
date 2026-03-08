package oauth

import (
	"context"
	"maps"

	"golang.org/x/oauth2"

	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/theopenlane/core/common/models"
	integrationauth "github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	iamauth "github.com/theopenlane/iam/auth"
)

const (
	stateLength = 32
)

// Provider implements the types.Provider interface using Zitadel's relying party helpers
type Provider struct {
	// BaseProvider holds shared provider metadata
	providers.BaseProvider
	spec         config.ProviderSpec
	oauthConfig  *oauth2.Config
	relyingParty rp.RelyingParty
	authParams   map[string]string
	tokenParams  map[string]string
}

// ProviderOption customizes OAuth provider construction.
type ProviderOption func(*Provider)

// WithClientDescriptors registers client descriptors for pooling.
func WithClientDescriptors(descriptors []types.ClientDescriptor) ProviderOption {
	return func(p *Provider) {
		p.Clients = providerkit.SanitizeClientDescriptors(p.Type(), descriptors)
	}
}

// New constructs a Provider from the supplied spec
func New(spec config.ProviderSpec, options ...ProviderOption) (*Provider, error) {
	if spec.OAuth == nil {
		return nil, providers.ErrSpecOAuthRequired
	}
	authKind := spec.AuthType.Normalize()
	if !authKind.SupportsInteractiveFlow() {
		return nil, ErrAuthTypeMismatch
	}

	cfg := &oauth2.Config{
		ClientID:     spec.OAuth.ClientID,
		ClientSecret: spec.OAuth.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  spec.OAuth.AuthURL,
			TokenURL: spec.OAuth.TokenURL,
		},
		RedirectURL: spec.OAuth.RedirectURI,
		Scopes:      spec.OAuth.Scopes,
	}

	var rpOpts []rp.Option
	if spec.OAuth.UsePKCE {
		rpOpts = append(rpOpts, rp.WithPKCE(nil))
	}

	rparty, err := rp.NewRelyingPartyOAuth(cfg, rpOpts...)
	if err != nil {
		return nil, providers.ErrRelyingPartyInit
	}

	caps := types.ProviderCapabilities{
		SupportsRefreshTokens: true,
		SupportsClientPooling: true,
		SupportsMetadataForm:  len(spec.CredentialsSchema) > 0,
	}

	provider := &Provider{
		BaseProvider: providers.NewBaseProvider(spec.ProviderType(), caps, nil, nil),
		spec:         spec,
		oauthConfig:  cfg,
		relyingParty: rparty,
		authParams:   maps.Clone(spec.OAuth.AuthParams),
		tokenParams:  maps.Clone(spec.OAuth.TokenParams),
	}

	for _, opt := range options {
		if opt != nil {
			opt(provider)
		}
	}

	return provider, nil
}

// BeginAuth starts an OAuth authorization flow
func (p *Provider) BeginAuth(_ context.Context, input types.AuthContext) (types.AuthSession, error) {
	scopes := p.spec.OAuth.Scopes
	if len(input.Scopes) > 0 {
		scopes = input.Scopes
	}

	state := input.State

	if state == "" {
		generated, err := iamauth.GenerateOAuthState(stateLength)
		if err != nil {
			return nil, providers.ErrStateGeneration
		}

		state = generated
	}

	authOpts := buildAuthURLOpts(scopes, p.authParams)

	authURL := rp.AuthURL(state, p.relyingParty, authOpts...)

	session := &Session{
		provider: p,
		state:    state,
		authURL:  authURL,
	}
	return session, nil
}

// Mint refreshes an access token using the stored credential set.
func (p *Provider) Mint(ctx context.Context, subject types.CredentialMintRequest) (models.CredentialSet, error) {
	if subject.Credential.OAuthAccessToken == "" && subject.Credential.OAuthRefreshToken == "" {
		return models.CredentialSet{}, providers.ErrTokenUnavailable
	}

	token := &oauth2.Token{
		AccessToken:  subject.Credential.OAuthAccessToken,
		RefreshToken: subject.Credential.OAuthRefreshToken,
		TokenType:    subject.Credential.OAuthTokenType,
	}
	if subject.Credential.OAuthExpiry != nil {
		token.Expiry = subject.Credential.OAuthExpiry.UTC()
	}

	tokenSource := p.oauthConfig.TokenSource(ctx, token)
	freshToken, err := tokenSource.Token()
	if err != nil {
		return models.CredentialSet{}, providers.ErrTokenRefresh
	}

	var claims *oidc.IDTokenClaims
	if len(subject.Credential.Claims) > 0 {
		var decoded oidc.IDTokenClaims
		if err := jsonx.RoundTrip(subject.Credential.Claims, &decoded); err == nil {
			claims = &decoded
		}
	}

	credential, err := integrationauth.BuildOAuthCredentialSet(freshToken, claims)
	if err != nil {
		return models.CredentialSet{}, err
	}

	return credential, nil
}

// WithOperations configures provider-managed operations.
func WithOperations(descriptors []types.OperationDescriptor) ProviderOption {
	return func(p *Provider) {
		p.Ops = providerkit.SanitizeOperationDescriptors(p.Type(), descriptors)
	}
}
