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

// EmailDispatcher describes a catalog-addressable email operation that can be invoked by key.
// The implementation owns typed decoding of the payload and projection into render.EmailContent;
// callers supply a raw JSON payload whose shape matches the registered input type
type EmailDispatcher interface {
	// Name returns the catalog key identifying the dispatcher
	Name() string
	// Registration returns the integration operation registration for this entry
	Registration() types.OperationRegistration
	// SendByKey decodes the payload into the entry's typed input, runs any registered PreHook,
	// and dispatches through the shared render/send pipeline. Additional newman options are
	// appended after the operation's own MessageOptions
	SendByKey(ctx context.Context, req types.OperationRequest, client *EmailClient, payload json.RawMessage, extraOpts ...newman.MessageOption) error
	// RenderMessage decodes the payload and renders the email into a newman message without sending it, for use in batch send paths
	RenderMessage(ctx context.Context, client *EmailClient, payload json.RawMessage, extraOpts ...newman.MessageOption) (*newman.EmailMessage, error)
}

// RecipientInfo holds recipient addressing fields embedded in every email operation input
type RecipientInfo struct {
	// Email is the recipient email address
	Email string `json:"email" jsonschema:"required,description=Recipient email address"`
	// FirstName is the recipient first name
	FirstName string `json:"firstName,omitempty" jsonschema:"description=Recipient first name"`
	// LastName is the recipient last name
	LastName string `json:"lastName,omitempty" jsonschema:"description=Recipient last name"`
	// Tags are delivery tracking tags forwarded to the email provider for webhook correlation
	Tags []newman.Tag `json:"tags,omitempty" jsonschema:"description=Delivery tracking tags"`
}

// GetRecipient returns the recipient info, satisfying the HasRecipient interface
func (r RecipientInfo) GetRecipient() RecipientInfo {
	return r
}

// CampaignContext carries campaign-scoped metadata for catalog entries that render campaign-bound
// emails. Entries that need campaign fields embed this struct in their input type; the campaign
// dispatcher populates the JSON overlay before calling SendByKey
type CampaignContext struct {
	// CampaignID is the identifier of the campaign producing the send
	CampaignID string `json:"campaignId,omitempty" jsonschema:"description=Campaign identifier"`
	// CampaignName is the display name of the campaign
	CampaignName string `json:"campaignName,omitempty" jsonschema:"description=Campaign display name"`
	// CampaignDescription is the long-form description of the campaign
	CampaignDescription string `json:"campaignDescription,omitempty" jsonschema:"description=Campaign description"`
	// CampaignDueDate is the response deadline for the campaign, formatted for display
	CampaignDueDate string `json:"campaignDueDate,omitempty" jsonschema:"description=Campaign response due date"`
}

// EmailOperation is a generic helper which defines a single system email type as a registered integration operation
// this allows us to do AllEmailOperations() in the builder rather than manually wiring each
type EmailOperation[T HasRecipient] struct {
	// Op is the typed operation ref with name derived from the schema definition key
	Op types.OperationRef[T]
	// Schema is the reflected JSON schema for the input type
	Schema json.RawMessage
	// Description is the human-readable summary shown in the catalog picker
	Description string
	// CustomerSelectable gates whether the entry is exposed via the customer-facing catalog query
	CustomerSelectable bool
	// UISchema is optional UI layout hints for the input form; nil when absent
	UISchema json.RawMessage
	// Subject returns the rendered subject line for the email
	Subject func(cfg RuntimeEmailConfig, input T) string
	// Theme is the newman render theme applied to this email
	Theme *render.Theme
	// Build returns the structured body content for newman rendering; the dispatcher
	// fills in Request/Config/Style slots around it at send time
	Build func(cfg RuntimeEmailConfig, input T) render.ContentBody
	// Config is an optional per-op override for the installation config. When nil,
	// dispatch passes the client's RuntimeEmailConfig through verbatim. Set this when
	// the op accepts config overrides (e.g. customer-supplied LogoURL)
	Config func(cfg RuntimeEmailConfig, input T) RuntimeEmailConfig
	// MessageOptions returns additional newman message options for per-operation customization such as attachment
	MessageOptions func(cfg RuntimeEmailConfig, input T) []newman.MessageOption
	// PreHook is an optional hook invoked before rendering to resolve dynamic fields
	// When set, the handler uses WithClientRequestConfig to pass the full OperationRequest through
	PreHook func(ctx context.Context, req types.OperationRequest, client *EmailClient, input *T) error
}

// Name returns the catalog key for the operation, satisfying the EmailDispatcher interface
func (e EmailOperation[T]) Name() string {
	return e.Op.Name()
}

