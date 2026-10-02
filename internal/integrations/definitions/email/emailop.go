package email

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/v2/internal/integrations/templatekit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

// templatePlaceholderOpen marks a still-unrendered template expression in a per-recipient value
const templatePlaceholderOpen = "{{"

// oneClickUnsubscribePath is the API route the RFC 8058 List-Unsubscribe header targets
const oneClickUnsubscribePath = "/api/unsubscribe"

// Recipient is implemented by all email operation input types to provide recipient information
type Recipient interface {
	// GetRecipient returns the RecipientInfo
	GetRecipient() RecipientInfo
}

// Dispatcher describes a catalog-addressable email operation that can be invoked by key
type Dispatcher interface {
	// Name returns the catalog key identifying the dispatcher
	Name() string
	// Registration returns the integration operation registration for this entry
	Registration() types.OperationRegistration
	// SendByKey decodes the payload into the entry's typed input, runs any registered PreHook, and dispatches through the shared render/send pipeline
	SendByKey(ctx context.Context, req types.OperationRequest, client *Client, payload json.RawMessage, extraOpts ...newman.MessageOption) error
	// RenderMessage decodes the payload and renders the email into a newman message without sending it, for use in batch send paths
	RenderMessage(ctx context.Context, client *Client, payload json.RawMessage, extraOpts ...newman.MessageOption) (*newman.EmailMessage, error)
	// ExamplePayload returns a representative example payload used to render catalog previews and to seed the form preview with demo values
	ExamplePayload() json.RawMessage
	// BuilderUISchema returns the RJSF-style UI schema describing how the configurable fields render as a form, or nil when none is defined
	BuilderUISchema() json.RawMessage
}

// RecipientInfo holds recipient addressing fields embedded in every email operation input
type RecipientInfo struct {
	// Email is the recipient email address; it is the primary recipient used for personalization and footer links
	Email string `json:"email" jsonschema:"required,description=Recipient email address"`
	// Recipients optionally addresses the message to multiple recipients in a single send; when set it replaces Email as the To list
	Recipients []string `json:"recipients,omitempty" jsonschema:"description=Recipient email addresses for a single multi-recipient message; replaces the single recipient when set"`
	// FirstName is the recipient first name
	FirstName string `json:"firstName,omitempty" jsonschema:"description=Recipient first name"`
	// LastName is the recipient last name
	LastName string `json:"lastName,omitempty" jsonschema:"description=Recipient last name"`
	// UnsubscribeToken is the per-recipient token used to build an unsubscribe link in templates via {{ .unsubscribeToken }}; it is populated per send and never authored in the template
	UnsubscribeToken string `json:"unsubscribeToken,omitempty" jsonschema:"description=Per-recipient unsubscribe token"`
	// Tags are delivery tracking tags forwarded to the email provider for webhook correlation
	Tags []newman.Tag `json:"tags,omitempty" jsonschema:"description=Delivery tracking tags"`
}

// GetRecipient returns the recipient info, satisfying the Recipient interface
func (r RecipientInfo) GetRecipient() RecipientInfo {
	return r
}

// CampaignContext carries campaign-scoped metadata for catalog entries that render campaign-bound emails
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

var (
	dispatchers     []Dispatcher
	dispatcherIndex = map[string]Dispatcher{}
)

// RegisterEmailOperation constructs an Operation[T] and adds it to the dispatcher registry
func RegisterEmailOperation[T Recipient](op Operation[T]) Operation[T] {
	dispatchers = append(dispatchers, op)
	dispatcherIndex[op.Name()] = op

	return op
}

// Operation is a generic helper which defines a single system email type as a registered integration operation
type Operation[T Recipient] struct {
	// Op is the typed operation ref with name derived from the schema definition key, declared against the email client
	Op types.OperationRef[T]
	// Description is the human-readable summary shown in the catalog picker
	Description string
	// CustomerSelectable gates whether the entry is exposed via the customer-facing catalog query
	CustomerSelectable *bool
	// Example is a representative input used to render catalog previews and to seed the form preview with demo values for fields the author has not yet filled in
	Example T
	// UISchema is the RJSF-style UI schema describing how the configurable fields render as a form
	UISchema json.RawMessage
	// Subject returns the rendered subject line for the email
	Subject func(cfg RuntimeEmailConfig, input T) string
	// Theme is the newman render theme applied to this email
	Theme *render.Theme
	// Build returns the structured body content for newman rendering
	Build func(cfg RuntimeEmailConfig, input T) render.ContentBody
	// Config is an optional per-op override for the installation config
	Config func(cfg RuntimeEmailConfig, input T) RuntimeEmailConfig
	// MessageOptions returns additional newman message options for per-operation customization such as attachment
	MessageOptions func(cfg RuntimeEmailConfig, input T) []newman.MessageOption
	// PreHook is an optional hook invoked before rendering to resolve dynamic fields
	PreHook func(ctx context.Context, req types.OperationRequest, cfg RuntimeEmailConfig, input *T) error
}

