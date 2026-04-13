package email

import (
	"github.com/theopenlane/newman"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// DefinitionID returns the stable identifier for the email integration definition
func DefinitionID() string {
	return definitionID.ID()
}

// CustomerClientID returns the client identity for customer-provisioned email clients
func CustomerClientID() types.ClientID {
	return emailClientRef.ID()
}

var (
	// emailUserInputSchema is the JSON schema for customer-provided email branding configuration
	emailUserInputSchema = providerkit.SchemaFrom[EmailUserInput]()
	// definitionID is the stable identifier for the email integration definition
	definitionID = types.NewDefinitionRef("def_01EMAILINT00000000000000001")
	// runtimeEmailSchema is the JSON schema and typed ref for the runtime email config
	runtimeEmailSchema, runtimeEmailRef = providerkit.RuntimeIntegrationSchema[RuntimeEmailConfig]()
	// emailCredentialSchema is the JSON schema and typed credential ref for customer-provisioned email
	emailCredentialSchema, emailCredentialRef = providerkit.CredentialSchema[EmailCredential]()
	// emailClientRef is the client ref for the email client used by this definition
	emailClientRef = types.NewClientRef[*EmailClient]()
	// sendCampaignSchema is the operation schema for the send-campaign operation
	sendCampaignSchema, sendCampaignOp = providerkit.OperationSchema[SendCampaignRequest]()
)

// RuntimeEmailConfig is the complete config for runtime-provisioned email.
// Sourced from koanf/environment at startup
type RuntimeEmailConfig struct {
	// APIKey is the email provider API key
	APIKey string `json:"api_key" koanf:"apiKey"`
	// Provider is the email service provider name (resend, sendgrid, postmark)
	Provider string `json:"provider" koanf:"provider" jsonschema:"required,enum=resend,enum=sendgrid,enum=postmark,description=Email service provider"`
	// FromEmail is the default sender email address
	FromEmail string `json:"from_email" koanf:"fromEmail"`
	// CompanyName is the display name of the sending company
	CompanyName string `json:"company_name" koanf:"companyName"`
	// CompanyAddress is the mailing address of the company
	CompanyAddress string `json:"company_address" koanf:"companyAddress"`
	// Corporation is the legal corporation name
	Corporation string `json:"corporation" koanf:"corporation"`
	// SupportEmail is the support contact email address
	SupportEmail string `json:"support_email" koanf:"supportEmail"`
	// LogoURL is the company logo image URL
	LogoURL string `json:"logo_url" koanf:"logoURL"`
	// RootURL is the root application URL used to construct email action links
	RootURL string `json:"root_url" koanf:"rootURL"`
	// ProductURL is the product home URL
	ProductURL string `json:"product_url" koanf:"productURL"`
	// DocsURL is the documentation URL
	DocsURL string `json:"docs_url" koanf:"docsURL"`
}

// Provisioned reports whether the runtime config has the minimum required fields
// to build a working email client (API key, provider, and from address)
func (c RuntimeEmailConfig) Provisioned() bool {
	return c.APIKey != "" && c.Provider != "" && c.FromEmail != ""
}

// EmailCredential is the credential schema for customer-provisioned email
type EmailCredential struct {
	// APIKey is the email provider API key
	APIKey string `json:"api_key" jsonschema:"required,description=Email provider API key"`
	// Provider is the email service provider name
	Provider string `json:"provider" jsonschema:"required,enum=resend,enum=sendgrid,enum=postmark,description=Email service provider"`
}

// EmailUserInput is the installation-scoped configuration that customers provide
// when setting up their own email integration. These fields supply the branding
// and sender identity that the EmailClient carries for all customer-initiated sends
type EmailUserInput struct {
	// FromEmail is the default sender email address
	FromEmail string `json:"from_email" jsonschema:"required,description=Default sender email address"`
	// CompanyName is the display name used in email templates
	CompanyName string `json:"company_name" jsonschema:"required,description=Company display name for email templates"`
	// CompanyAddress is the mailing address for CAN-SPAM compliance
	CompanyAddress string `json:"company_address,omitempty" jsonschema:"description=Company mailing address"`
	// Corporation is the legal corporation name
	Corporation string `json:"corporation,omitempty" jsonschema:"description=Legal corporation name"`
	// SupportEmail is the support contact email address
	SupportEmail string `json:"support_email,omitempty" jsonschema:"description=Support contact email address"`
	// LogoURL is the company logo URL for email templates
	LogoURL string `json:"logo_url,omitempty" jsonschema:"description=Company logo URL for email templates"`
	// RootURL is the root application URL used to construct email action links
	RootURL string `json:"root_url,omitempty" jsonschema:"description=Root application URL"`
	// ProductURL is the product home URL
	ProductURL string `json:"product_url,omitempty" jsonschema:"description=Product home URL"`
	// DocsURL is the documentation URL
	DocsURL string `json:"docs_url,omitempty" jsonschema:"description=Documentation URL"`
}

// ToRuntimeConfig converts customer user input to a RuntimeEmailConfig for rendering.
// Credential fields (APIKey, Provider) are not set — those come from the credential ref
func (u EmailUserInput) ToRuntimeConfig() RuntimeEmailConfig {
	return RuntimeEmailConfig{
		FromEmail:      u.FromEmail,
		CompanyName:    u.CompanyName,
		CompanyAddress: u.CompanyAddress,
		Corporation:    u.Corporation,
		SupportEmail:   u.SupportEmail,
		LogoURL:        u.LogoURL,
		RootURL:        u.RootURL,
		ProductURL:     u.ProductURL,
		DocsURL:        u.DocsURL,
	}
}

// SendCampaignRequest is the operation config for dispatching a full campaign
type SendCampaignRequest struct {
	// CampaignID is the identifier of the campaign to dispatch
	CampaignID string `json:"campaign_id" jsonschema:"required"`
}

// TemplateRef selects a notification template by stable key or explicit record ID
type TemplateRef struct {
	// ID references a notification template by database ID
	ID string
	// Key references a notification template by stable key
	Key string
}

// Validate checks that the reference contains exactly one selector
func (r TemplateRef) Validate() error {
	hasID := r.ID != ""
	hasKey := r.Key != ""

	switch {
	case hasID && hasKey:
		return ErrTemplateReferenceConflict
	case !hasID && !hasKey:
		return ErrMissingTemplateReference
	default:
		return nil
	}
}

// ComposeRequest defines the input required to compose a message from notification/email templates
type ComposeRequest struct {
	// OwnerID is the organization owner context for owner-scoped template lookup
	OwnerID string
	// Template identifies the notification template to resolve
	Template TemplateRef
	// To contains recipient email addresses
	To []string
	// From is the sender email address
	From string
	// ReplyTo is an optional reply-to email address
	ReplyTo string
	// Data contains template rendering variables
	Data map[string]any
	// Tags are delivery metadata tags
	Tags []newman.Tag
	// Headers are optional custom email headers
	Headers map[string]string
	// Attachments are dynamic per-send attachments appended to the composed message
	Attachments []*newman.Attachment
	// OwnerOnly restricts template lookup to owner-scoped templates, excluding system-owned templates
	OwnerOnly bool
}
