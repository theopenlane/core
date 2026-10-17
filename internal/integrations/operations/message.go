package operations

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/templatekit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// MessageContent is the provider-neutral message every messaging integration's send operation accepts
type MessageContent struct {
	// TemplateID references a notification template by database ID
	TemplateID string `json:"templateId,omitempty" jsonschema:"title=Template ID"`
	// TemplateKey references a notification template by key
	TemplateKey string `json:"templateKey,omitempty" jsonschema:"title=Template Key"`
	// TeamID is the workspace or team the channel belongs to, required by providers that scope channels to a team
	TeamID string `json:"teamId,omitempty" jsonschema:"title=Team ID"`
	// Channel is the channel identifier to post to
	Channel string `json:"channel,omitempty" jsonschema:"title=Channel"`
	// Text is the message content
	Text string `json:"text,omitempty" jsonschema:"title=Message Text"`
	// Subject is an optional message subject, ignored by providers without one
	Subject string `json:"subject,omitempty" jsonschema:"title=Subject"`
}

// MessageSendResult is the result every messaging integration's send operation returns
type MessageSendResult struct {
	// Channel is the first channel the message was posted to
	Channel string `json:"channel,omitempty"`
	// ID is the identifier of the first message posted
	ID string `json:"id,omitempty"`
	// Deliveries captures every channel delivery performed by the operation
	Deliveries []MessageDelivery `json:"deliveries,omitempty"`
}

// MessageDelivery captures one message delivery
type MessageDelivery struct {
	// TeamID is the team the channel belongs to, empty for providers that do not scope channels to a team
	TeamID string `json:"teamId,omitempty"`
	// Channel is the channel the message was posted to
	Channel string `json:"channel"`
	// ID is the identifier of the posted message
	ID string `json:"id"`
}

// NewMessageSendResult builds the send result from its deliveries, with the first delivery as the primary channel and message
func NewMessageSendResult(deliveries []MessageDelivery) MessageSendResult {
	result := MessageSendResult{Deliveries: deliveries}

	if len(deliveries) > 0 {
		result.Channel = deliveries[0].Channel
		result.ID = deliveries[0].ID
	}

	return result
}

// SetMessageContent replaces the content, letting shared operations fill any config that embeds it
func (m *MessageContent) SetMessageContent(content MessageContent) {
	*m = content
}

// ResolveTemplate merges the referenced notification template's defaults into cfg, the operation config embedding this content
func (m MessageContent) ResolveTemplate(ctx context.Context, req types.OperationRequest, cfg any) error {
	return templatekit.ResolveOperationTemplate(ctx, req, m.TemplateID, m.TemplateKey, cfg)
}
