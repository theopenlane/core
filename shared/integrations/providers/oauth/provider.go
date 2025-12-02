package oauth

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"golang.org/x/oauth2"

	"github.com/zitadel/oidc/v3/pkg/client/rp"

	"github.com/theopenlane/shared/integrations/config"
	"github.com/theopenlane/shared/integrations/providers"
	"github.com/theopenlane/shared/integrations/providers/helpers"
	"github.com/theopenlane/shared/integrations/types"
	"github.com/theopenlane/shared/models"
)

const (
	stateLength = 32
)

// Provider implements the types.Provider interface using Zitadel's relying party helpers
type Provider struct {
	providerType types.ProviderType
	spec         config.ProviderSpec
	oauthConfig  *oauth2.Config
	relyingParty rp.RelyingParty
	authParams   map[string]string
	tokenParams  map[string]string
	capabilities types.ProviderCapabilities
	operations   []types.OperationDescriptor
}

// ProviderOption customizes OAuth provider construction.
type ProviderOption func(*Provider)

// New constructs a Provider from the supplied spec
func New(_ context.Context, spec config.ProviderSpec, options ...ProviderOption) (*Provider, error) {
	if spec.OAuth == nil {
		return nil, providers.ErrSpecOAuthRequired
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
		return nil, fmt.Errorf("%w: %w", providers.ErrRelyingPartyInit, err)
	}

	provider := &Provider{
		providerType: spec.ProviderType(),
		spec:         spec,
		oauthConfig:  cfg,
		relyingParty: rparty,
		authParams:   lo.Assign(map[string]string{}, spec.OAuth.AuthParams),
		tokenParams:  lo.Assign(map[string]string{}, spec.OAuth.TokenParams),
		capabilities: types.ProviderCapabilities{
			SupportsRefreshTokens: true,
			SupportsClientPooling: true,
			SupportsMetadataForm:  len(spec.CredentialsSchema) > 0,
		},
	}

	for _, opt := range options {
		if opt != nil {
			opt(provider)
		}
	}

	return provider, nil
}

// Type returns the provider identifier
func (p *Provider) Type() types.ProviderType {
	return p.providerType
}

// Capabilities describes supported behaviours
func (p *Provider) Capabilities() types.ProviderCapabilities {
	return p.capabilities
}

// Operations returns the provider-published operations if configured.
func (p *Provider) Operations() []types.OperationDescriptor {
	if len(p.operations) == 0 {
		return nil
	}

	ops := make([]types.OperationDescriptor, len(p.operations))
	copy(ops, p.operations)

	return ops
}

// BeginAuth starts an OAuth authorization flow
func (p *Provider) BeginAuth(_ context.Context, input types.AuthContext) (types.AuthSession, error) {
	scopes := p.spec.OAuth.Scopes
	if len(input.Scopes) > 0 {
		scopes = input.Scopes
	}

	state := input.State

	if strings.TrimSpace(state) == "" {
		generated, err := helpers.RandomState(stateLength)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", providers.ErrStateGeneration, err)
		}

		state = generated
	}

	var authOpts []rp.AuthURLOpt

	if len(scopes) > 0 {
		scopeValue := strings.Join(scopes, " ")
		authOpts = append(authOpts, func() []oauth2.AuthCodeOption {
			return []oauth2.AuthCodeOption{oauth2.SetAuthURLParam("scope", scopeValue)}
		})
	}

	for key, value := range p.authParams {
		k := key
		v := value
		authOpts = append(authOpts, func() []oauth2.AuthCodeOption {
			return []oauth2.AuthCodeOption{oauth2.SetAuthURLParam(k, v)}
		})
	}

	authURL := rp.AuthURL(state, p.relyingParty, authOpts...)

	session := &Session{
		provider: p,
		state:    state,
		authURL:  authURL,
	}
	return session, nil
}

// Mint refreshes an access token using the stored credential payload
func (p *Provider) Mint(ctx context.Context, subject types.CredentialSubject) (types.CredentialPayload, error) {
	tokenOpt := subject.Credential.OAuthTokenOption()
	if !tokenOpt.IsPresent() {
		return types.CredentialPayload{}, providers.ErrTokenUnavailable
	}

	tokenSource := p.oauthConfig.TokenSource(ctx, tokenOpt.MustGet())
	freshToken, err := tokenSource.Token()
	if err != nil {
		return types.CredentialPayload{}, fmt.Errorf("%w: %w", providers.ErrTokenRefresh, err)
	}

	builder := types.NewCredentialBuilder(p.providerType).
		With(
			types.WithCredentialSet(models.CredentialSet{}),
			types.WithOAuthToken(freshToken),
			types.WithCredentialKind(types.CredentialKindOAuthToken),
		)
	if claimOpt := subject.Credential.ClaimsOption(); claimOpt.IsPresent() {
		builder = builder.With(types.WithOIDCClaims(claimOpt.MustGet()))
	}

	payload, buildErr := builder.Build()
	if buildErr != nil {
		return types.CredentialPayload{}, buildErr
	}

	return payload, nil
}

// WithOperations configures provider-managed operations.
func WithOperations(descriptors []types.OperationDescriptor) ProviderOption {
	return func(p *Provider) {
		p.operations = helpers.SanitizeOperationDescriptors(p.providerType, descriptors)
	}
}
