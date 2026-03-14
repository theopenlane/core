package slack

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	slackAuthURL  = "https://slack.com/oauth/v2/authorize"
	slackTokenURL = "https://slack.com/api/oauth.v2.access"
)

var slackScopes = []string{
	"chat:write",
	"chat:write.public",
	"chat:write.customize",
	"team:read",
	"users:read",
}

// startInstallAuth starts the Slack OAuth install flow
func (d *def) startInstallAuth(ctx context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
	return providerkit.StartOAuthFlow(ctx, providerkit.OAuthFlowConfig{
		ClientID:     d.cfg.ClientID,
		ClientSecret: d.cfg.ClientSecret,
		AuthURL:      slackAuthURL,
		TokenURL:     slackTokenURL,
		RedirectURL:  d.cfg.RedirectURL,
		Scopes:       slackScopes,
	})
}

// completeInstallAuth completes the Slack OAuth install flow
func (d *def) completeInstallAuth(ctx context.Context, state json.RawMessage, input json.RawMessage) (types.AuthCompleteResult, error) {
	return providerkit.CompleteOAuthFlow(ctx, providerkit.OAuthFlowConfig{
		ClientID:     d.cfg.ClientID,
		ClientSecret: d.cfg.ClientSecret,
		AuthURL:      slackAuthURL,
		TokenURL:     slackTokenURL,
		RedirectURL:  d.cfg.RedirectURL,
		Scopes:       slackScopes,
	}, state, input)
}
