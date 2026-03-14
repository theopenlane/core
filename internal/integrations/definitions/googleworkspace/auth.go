package googleworkspace

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	googleAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL = "https://oauth2.googleapis.com/token"
)

var googleWorkspaceScopes = []string{
	"https://www.googleapis.com/auth/admin.directory.user.readonly",
	"https://www.googleapis.com/auth/admin.directory.group.readonly",
	"https://www.googleapis.com/auth/apps.groups.migration",
}

var googleWorkspaceAuthParams = map[string]string{
	"access_type": "offline",
	"prompt":      "consent",
}

// startInstallAuth starts the Google Workspace OAuth install flow
func (d *def) startInstallAuth(ctx context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
	return providerkit.StartOAuthFlow(ctx, providerkit.OAuthFlowConfig{
		ClientID:     d.cfg.ClientID,
		ClientSecret: d.cfg.ClientSecret,
		AuthURL:      googleAuthURL,
		TokenURL:     googleTokenURL,
		RedirectURL:  d.cfg.RedirectURL,
		Scopes:       googleWorkspaceScopes,
		AuthParams:   googleWorkspaceAuthParams,
	})
}

// completeInstallAuth completes the Google Workspace OAuth install flow
func (d *def) completeInstallAuth(ctx context.Context, state json.RawMessage, input json.RawMessage) (types.AuthCompleteResult, error) {
	return providerkit.CompleteOAuthFlow(ctx, providerkit.OAuthFlowConfig{
		ClientID:     d.cfg.ClientID,
		ClientSecret: d.cfg.ClientSecret,
		AuthURL:      googleAuthURL,
		TokenURL:     googleTokenURL,
		RedirectURL:  d.cfg.RedirectURL,
		Scopes:       googleWorkspaceScopes,
		AuthParams:   googleWorkspaceAuthParams,
	}, state, input)
}
