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

// SetMessageContent replaces the content, letting shared operations fill any config that embeds it
func (m *MessageContent) SetMessageContent(content MessageContent) {
	*m = content
}

// ResolveTemplate merges the referenced notification template's defaults into cfg, the operation config embedding this content
func (m MessageContent) ResolveTemplate(ctx context.Context, req types.OperationRequest, cfg any) error {
	return templatekit.ResolveOperationTemplate(ctx, req, m.TemplateID, m.TemplateKey, cfg)
}
