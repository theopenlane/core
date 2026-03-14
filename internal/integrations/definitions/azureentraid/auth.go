package azureentraid

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	azureAuthURL  = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	azureTokenURL = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
)

var azureEntraScopes = []string{
	"https://graph.microsoft.com/.default",
	"offline_access",
}

// startInstallAuth starts the Azure Entra ID OAuth install flow
func (d *def) startInstallAuth(ctx context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
	return providerkit.StartOAuthFlow(ctx, providerkit.OAuthFlowConfig{
		ClientID:     d.cfg.ClientID,
		ClientSecret: d.cfg.ClientSecret,
		AuthURL:      azureAuthURL,
		TokenURL:     azureTokenURL,
		RedirectURL:  d.cfg.RedirectURL,
		Scopes:       azureEntraScopes,
	})
}

// completeInstallAuth completes the Azure Entra ID OAuth install flow
func (d *def) completeInstallAuth(ctx context.Context, state json.RawMessage, input json.RawMessage) (types.AuthCompleteResult, error) {
	return providerkit.CompleteOAuthFlow(ctx, providerkit.OAuthFlowConfig{
		ClientID:     d.cfg.ClientID,
		ClientSecret: d.cfg.ClientSecret,
		AuthURL:      azureAuthURL,
		TokenURL:     azureTokenURL,
		RedirectURL:  d.cfg.RedirectURL,
		Scopes:       azureEntraScopes,
	}, state, input)
}