// decodePayload interpolates template expressions in the raw JSON and unmarshals into T
func decodePayload[T HasRecipient](client *EmailClient, payload json.RawMessage) (T, error) {
	var input T
	if len(payload) == 0 {
		return input, nil
	}

	resolved, err := interpolatePayload(client, payload)
	if err != nil {
		return input, err
	}

	if err := json.Unmarshal(resolved, &input); err != nil {
		return input, fmt.Errorf("%w: %w", ErrTemplateRenderFailed, err)
	}

	return input, nil
}

// SendByKey decodes the payload into T and dispatches through the shared render pipeline
func (e EmailOperation[T]) SendByKey(ctx context.Context, req types.OperationRequest, client *EmailClient, payload json.RawMessage, extraOpts ...newman.MessageOption) error {
	input, err := decodePayload[T](client, payload)
	if err != nil {
		return err
	}

	return e.dispatch(ctx, req, client, input, extraOpts...)
}

// RenderMessage decodes the payload and renders the email into a newman message without sending it
func (e EmailOperation[T]) RenderMessage(ctx context.Context, client *EmailClient, payload json.RawMessage, extraOpts ...newman.MessageOption) (*newman.EmailMessage, error) {
	input, err := decodePayload[T](client, payload)
	if err != nil {
		return nil, err
	}

	return e.renderToMessage(client, input, extraOpts...)
}

// dispatch runs PreHook, assembles the per-op newman options, and invokes renderAndSend.
// It is the shared tail of both the operation-framework handler and the catalog-dispatcher
// SendByKey entry points, so the two invocation paths render identically
func (e EmailOperation[T]) dispatch(ctx context.Context, req types.OperationRequest, client *EmailClient, input T, extraOpts ...newman.MessageOption) error {
	if e.PreHook != nil {
		if err := e.PreHook(ctx, req, client, &input); err != nil {
			return err
		}
	}

	msg, err := e.renderToMessage(client, input, extraOpts...)
	if err != nil {
		return err
	}

	if err := client.Sender.SendEmailWithContext(ctx, msg); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed sending email")

		return fmt.Errorf("%w: %w", ErrSendFailed, err)
	}

	return nil
}

// renderToMessage renders the email content and returns a newman message without sending it
func (e EmailOperation[T]) renderToMessage(client *EmailClient, input T, extraOpts ...newman.MessageOption) (*newman.EmailMessage, error) {
	var opts []newman.MessageOption
	if e.MessageOptions != nil {
		opts = e.MessageOptions(client.Config, input)
	}

	recipient := input.GetRecipient()
	for _, tag := range recipient.Tags {
		opts = append(opts, newman.WithTag(tag))
	}

	opts = append(opts, extraOpts...)

	content := render.EmailContent{
		Request: input,
		Config:  client.Config,
		Body:    e.Build(client.Config, input),
	}
	if e.Config != nil {
		content.Config = e.Config(client.Config, input)
	}

	return renderMessage(client, e.Theme, recipient, e.Subject(client.Config, input), content, opts...)
}

// Registration returns the types.OperationRegistration for wiring into the definition builder.
// Catalog-facing fields (Description, UISchema, CustomerSelectable) travel on the registration
// so downstream filters (e.g. customer-facing catalog query) can operate on AllEmailOperations directly
func (e EmailOperation[T]) Registration() types.OperationRegistration {
	return types.OperationRegistration{
		Name:               e.Op.Name(),
		Description:        e.Description,
		Topic:              DefinitionID.OperationTopic(e.Op.Name()),
		ClientRef:          emailClientRef.ID(),
		ConfigSchema:       e.Schema,
		UISchema:           e.UISchema,
		CustomerSelectable: e.CustomerSelectable,
		Handle:             e.handler(),
	}
}

// handler returns the typed operation handler that renders and sends the email
func (e EmailOperation[T]) handler() types.OperationHandler {
	return providerkit.WithClientRequestConfig(emailClientRef, e.Op, ErrTemplateRenderFailed,
		func(ctx context.Context, req types.OperationRequest, client *EmailClient, input T) (json.RawMessage, error) {
			return nil, e.dispatch(ctx, req, client, input)
		},
	)
}

// renderMessage renders an email into a newman message without sending it
func renderMessage(client *EmailClient, theme *render.Theme, recipient RecipientInfo, subject string, content render.EmailContent, extraOpts ...newman.MessageOption) (*newman.EmailMessage, error) {
	r := render.NewRenderer(render.WithTheme(theme))

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

	return newman.NewEmailMessageWithOptions(opts...), nil
}
