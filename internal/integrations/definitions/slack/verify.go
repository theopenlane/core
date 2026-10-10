package slack

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/auth"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// incomingWebhookKey is the token response field Slack fills with the channel chosen during app install
const incomingWebhookKey = "incoming_webhook"

// incomingWebhookChannelKey is the incoming webhook field carrying the chosen channel id
const incomingWebhookChannelKey = "channel_id"

// verify probes the workspace with auth.test and returns the installation identity with its default channel
func verify[T any](ctx context.Context, req types.ConnectionRequest[T], c *SlackClient) (InstallationMetadata, error) {
	resp, err := c.API.AuthTestContext(ctx)
	if err != nil {
		return InstallationMetadata{}, ErrAuthTestFailed
	}

	if resp.TeamID == "" && resp.Team == "" {
		return InstallationMetadata{}, ErrTeamIDMissing
	}

	channel, err := defaultChannel(req)
	if err != nil {
		return InstallationMetadata{}, err
	}

	return InstallationMetadata{TeamID: resp.TeamID, TeamName: resp.Team, DefaultChannel: channel}, nil
}

// defaultChannel prefers the channel chosen during app install over the one supplied as user input
func defaultChannel[T any](req types.ConnectionRequest[T]) (string, error) {
	if cred, ok := any(req.Credential).(slackCred); ok && cred.DefaultChannel != "" {
		return cred.DefaultChannel, nil
	}

	var input UserInput
	if err := jsonx.UnmarshalIfPresent(req.Integration.UserInput.Data, &input); err != nil {
		return "", ErrUserInputDecode
	}

	return input.DefaultChannel, nil
}

// installChannel returns the channel id Slack reports on the token response when the installer picked one
func installChannel(material auth.OAuthMaterial) string {
	hook, ok := material.ExtraValue(incomingWebhookKey).(map[string]any)
	if !ok {
		return ""
	}

	channel, _ := hook[incomingWebhookChannelKey].(string)

	return channel
}