// Name returns the catalog key for the operation, satisfying the Dispatcher interface
func (e Operation[T]) Name() string {
	return e.Op.Name()
}

// ExamplePayload returns the marshaled example input, satisfying the Dispatcher interface
func (e Operation[T]) ExamplePayload() json.RawMessage {
	data, err := json.Marshal(e.Example)
	if err != nil {
		return nil
	}

	return data
}

// BuilderUISchema returns the UI schema for the operation, satisfying the Dispatcher interface
func (e Operation[T]) BuilderUISchema() json.RawMessage {
	return e.UISchema
}

// decodePayload interpolates template expressions in the raw JSON and unmarshals into T
func decodePayload[T Recipient](client *Client, payload json.RawMessage) (T, error) {
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
func (e Operation[T]) SendByKey(ctx context.Context, req types.OperationRequest, client *Client, payload json.RawMessage, extraOpts ...newman.MessageOption) error {
	input, err := decodePayload[T](client, payload)
	if err != nil {
		return err
	}

	return e.dispatch(ctx, req, client, input, extraOpts...)
}

// RenderMessage decodes the payload and renders the email into a newman message without sending it
func (e Operation[T]) RenderMessage(_ context.Context, client *Client, payload json.RawMessage, extraOpts ...newman.MessageOption) (*newman.EmailMessage, error) {
	input, err := decodePayload[T](client, payload)
	if err != nil {
		return nil, err
	}

	return e.renderToMessage(client, input, extraOpts...)
}

// dispatch runs PreHook, renders the message, and sends it through the client sender
func (e Operation[T]) dispatch(ctx context.Context, req types.OperationRequest, client *Client, input T, extraOpts ...newman.MessageOption) error {
	if e.PreHook != nil {
		if err := e.PreHook(ctx, req, client.Config, &input); err != nil {
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
func (e Operation[T]) renderToMessage(client *Client, input T, extraOpts ...newman.MessageOption) (*newman.EmailMessage, error) {
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

// Registration returns the types.OperationRegistration for wiring into the definition builder
func (e Operation[T]) Registration() types.OperationRegistration {
	return e.Op.Handles(emailClientRef, func(ctx context.Context, req types.OperationRequest, client *Client, input T) (json.RawMessage, error) {
		return nil, e.dispatch(ctx, req, client, input)
	}).
		CustomerSelectable(e.CustomerSelectable != nil && *e.CustomerSelectable).
		Registration(DefinitionID, types.OperationRegistration{Description: e.Description})
}

// RenderCatalogPreview renders a customer-selectable catalog entry to HTML for UI preview
func RenderCatalogPreview(ctx context.Context, client *Client, d Dispatcher, draft map[string]any) (string, error) {
	var example map[string]any
	if err := jsonx.RoundTrip(d.ExamplePayload(), &example); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("key", d.Name()).Msg("failed decoding catalog example payload")

		return "", fmt.Errorf("%w: %w", ErrTemplateRenderFailed, err)
	}

	overlays := make([]any, 0, 1)
	if len(draft) > 0 {
		overlays = append(overlays, draft)
	}

	payload, err := templatekit.BuildDispatchPayload(example, overlays...)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("key", d.Name()).Msg("failed building catalog preview payload")

		return "", err
	}

	msg, err := d.RenderMessage(ctx, client, payload)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("key", d.Name()).Msg("failed rendering catalog preview")

		return "", err
	}

	return msg.GetHTML(), nil
}

// renderMessage renders an email into a newman message without sending it
func renderMessage(client *Client, theme *render.Theme, recipient RecipientInfo, subject string, content render.EmailContent, extraOpts ...newman.MessageOption) (*newman.EmailMessage, error) {
	r := render.NewRenderer(render.WithTheme(theme))

	htmlBody, err := r.GenerateHTML(content)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTemplateRenderFailed, err)
	}

	textBody, err := r.GeneratePlainText(content)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTemplateRenderFailed, err)
	}

	to := recipient.Recipients
	if len(to) == 0 {
		to = []string{recipient.Email}
	}

	opts := []newman.MessageOption{
		newman.WithFrom(client.Config.FromEmail),
		newman.WithTo(to),
		newman.WithSubject(subject),
		newman.WithHTML(htmlBody),
		newman.WithText(textBody),
	}
	opts = append(opts, extraOpts...)

	if cfg, ok := content.Config.(RuntimeEmailConfig); ok && cfg.UnsubscribeURL != "" && !strings.Contains(cfg.UnsubscribeURL, templatePlaceholderOpen) {
		if u, err := urlx.ParseAbsolute(cfg.UnsubscribeURL); err == nil {
			u.Path = oneClickUnsubscribePath
			u.RawPath = ""
			u.Fragment = ""

			opts = append(opts,
				newman.WithHeader("List-Unsubscribe", "<"+u.String()+">"),
				newman.WithHeader("List-Unsubscribe-Post", "List-Unsubscribe=One-Click"),
			)
		}
	}

	return newman.NewEmailMessageWithOptions(opts...), nil
}
