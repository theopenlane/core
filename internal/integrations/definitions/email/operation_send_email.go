package email

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/newman"

	"github.com/theopenlane/core/v2/internal/integrations/templatekit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// SendEmailOp is the operation ref for the generic send-email operation
var SendEmailOp = types.OperationRefOf[SendEmailRequest]().Handles(emailClientRef, sendEmail).Policy(types.ExecutionPolicy{SkipRunRecord: true}) //nolint:revive

// SendEmailRequest is the operation config for dispatching a single templated email
type SendEmailRequest struct {
	// TemplateID references the email template by database ID
	TemplateID string `json:"templateId" jsonschema:"required,description=Email template ID"`
	// OwnerID is the organization owner for template resolution
	OwnerID string `json:"ownerId" jsonschema:"required,description=Organization owner ID"`
	// To is the recipient email address
	To string `json:"to" jsonschema:"required,description=Recipient email address"`
	// Tags are delivery metadata tags for provider webhook correlation
	Tags []newman.Tag `json:"tags,omitempty" jsonschema:"description=Delivery tracking tags"`
	// From is an optional sender address override
	From string `json:"from,omitempty" jsonschema:"description=Sender email address override"`
	// ReplyTo is an optional reply-to address
	ReplyTo string `json:"replyTo,omitempty" jsonschema:"description=Reply-to email address"`
}

// sendEmail resolves the email template, looks up the catalog dispatcher by Key, builds a typed payload from the template defaults + per-invocation recipient, and sends through the dispatcher
func sendEmail(ctx context.Context, req types.OperationRequest, client *Client, cfg SendEmailRequest) (json.RawMessage, error) {
	template, err := loadEmailTemplate(ctx, req.DB, cfg.OwnerID, cfg.TemplateID)
	if err != nil {
		return nil, err
	}

	dispatcher, ok := DispatcherByKey(template.Key)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDispatcherNotFound, template.Key)
	}

	payload, err := templatekit.BuildDispatchPayload(template.Defaults, RecipientInfo{Email: cfg.To, Tags: cfg.Tags})
	if err != nil {
		return nil, err
	}

	extraOpts := []newman.MessageOption{
		newman.WithAttachments(staticAttachmentsFromFiles(ctx, template.Edges.Files)),
	}
	if cfg.ReplyTo != "" {
		extraOpts = append(extraOpts, newman.WithReplyTo(cfg.ReplyTo))
	}
	if cfg.From != "" {
		extraOpts = append(extraOpts, newman.WithFrom(cfg.From))
	}

	if err := dispatcher.SendByKey(ctx, req, client, payload, extraOpts...); err != nil {
		return nil, err
	}

	return nil, nil
}
