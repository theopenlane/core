package email

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/newman"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// SendEmailRequest is the operation config for dispatching a single templated email.
// The template is resolved by ID from the database and rendered with the supplied data
type SendEmailRequest struct {
	// TemplateID references the email template by database ID
	TemplateID string `json:"templateId" jsonschema:"required,description=Email template ID"`
	// OwnerID is the organization owner for template resolution
	OwnerID string `json:"ownerId" jsonschema:"required,description=Organization owner ID"`
	// To is the list of recipient email addresses
	To []string `json:"to" jsonschema:"required,description=Recipient email addresses"`
	// Data contains template rendering variables
	Data map[string]any `json:"data,omitempty" jsonschema:"description=Template rendering variables"`
	// Tags are delivery metadata tags for provider webhook correlation
	Tags []newman.Tag `json:"tags,omitempty" jsonschema:"description=Delivery tracking tags"`
	// ReplyTo is an optional reply-to address
	ReplyTo string `json:"replyTo,omitempty" jsonschema:"description=Reply-to email address"`
}

// SendEmail dispatches a single templated email through the resolved email client
type SendEmail struct{}

// Handle returns the typed operation handler for builder registration
func (s SendEmail) Handle() types.OperationHandler {
	return providerkit.WithClientRequestConfig(emailClientRef, SendEmailOp, ErrTemplateRenderFailed, s.Run)
}

// Run resolves the email template, renders it, and sends through the provided client
func (SendEmail) Run(ctx context.Context, req types.OperationRequest, client *EmailClient, cfg SendEmailRequest) (json.RawMessage, error) {
	if len(cfg.To) == 0 {
		return nil, ErrMissingRecipientAddress
	}

	emailRecord, err := loadEmailTemplate(ctx, req.DB, cfg.OwnerID, cfg.TemplateID)
	if err != nil {
		return nil, err
	}

	data, err := buildTemplateData(client.Config, cfg.Data)
	if err != nil {
		return nil, err
	}

	if err := validateTemplateData(emailRecord.Jsonconfig, data); err != nil {
		return nil, err
	}

	rendered, err := renderDBEnvelope(emailRecord, data, nil)
	if err != nil {
		return nil, err
	}

	opts := []newman.MessageOption{
		newman.WithFrom(client.Config.FromEmail),
		newman.WithTo(cfg.To),
		newman.WithSubject(rendered.Subject),
		newman.WithHTML(rendered.HTML),
		newman.WithText(rendered.Text),
	}

	for _, tag := range cfg.Tags {
		opts = append(opts, newman.WithTag(tag))
	}

	if cfg.ReplyTo != "" {
		opts = append(opts, newman.WithReplyTo(cfg.ReplyTo))
	}

	opts = append(opts, newman.WithAttachments(staticAttachmentsFromFiles(ctx, emailRecord.Edges.Files)))

	message := newman.NewEmailMessageWithOptions(opts...)

	if err := client.Sender.SendEmailWithContext(ctx, message); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("template_id", cfg.TemplateID).Msg("failed sending email")

		return nil, fmt.Errorf("%w: %w", ErrSendFailed, err)
	}

	return nil, nil
}
