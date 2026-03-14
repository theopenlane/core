package microsoftteams

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	teamsAuthURL  = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	teamsTokenURL = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
)

var teamsScopes = []string{
	"https://graph.microsoft.com/User.Read",
	"https://graph.microsoft.com/Team.ReadBasic.All",
	"https://graph.microsoft.com/Channel.ReadBasic.All",
	"https://graph.microsoft.com/ChannelMessage.Send",
	"offline_access",
}

// startInstallAuth starts the Microsoft Teams OAuth install flow
func (d *def) startInstallAuth(ctx context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
	return providerkit.StartOAuthFlow(ctx, providerkit.OAuthFlowConfig{
		ClientID:     d.cfg.ClientID,
		ClientSecret: d.cfg.ClientSecret,
		AuthURL:      teamsAuthURL,
		TokenURL:     teamsTokenURL,
		RedirectURL:  d.cfg.RedirectURL,
		Scopes:       teamsScopes,
	})
}

// completeInstallAuth completes the Microsoft Teams OAuth install flow
func (d *def) completeInstallAuth(ctx context.Context, state json.RawMessage, input json.RawMessage) (types.AuthCompleteResult, error) {
	return providerkit.CompleteOAuthFlow(ctx, providerkit.OAuthFlowConfig{
		ClientID:     d.cfg.ClientID,
		ClientSecret: d.cfg.ClientSecret,
		AuthURL:      teamsAuthURL,
		TokenURL:     teamsTokenURL,
		RedirectURL:  d.cfg.RedirectURL,
		Scopes:       teamsScopes,
	}, state, input)
}
