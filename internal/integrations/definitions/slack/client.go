package slack

import (
	"context"
	"encoding/json"
	"fmt"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// Client builds Slack clients for one customer installation
type Client struct{}

// Build constructs the unified SlackClient for one customer installation
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	token, err := resolveAccessToken(req.Credentials)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
	}

	var metadata InstallationMetadata
	if req.Integration != nil && len(req.Integration.Metadata) > 0 {
		raw, err := jsonx.ToRawMessage(req.Integration.Metadata)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
		}

		if err := jsonx.UnmarshalIfPresent(raw, &metadata); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrClientBuildFailed, err)
		}
	}

	return &SlackClient{
		API:            slackgo.New(token),
		DefaultChannel: metadata.DefaultChannel,
	}, nil
}

// buildRuntimeSlackClient constructs a SlackClient for the runtime (system) path.
// When a bot token is configured the client gets full Web API access (channel targeting,
// Block Kit); otherwise it falls back to the incoming webhook for fire-and-forget delivery
func buildRuntimeSlackClient(_ context.Context, config json.RawMessage) (any, error) {
	var cfg RuntimeSlackConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRuntimeConfigDecode, err)
	}

	if !cfg.Provisioned() {
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

// sendText delivers a plain-text system message through the client's active transport.
// When an API client is available it posts via chat.postMessage (supports channel targeting
// and richer formatting); otherwise it falls back to the incoming webhook
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

	return ErrDefaultChannelMissing
}

// resolveAccessToken returns a usable Slack API token from whichever credential slot is bound
func resolveAccessToken(bindings types.CredentialBindings) (string, error) {
	if oauthCred, ok, err := slackCredential.Resolve(bindings); err != nil {
		return "", ErrCredentialDecode
	} else if ok {
		if oauthCred.AccessToken == "" {
			return "", ErrOAuthTokenMissing
		}

		return oauthCred.AccessToken, nil
	}

	if botCred, ok, err := slackBotTokenCredential.Resolve(bindings); err != nil {
		return "", ErrCredentialDecode
	} else if ok {
		if botCred.BotToken == "" {
			return "", ErrBotTokenMissing
		}

		return botCred.BotToken, nil
	}

	return "", ErrNoCredentialResolved
}
