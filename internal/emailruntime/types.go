package emailruntime

import (
	"maps"
	"strings"

	"github.com/theopenlane/newman"
)

// MetadataKey identifies keys used in EmailTemplate.Metadata.
type MetadataKey string

// String returns the metadata key as a string.
func (k MetadataKey) String() string {
	return string(k)
}

const (
	// MetadataKeyRenderMode identifies template rendering mode in EmailTemplate.Metadata.
	MetadataKeyRenderMode MetadataKey = "render_mode"
	// MetadataKeyHTMLEntrypoint identifies template entrypoint for HTML bundle rendering.
	MetadataKeyHTMLEntrypoint MetadataKey = "html_entrypoint"
	// MetadataKeyTextEntrypoint identifies template entrypoint for text bundle rendering.
	MetadataKeyTextEntrypoint MetadataKey = "text_entrypoint"
	// MetadataKeyBaseTemplate identifies a system base email template key to use for RAW_HTML assembly.
	// When set, the customer body_template is composed with the named system base template at render time.
	MetadataKeyBaseTemplate MetadataKey = "base_template_key"
)

// RenderMode controls how an email template body is rendered.
type RenderMode string

// String returns the render mode as a string.
func (m RenderMode) String() string {
	return string(m)
}

// IsValid returns true when the mode is supported by the email renderer.
func (m RenderMode) IsValid() bool {
	switch m {
	case RenderModeGoTemplateBundle, RenderModeRawHTML:
		return true
	default:
		return false
	}
}

const (
	// RenderModeGoTemplateBundle renders template fields as full Go template bundles with partials.
	RenderModeGoTemplateBundle RenderMode = "GO_TEMPLATE_BUNDLE"
	// RenderModeRawHTML renders body_template as a single HTML template.
	// If MetadataKeyBaseTemplate is set, the body_template is composed with a system base template.
	RenderModeRawHTML RenderMode = "RAW_HTML"
	// BaseTemplateEntrypoint is the named template executed when assembling RAW_HTML with a base template.
	// The base template must define a block named "content" that the customer body_template overrides.
	BaseTemplateEntrypoint = "base"
)

// TemplateKey is a stable notification template identifier used for email composition.
type TemplateKey string

// String returns the key as a string.
func (k TemplateKey) String() string {
	return string(k)
}

// IsZero reports whether the key is empty after trimming whitespace.
func (k TemplateKey) IsZero() bool {
	return strings.TrimSpace(string(k)) == ""
}

// Normalize returns the key with surrounding whitespace removed.
func (k TemplateKey) Normalize() TemplateKey {
	return TemplateKey(strings.TrimSpace(string(k)))
}

// ParseTemplateKey converts raw input into a normalized TemplateKey.
func ParseTemplateKey(raw string) TemplateKey {
	return TemplateKey(strings.TrimSpace(raw))
}

const (
	// TemplateKeyVerifyEmail is the notification template key for verification emails.
	TemplateKeyVerifyEmail TemplateKey = "verify_email"
	// TemplateKeyWelcome is the notification template key for welcome emails.
	TemplateKeyWelcome TemplateKey = "welcome"
	// TemplateKeyInvite is the notification template key for organization invite emails.
	TemplateKeyInvite TemplateKey = "invite"
	// TemplateKeyInviteJoined is the notification template key for accepted-invite emails.
	TemplateKeyInviteJoined TemplateKey = "invite_joined"
	// TemplateKeyPasswordResetRequest is the notification template key for password reset request emails.
	TemplateKeyPasswordResetRequest TemplateKey = "password_reset_request"
	// TemplateKeyPasswordResetSuccess is the notification template key for password reset confirmation emails.
	TemplateKeyPasswordResetSuccess TemplateKey = "password_reset_success"
	// TemplateKeySubscribe is the notification template key for subscription verification emails.
	TemplateKeySubscribe TemplateKey = "subscribe"
	// TemplateKeyVerifyBilling is the notification template key for billing verification emails.
	TemplateKeyVerifyBilling TemplateKey = "verify_billing"
	// TemplateKeyTrustCenterNDARequest is the notification template key for trust center NDA request emails.
	TemplateKeyTrustCenterNDARequest TemplateKey = "trust_center_nda_request"
	// TemplateKeyTrustCenterNDASigned is the notification template key for trust center NDA signed emails.
	TemplateKeyTrustCenterNDASigned TemplateKey = "trust_center_nda_signed"
	// TemplateKeyTrustCenterAuth is the notification template key for trust center auth emails.
	TemplateKeyTrustCenterAuth TemplateKey = "trust_center_auth"
	// TemplateKeyQuestionnaireAuth is the notification template key for questionnaire auth emails.
	TemplateKeyQuestionnaireAuth TemplateKey = "questionnaire_auth"
	// TemplateKeyBillingEmailChanged is the notification template key for billing email changed notifications.
	TemplateKeyBillingEmailChanged TemplateKey = "billing_email_changed"
)

// TemplateRef selects a notification template by stable key or explicit record ID.
type TemplateRef struct {
	// ID references a notification template by database ID.
	ID string
	// Key references a notification template by stable key.
	Key TemplateKey
}

