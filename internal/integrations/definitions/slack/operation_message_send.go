package slack

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/samber/lo"
	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/logx"
)

// MessageSendOp is the operation ref for the Slack message send operation
var MessageSendOp = types.OperationPayloadOf[MessageSendOperation]().Handles(sendMessage).Permissions(scopes...)

// MessageSendOperation holds per-invocation parameters for the message.send operation
type MessageSendOperation struct {
	operations.MessageContent
	// Destinations are Slack channel identifiers to post the same message to
	Destinations []string `json:"destinations,omitempty" jsonschema:"title=Destinations"`
	// Blocks is a Block Kit payload for rich messages
	Blocks []json.RawMessage `json:"blocks,omitempty" jsonschema:"title=Block Kit Payload"`
	// ThreadTS is the thread timestamp for replies
	ThreadTS string `json:"thread_ts,omitempty" jsonschema:"title=Thread Timestamp"`
	// UnfurlLinks controls whether links are unfurled
	UnfurlLinks *bool `json:"unfurl_links,omitempty" jsonschema:"title=Unfurl Links"`
}

// sendMessage sends a Slack message via chat.postMessage
func sendMessage(ctx context.Context, req types.OperationRequest, c *Client, cfg MessageSendOperation) (json.RawMessage, error) {
	if err := cfg.ResolveTemplate(ctx, req, &cfg); err != nil {
		return nil, err
	}

	destinations := lo.Uniq(lo.Compact(append([]string{cfg.Channel}, cfg.Destinations...)))
	if len(destinations) == 0 {
		channel, err := installationDefaultChannel(req.Integration)
		if err != nil {
			return nil, err
		}

		if channel == "" {
			return nil, ErrChannelMissing
		}

		destinations = []string{channel}
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
			return nil, fmt.Errorf("%w: %w", types.ErrOperationConfigInvalid, err)
		}
		if err := json.Unmarshal(encoded, &blocks); err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrOperationConfigInvalid, err)
		}

		opts = append(opts, slackgo.MsgOptionBlocks(blocks.BlockSet...))
	}

	if cfg.ThreadTS != "" {
		opts = append(opts, slackgo.MsgOptionTS(cfg.ThreadTS))
	}

	if cfg.UnfurlLinks != nil && !*cfg.UnfurlLinks {
		opts = append(opts, slackgo.MsgOptionDisableLinkUnfurl())
	}

	deliveries := make([]operations.MessageDelivery, 0, len(destinations))
	for _, destination := range destinations {
		respChannel, ts, err := c.API.PostMessageContext(ctx, destination, opts...)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("channel", destination).Msg("slack: message send failed")
			return nil, ErrMessageSendFailed
		}

		deliveries = append(deliveries, operations.MessageDelivery{
			Channel: respChannel,
			ID:      ts,
		})
	}

	return providerkit.EncodeResult(operations.NewMessageSendResult(deliveries), ErrResultEncode)
}
