package slack

import (
	"context"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/types"
)

// Client builds Slack Web API clients for one installation
type Client struct{}

// Build constructs the Slack Web API client for one installation
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	if req.Credential.OAuthAccessToken == "" {
		return nil, ErrOAuthTokenMissing
	}

	return slackgo.New(req.Credential.OAuthAccessToken), nil
}