// Normalize trims surrounding whitespace from reference fields.
func (r TemplateRef) Normalize() TemplateRef {
	r.ID = strings.TrimSpace(r.ID)
	r.Key = r.Key.Normalize()

	return r
}

// Validate checks that the reference contains exactly one selector.
func (r TemplateRef) Validate() error {
	ref := r.Normalize()
	hasID := ref.ID != ""
	hasKey := !ref.Key.IsZero()

	switch {
	case hasID && hasKey:
		return ErrTemplateReferenceConflict
	case !hasID && !hasKey:
		return ErrMissingTemplateReference
	default:
		return nil
	}
}

// RenderMetadata holds typed render configuration decoded from EmailTemplate.Metadata.
type RenderMetadata struct {
	// Mode controls html/body rendering behavior.
	Mode RenderMode
	// HTMLEntrypoint defines the template entrypoint when using bundle rendering.
	HTMLEntrypoint string
	// TextEntrypoint defines the text template entrypoint when using bundle rendering.
	TextEntrypoint string
	// BaseTemplateKey defines the system base email template key used for raw-html assembly.
	BaseTemplateKey string
}

// EffectiveRenderMode applies defaults when Mode is unset or invalid.
func (m RenderMetadata) EffectiveRenderMode() RenderMode {
	if m.Mode.IsValid() {
		return m.Mode
	}

	if strings.TrimSpace(m.HTMLEntrypoint) != "" {
		return RenderModeGoTemplateBundle
	}

	return RenderModeRawHTML
}

// DecodeRenderMetadata decodes typed renderer options from EmailTemplate.Metadata.
func DecodeRenderMetadata(metadata map[string]any) RenderMetadata {
	if len(metadata) == 0 {
		return RenderMetadata{}
	}

	cfg := RenderMetadata{
		Mode:            RenderMode(readMetadataValue(metadata, MetadataKeyRenderMode)),
		HTMLEntrypoint:  readMetadataValue(metadata, MetadataKeyHTMLEntrypoint),
		TextEntrypoint:  readMetadataValue(metadata, MetadataKeyTextEntrypoint),
		BaseTemplateKey: readMetadataValue(metadata, MetadataKeyBaseTemplate),
	}

	cfg.Mode = RenderMode(strings.TrimSpace(cfg.Mode.String()))
	cfg.HTMLEntrypoint = strings.TrimSpace(cfg.HTMLEntrypoint)
	cfg.TextEntrypoint = strings.TrimSpace(cfg.TextEntrypoint)
	cfg.BaseTemplateKey = strings.TrimSpace(cfg.BaseTemplateKey)

	return cfg
}

// readMetadataValue reads a string value from metadata by typed key.
func readMetadataValue(metadata map[string]any, key MetadataKey) string {
	if len(metadata) == 0 {
		return ""
	}

	value, ok := metadata[key.String()]
	if !ok || value == nil {
		return ""
	}

	typed, ok := value.(string)
	if !ok {
		return ""
	}

	return typed
}

// SendOption is a functional option applied to a ComposeRequest before dispatch.
// Use SendOption to attach per-call extras (tags, attachments, reply-to, custom headers)
// without adding per-email-type function signatures.
type SendOption func(*ComposeRequest)

// WithTags appends delivery metadata tags to the ComposeRequest.
func WithTags(tags ...newman.Tag) SendOption {
	return func(r *ComposeRequest) {
		r.Tags = append(r.Tags, tags...)
	}
}

// WithReplyTo sets the reply-to address on the ComposeRequest.
func WithReplyTo(addr string) SendOption {
	return func(r *ComposeRequest) {
		r.ReplyTo = addr
	}
}

// WithAttachments appends dynamic attachments to the ComposeRequest.
func WithAttachments(attachments ...*newman.Attachment) SendOption {
	return func(r *ComposeRequest) {
		r.Attachments = append(r.Attachments, attachments...)
	}
}

// WithHeaders merges custom headers into the ComposeRequest.
func WithHeaders(headers map[string]string) SendOption {
	return func(r *ComposeRequest) {
		if r.Headers == nil {
			r.Headers = make(map[string]string, len(headers))
		}

		maps.Copy(r.Headers, headers)
	}
}

// WithOwnerOnly restricts template lookup to owner-scoped templates, excluding system-owned templates.
func WithOwnerOnly() SendOption {
	return func(r *ComposeRequest) {
		r.OwnerOnly = true
	}
}

// ComposeRequest defines the input required to compose a message from notification/email templates.
type ComposeRequest struct {
	// OwnerID is the organization owner context for owner-scoped template lookup.
	OwnerID string
	// Template identifies the notification template to resolve.
	Template TemplateRef
	// To contains recipient email addresses.
	To []string
	// From is the sender email address.
	From string
	// ReplyTo is an optional reply-to email address.
	ReplyTo string
	// Data contains template rendering variables.
	Data map[string]any
	// Tags are delivery metadata tags.
	Tags []newman.Tag
	// Headers are optional custom email headers.
	Headers map[string]string
	// Attachments are dynamic per-send attachments appended to the composed message.
	// Static attachments associated with the email template record are loaded separately.
	Attachments []*newman.Attachment
	// OwnerOnly restricts template lookup to owner-scoped templates, excluding system-owned templates.
	// Set to true when composing from a user-defined workflow to prevent access to system templates.
	OwnerOnly bool
}
