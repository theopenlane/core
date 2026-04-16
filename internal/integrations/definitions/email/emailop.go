package email

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// HasRecipient is implemented by all email operation input types to provide recipient information
type HasRecipient interface {
	// GetRecipient returns the RecipientInfo
	GetRecipient() RecipientInfo
}

// RecipientInfo holds recipient addressing fields embedded in every email operation input
type RecipientInfo struct {
	// Email is the recipient email address
	Email string `json:"email" jsonschema:"required,description=Recipient email address"`
	// FirstName is the recipient first name
	FirstName string `json:"first_name,omitempty" jsonschema:"description=Recipient first name"`
	// LastName is the recipient last name
	LastName string `json:"last_name,omitempty" jsonschema:"description=Recipient last name"`
	// Tags are delivery tracking tags forwarded to the email provider for webhook correlation
	Tags []newman.Tag `json:"tags,omitempty" jsonschema:"description=Delivery tracking tags"`
}

// GetRecipient returns the recipient info, satisfying the HasRecipient interface
func (r RecipientInfo) GetRecipient() RecipientInfo {
	return r
}

// EmailOperation is a generic helper which defines a single system email type as a registered integration operation
// this allows us to do AllEmailOperations() in the builder rather than manually wiring each
type EmailOperation[T HasRecipient] struct {
	// Op is the typed operation ref with name derived from the schema definition key
	Op types.OperationRef[T]
	// Schema is the reflected JSON schema for the input type
	Schema json.RawMessage
	// Subject returns the rendered subject line for the email
	Subject func(cfg RuntimeEmailConfig, input T) string
	// Theme is the newman render theme applied to this email
	Theme *render.Theme
	// Content returns the structured email content for newman rendering
	Content func(cfg RuntimeEmailConfig, input T) render.EmailContent
	// MessageOptions returns additional newman message options for per-operation customization such as attachment
	MessageOptions func(cfg RuntimeEmailConfig, input T) []newman.MessageOption
	// PreHook is an optional hook invoked before rendering to resolve dynamic fields
	// When set, the handler uses WithClientRequestConfig to pass the full OperationRequest through
	PreHook func(ctx context.Context, req types.OperationRequest, client *EmailClient, input *T) error
}

// Registration returns the types.OperationRegistration for wiring into the definition builder
func (e EmailOperation[T]) Registration() types.OperationRegistration {
	return types.OperationRegistration{
		Name:         e.Op.Name(),
		Topic:        DefinitionID.OperationTopic(e.Op.Name()),
		ClientRef:    emailClientRef.ID(),
		ConfigSchema: e.Schema,
		Handle:       e.handler(),
	}
}

// handler returns the typed operation handler that renders and sends the email
func (e EmailOperation[T]) handler() types.OperationHandler {
	return providerkit.WithClientRequestConfig(emailClientRef, e.Op, ErrTemplateRenderFailed,
		func(ctx context.Context, req types.OperationRequest, client *EmailClient, input T) (json.RawMessage, error) {
			if e.PreHook != nil {
				if err := e.PreHook(ctx, req, client, &input); err != nil {
					return nil, err
				}
			}

			var extraOpts []newman.MessageOption
			if e.MessageOptions != nil {
				extraOpts = e.MessageOptions(client.Config, input)
			}

			recipient := input.GetRecipient()
			for _, tag := range recipient.Tags {
				extraOpts = append(extraOpts, newman.WithTag(tag))
			}

			return renderAndSend(ctx, client, e.Theme,
				recipient,
				e.Subject(client.Config, input),
				e.Content(client.Config, input),
				extraOpts...,
			)
		},
	)
}

// renderAndSend renders an email using the newman render engine and sends it through the client
func renderAndSend(ctx context.Context, client *EmailClient, theme *render.Theme, recipient RecipientInfo, subject string, content render.EmailContent, extraOpts ...newman.MessageOption) (json.RawMessage, error) {
	r := render.NewRenderer(
		render.WithTheme(theme),
		render.WithBranding(BrandingFromConfig(client.Config)),
	)

	htmlBody, err := r.GenerateHTML(content)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTemplateRenderFailed, err)
	}

	textBody, err := r.GeneratePlainText(content)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTemplateRenderFailed, err)
	}

	opts := []newman.MessageOption{
		newman.WithFrom(client.Config.FromEmail),
		newman.WithTo([]string{recipient.Email}),
		newman.WithSubject(subject),
		newman.WithHTML(htmlBody),
		newman.WithText(textBody),
	}
	opts = append(opts, extraOpts...)

	message := newman.NewEmailMessageWithOptions(opts...)

	if err := client.Sender.SendEmailWithContext(ctx, message); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("op", subject).Msg("failed sending email")
		return nil, fmt.Errorf("%w: %w", ErrSendFailed, err)
	}

	return nil, nil
}
