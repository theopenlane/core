package oauth

import (
	"context"

	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/theopenlane/core/common/models"
	integrationauth "github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Session implements types.AuthSession for OAuth flows
type Session struct {
	provider *Provider
	state    string
	authURL  string
}

// ProviderType returns the provider identifier
func (s *Session) ProviderType() types.ProviderType {
	return s.provider.Type()
}

// State returns the authorization state value
func (s *Session) State() string {
	return s.state
}

// AuthURL returns the URL the client should redirect to
func (s *Session) AuthURL() string {
	return s.authURL
}

// Finish exchanges the authorization code for a credential set.
func (s *Session) Finish(ctx context.Context, code string) (models.CredentialSet, error) {
	codeOpts := buildCodeExchangeOpts(s.provider.tokenParams)

	tokens, err := rp.CodeExchange[*oidc.IDTokenClaims](ctx, code, s.provider.relyingParty, codeOpts...)
	if err != nil {
		return models.CredentialSet{}, providers.ErrCodeExchange
	}

	credential, err := integrationauth.BuildOAuthCredentialSet(tokens.Token, tokens.IDTokenClaims)
	if err != nil {
		return models.CredentialSet{}, err
	}

	return credential, nil
}
