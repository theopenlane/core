package oauth

import (
	"context"

	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/theopenlane/core/internal/integrations/types"
)

// Session implements types.AuthSession for OAuth2/OIDC authorization flows
type Session struct {
	// provider is the oauth.Provider that created this session
	provider *Provider
	// state is the CSRF state value for this authorization transaction
	state string
	// authURL is the provider authorization endpoint URL the client should redirect to
	authURL string
}

// ProviderType returns the provider identifier for this session
func (s *Session) ProviderType() types.ProviderType {
	return s.provider.Type()
}

// State returns the CSRF state value
func (s *Session) State() string {
	return s.state
}

// AuthURL returns the URL the client should redirect to for authorization
func (s *Session) AuthURL() string {
	return s.authURL
}

// Finish exchanges the authorization code for a credential set
func (s *Session) Finish(ctx context.Context, code string) (types.CredentialSet, error) {
	tokens, err := rp.CodeExchange[*oidc.IDTokenClaims](ctx, code, s.provider.relyingParty, buildCodeExchangeOpts(s.provider.tokenParams)...)
	if err != nil {
		return types.CredentialSet{}, ErrCodeExchange
	}

	return buildCredentialSet(tokens.Token, tokens.IDTokenClaims)
}
