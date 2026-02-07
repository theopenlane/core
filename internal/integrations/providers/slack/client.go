package slack

import (
	"context"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientSlackAPI identifies the Slack HTTP API client.
	ClientSlackAPI types.ClientName = "api"
)

// slackClientDescriptors returns the client descriptors published by Slack.
func slackClientDescriptors() []types.ClientDescriptor {
	return auth.DefaultClientDescriptors(TypeSlack, ClientSlackAPI, "Slack Web API client", buildSlackClient)
}

// buildSlackClient constructs an authenticated Slack API client.
func buildSlackClient(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	token, err := auth.OAuthTokenFromPayload(payload, string(TypeSlack))
	if err != nil {
		return nil, err
	}

	return auth.NewAuthenticatedClient(token, nil), nil
}
