package slack

import (
	"context"
	"encoding/json"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// MessageOperationInput holds per-invocation parameters for the message.send operation
type MessageOperationInput struct {
	// Channel is the Slack channel identifier to post to
	Channel string `json:"channel" jsonschema:"required,title=Channel"`
	// Text is the plain-text message content
	Text string `json:"text,omitempty" jsonschema:"title=Message Text"`
	// Blocks is a Block Kit payload for rich messages
	Blocks []json.RawMessage `json:"blocks,omitempty" jsonschema:"title=Block Kit Payload"`
	// ThreadTS is the thread timestamp for replies
	ThreadTS string `json:"thread_ts,omitempty" jsonschema:"title=Thread Timestamp"`
	// UnfurlLinks controls whether links are unfurled
	UnfurlLinks *bool `json:"unfurl_links,omitempty" jsonschema:"title=Unfurl Links"`
}

// MessageSend sends a Slack message via chat.postMessage
type MessageSend struct {
	// Channel is the channel the message was posted to
	Channel string `json:"channel"`
	// TS is the message timestamp
	TS string `json:"ts"`
}

// Handle adapts message send to the generic operation registration boundary
func (m MessageSend) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		var cfg MessageOperationInput
		if err := jsonx.UnmarshalIfPresent(request.Config, &cfg); err != nil {
			return nil, ErrOperationConfigInvalid
		}

		return m.Run(ctx, c, cfg)
	}
}

// Run sends a Slack message via chat.postMessage
func (MessageSend) Run(ctx context.Context, c *slackgo.Client, cfg MessageOperationInput) (json.RawMessage, error) {
	if cfg.Channel == "" {
		return nil, ErrChannelMissing
	}

	hasText := cfg.Text != ""
	hasBlocks := len(cfg.Blocks) > 0

	if !hasText && !hasBlocks {
		return nil, ErrMessageEmpty
	}

	opts := []slackgo.MsgOption{slackgo.MsgOptionAsUser(true)}

	if hasText {
		opts = append(opts, slackgo.MsgOptionText(cfg.Text, false))
	}

	if cfg.ThreadTS != "" {
		opts = append(opts, slackgo.MsgOptionTS(cfg.ThreadTS))
	}

	if cfg.UnfurlLinks != nil && !*cfg.UnfurlLinks {
		opts = append(opts, slackgo.MsgOptionDisableLinkUnfurl())
	}

	respChannel, ts, err := c.PostMessageContext(ctx, cfg.Channel, opts...)
	if err != nil {
		return nil, ErrMessageSendFailed
	}

	return providerkit.EncodeResult(MessageSend{
		Channel: respChannel,
		TS:      ts,
	}, ErrResultEncode)
}
