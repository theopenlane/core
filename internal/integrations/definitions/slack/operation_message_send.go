package slack

import (
	"context"
	"encoding/json"

	"github.com/samber/lo"
	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/templatekit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// MessageSendOperation holds per-invocation parameters for the message.send operation
type MessageSendOperation struct {
	// TemplateID references a notification template by database ID
	TemplateID string `json:"templateId,omitempty" jsonschema:"title=Template ID"`
	// TemplateKey references a notification template by key
	TemplateKey string `json:"templateKey,omitempty" jsonschema:"title=Template Key"`
	// Channel is the Slack channel identifier to post to
	Channel string `json:"channel,omitempty" jsonschema:"title=Channel"`
	// Destinations are Slack channel identifiers to post the same message to
	Destinations []string `json:"destinations,omitempty" jsonschema:"title=Destinations"`
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
	// Channel is the first channel the message was posted to
	Channel string `json:"channel,omitempty"`
	// TS is the first message timestamp
	TS string `json:"ts,omitempty"`
	// Deliveries captures every channel delivery performed by the operation
	Deliveries []MessageDelivery `json:"deliveries,omitempty"`
}

// MessageDelivery captures one Slack message delivery
type MessageDelivery struct {
	// Channel is the channel the message was posted to
	Channel string `json:"channel"`
	// TS is the message timestamp
	TS string `json:"ts"`
}

// Handle adapts message send to the generic operation registration boundary
func (m MessageSend) Handle() types.OperationHandler {
	return providerkit.WithClientRequestConfig(slackClient, MessageSendOp, ErrOperationConfigInvalid, m.Run)
}

// Run sends a Slack message via chat.postMessage
func (MessageSend) Run(ctx context.Context, req types.OperationRequest, c *SlackClient, cfg MessageSendOperation) (json.RawMessage, error) {
	if err := templatekit.ResolveOperationTemplate(ctx, req, cfg.TemplateID, cfg.TemplateKey, &cfg); err != nil {
		return nil, err
	}

	destinations := lo.Uniq(lo.Compact(append([]string{cfg.Channel}, cfg.Destinations...)))
	if len(destinations) == 0 {
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

	if hasBlocks {
		var blocks slackgo.Blocks
		encoded, err := json.Marshal(cfg.Blocks)
		if err != nil {
			return nil, ErrOperationConfigInvalid
		}
		if err := json.Unmarshal(encoded, &blocks); err != nil {
			return nil, ErrOperationConfigInvalid
		}

		opts = append(opts, slackgo.MsgOptionBlocks(blocks.BlockSet...))
	}

	if cfg.ThreadTS != "" {
		opts = append(opts, slackgo.MsgOptionTS(cfg.ThreadTS))
	}

	if cfg.UnfurlLinks != nil && !*cfg.UnfurlLinks {
		opts = append(opts, slackgo.MsgOptionDisableLinkUnfurl())
	}

	deliveries := make([]MessageDelivery, 0, len(destinations))
	for _, destination := range destinations {
		respChannel, ts, err := c.API.PostMessageContext(ctx, destination, opts...)
		if err != nil {
			return nil, ErrMessageSendFailed
		}

		deliveries = append(deliveries, MessageDelivery{
			Channel: respChannel,
			TS:      ts,
		})
	}

	result := MessageSend{
		Deliveries: deliveries,
	}
	if len(deliveries) > 0 {
		result.Channel = deliveries[0].Channel
		result.TS = deliveries[0].TS
	}

	return providerkit.EncodeResult(result, ErrResultEncode)
}
