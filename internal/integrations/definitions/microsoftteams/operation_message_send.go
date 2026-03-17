package microsoftteams

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// MessageOperationInput holds per-invocation parameters for the message.send operation
type MessageOperationInput struct {
	// TeamID is the target Microsoft Teams team identifier
	TeamID string `json:"team_id" jsonschema:"required,title=Team ID"`
	// ChannelID is the target Teams channel identifier
	ChannelID string `json:"channel_id" jsonschema:"required,title=Channel ID"`
	// Body is the message body content
	Body string `json:"body" jsonschema:"required,title=Message Body"`
	// BodyFormat controls the content type (text or html)
	BodyFormat string `json:"body_format,omitempty" jsonschema:"title=Body Format"`
	// Subject is an optional message subject
	Subject string `json:"subject,omitempty" jsonschema:"title=Subject"`
}

// MessageSend sends a Microsoft Teams channel message
type MessageSend struct {
	// TeamID is the target team identifier
	TeamID string `json:"teamId"`
	// ChannelID is the target channel identifier
	ChannelID string `json:"channelId"`
	// MessageID is the identifier of the created message
	MessageID string `json:"messageId"`
}

// channelMessageBody is the Graph API chat message body payload
type channelMessageBody struct {
	// ContentType is the content type of the message body (text or html)
	ContentType string `json:"contentType"`
	// Content is the message body text
	Content string `json:"content"`
}

// channelMessageRequest is the Graph API request payload for creating a channel message
type channelMessageRequest struct {
	// Body holds the message body content and format
	Body channelMessageBody `json:"body"`
	// Subject is the optional message subject line
	Subject string `json:"subject,omitempty"`
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

// Run sends a Microsoft Teams channel message via Microsoft Graph
func (MessageSend) Run(ctx context.Context, c *providerkit.AuthenticatedClient, cfg MessageOperationInput) (json.RawMessage, error) {
	if cfg.TeamID == "" || cfg.ChannelID == "" {
		return nil, ErrChannelMissing
	}

	if cfg.Body == "" {
		return nil, ErrMessageEmpty
	}

	contentType := cfg.BodyFormat
	if contentType == "" {
		contentType = "text"
	}

	switch contentType {
	case "text", "html":
	default:
		return nil, ErrBodyFormatInvalid
	}

	payload := channelMessageRequest{
		Body: channelMessageBody{
			ContentType: contentType,
			Content:     cfg.Body,
		},
		Subject: cfg.Subject,
	}

	path := fmt.Sprintf("teams/%s/channels/%s/messages", url.PathEscape(cfg.TeamID), url.PathEscape(cfg.ChannelID))

	var resp struct {
		ID string `json:"id"`
	}

	if err := c.PostJSON(ctx, path, payload, &resp); err != nil {
		return nil, ErrChannelMessageSendFailed
	}

	return providerkit.EncodeResult(MessageSend{
		TeamID:    cfg.TeamID,
		ChannelID: cfg.ChannelID,
		MessageID: resp.ID,
	}, ErrResultEncode)
}
