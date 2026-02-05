package oauth

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"

	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/providers"
)

// Session implements types.AuthSession for OAuth flows
type Session struct {
	provider *Provider
	state    string
	authURL  string
}

// ProviderType returns the provider identifier
func (s *Session) ProviderType() types.ProviderType {
	return s.provider.providerType
}

// State returns the authorization state value
func (s *Session) State() string {
	return s.state
}

// AuthURL returns the URL the client should redirect to
func (s *Session) AuthURL() string {
	return s.authURL
}

// Finish exchanges the authorization code for a credential payload
func (s *Session) Finish(ctx context.Context, code string) (types.CredentialPayload, error) {
	codeOpts := make([]rp.CodeExchangeOpt, 0, len(s.provider.tokenParams))
	for key, value := range s.provider.tokenParams {
		k := key
		v := value
		codeOpts = append(codeOpts, func() []oauth2.AuthCodeOption {
			return []oauth2.AuthCodeOption{oauth2.SetAuthURLParam(k, v)}
		})
	}

	tokens, err := rp.CodeExchange[*oidc.IDTokenClaims](ctx, code, s.provider.relyingParty, codeOpts...)
	if err != nil {
		return types.CredentialPayload{}, fmt.Errorf("%w: %w", providers.ErrCodeExchange, err)
	}

	builder := types.NewCredentialBuilder(s.provider.providerType).
		With(
			types.WithCredentialSet(models.CredentialSet{}),
			types.WithOAuthToken(tokens.Token),
		)
	if tokens.IDTokenClaims != nil {
		builder = builder.With(types.WithOIDCClaims(tokens.IDTokenClaims))
	}

	payload, buildErr := builder.Build()
	if buildErr != nil {
		return types.CredentialPayload{}, buildErr
	}

	return payload, nil
}
