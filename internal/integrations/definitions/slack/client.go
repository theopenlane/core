package slack

import (
	"context"
	"encoding/json"
	"fmt"

	slackgo "github.com/slack-go/slack"

	generated "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// oauthClient builds the SlackClient from the stored OAuth credential
func oauthClient(_ context.Context, req types.ConnectionRequest[slackCred]) (*SlackClient, error) {
	if req.Credential.AccessToken == "" {
		return nil, ErrOAuthTokenMissing
	}

	return newSlackClient(req.Credential.AccessToken, req.Integration)
}

// botTokenClient builds the SlackClient from the stored bot token credential
func botTokenClient(_ context.Context, req types.ConnectionRequest[slackBotTokenCred]) (*SlackClient, error) {
	if req.Credential.BotToken == "" {
		return nil, ErrBotTokenMissing
	}

	return newSlackClient(req.Credential.BotToken, req.Integration)
}

// newSlackClient constructs the unified SlackClient for one customer installation
func newSlackClient(token string, integration *generated.Integration) (*SlackClient, error) {
	var metadata InstallationMetadata
	if err := jsonx.UnmarshalIfPresent(integration.InstallationMetadata.Attributes, &metadata); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
	}

	return &SlackClient{
		API:            slackgo.New(token),
		DefaultChannel: metadata.DefaultChannel,
	}, nil
}

// runtimeSlackClientBuilder builds a SlackClient for the runtime system path
func runtimeSlackClientBuilder(devMode bool) func(context.Context, json.RawMessage) (any, error) {
	return func(_ context.Context, config json.RawMessage) (any, error) {
		var cfg RuntimeSlackConfig
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrRuntimeConfigDecode, err)
		}

		if !cfg.Provisioned() {
			if devMode {
				return &SlackClient{devMode: true}, nil
			}

			return nil, ErrRuntimeConfigInvalid
		}

		client := &SlackClient{
			WebhookURL:     cfg.WebhookURL,
			DefaultChannel: cfg.DefaultChannel,
		}

		if cfg.BotToken != "" {
			client.API = slackgo.New(cfg.BotToken)
		}

		return client, nil
	}
}

// sendText delivers a plain-text system message through the active transport
func (c *SlackClient) sendText(ctx context.Context, text, channel string) error {
	if text == "" {
		return ErrMessageEmpty
	}

	if c.API != nil {
		if channel == "" {
			channel = c.DefaultChannel
		}

		if channel == "" {
			return ErrDefaultChannelMissing
		}

		if _, _, err := c.API.PostMessageContext(ctx, channel, slackgo.MsgOptionText(text, false)); err != nil {
			return fmt.Errorf("%w: %w", ErrMessageSendFailed, err)
		}

		return nil
	}

	if c.WebhookURL != "" {
		if err := slackgo.PostWebhookContext(ctx, c.WebhookURL, &slackgo.WebhookMessage{Text: text}); err != nil {
			return fmt.Errorf("%w: %w", ErrMessageSendFailed, err)
		}

		return nil
	}

	if c.devMode {
		return nil
	}

	return ErrDefaultChannelMissing
}
