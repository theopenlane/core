package oidcgeneric

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var oidcGenericScopes = []string{
	"openid",
	"profile",
	"email",
	"offline_access",
}

// startInstallAuth starts the Generic OIDC install flow
func (d *def) startInstallAuth(ctx context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
	return providerkit.StartOAuthFlow(ctx, providerkit.OAuthFlowConfig{
		ClientID:     d.cfg.ClientID,
		ClientSecret: d.cfg.ClientSecret,
		DiscoveryURL: d.cfg.DiscoveryURL,
		RedirectURL:  d.cfg.RedirectURL,
		Scopes:       oidcGenericScopes,
	})
}

// completeInstallAuth completes the Generic OIDC install flow
func (d *def) completeInstallAuth(ctx context.Context, state json.RawMessage, input json.RawMessage) (types.AuthCompleteResult, error) {
	return providerkit.CompleteOAuthFlow(ctx, providerkit.OAuthFlowConfig{
		ClientID:     d.cfg.ClientID,
		ClientSecret: d.cfg.ClientSecret,
		DiscoveryURL: d.cfg.DiscoveryURL,
		RedirectURL:  d.cfg.RedirectURL,
		Scopes:       oidcGenericScopes,
	}, state, input)
}
