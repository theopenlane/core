package email

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/newman"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// SendEmailRequest is the operation config for dispatching a single templated email.
// The template is resolved by database ID; its Key selects the catalog entry that owns
// the render pipeline and its Defaults supplies the pre-filled typed fields
type SendEmailRequest struct {
	// TemplateID references the email template by database ID
	TemplateID string `json:"templateId" jsonschema:"required,description=Email template ID"`
	// OwnerID is the organization owner for template resolution
	OwnerID string `json:"ownerId" jsonschema:"required,description=Organization owner ID"`
	// To is the recipient email address
	To string `json:"to" jsonschema:"required,description=Recipient email address"`
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

// Run resolves the email template, looks up the catalog dispatcher by Key, builds a typed
// payload from the template defaults + per-invocation recipient, and sends through the dispatcher
func (SendEmail) Run(ctx context.Context, req types.OperationRequest, client *EmailClient, cfg SendEmailRequest) (json.RawMessage, error) {
	template, err := loadEmailTemplate(ctx, req.DB, cfg.OwnerID, cfg.TemplateID)
	if err != nil {
		return nil, err
	}

	dispatcher, ok := DispatcherByKey(template.Key)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDispatcherNotFound, template.Key)
	}

	payload, err := buildDispatchPayload(template.Defaults,
		RecipientInfo{Email: cfg.To, Tags: cfg.Tags},
	)
	if err != nil {
		return nil, err
	}

	extraOpts := []newman.MessageOption{
		newman.WithAttachments(staticAttachmentsFromFiles(ctx, template.Edges.Files)),
	}
	if cfg.ReplyTo != "" {
		extraOpts = append(extraOpts, newman.WithReplyTo(cfg.ReplyTo))
	}

	if err := dispatcher.SendByKey(ctx, req, client, payload, extraOpts...); err != nil {
		return nil, err
	}

	return nil, nil
}
