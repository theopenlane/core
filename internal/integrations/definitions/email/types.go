package email

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// emailUserInputSchema is the JSON schema for customer-provided email branding configuration
	emailUserInputSchema = providerkit.SchemaFrom[EmailUserInput]()
	// DefinitionID is the stable identifier for the email integration definition
	DefinitionID = types.NewDefinitionRef("def_01EMAILINT00000000000000001")
	// runtimeEmailSchema is the JSON schema and typed ref for the runtime email config
	runtimeEmailSchema, runtimeEmailRef = providerkit.RuntimeIntegrationSchema[RuntimeEmailConfig]()
	// emailCredentialSchema is the JSON schema and typed credential ref for customer-provisioned email
	emailCredentialSchema, emailCredentialRef = providerkit.CredentialSchema[EmailCredential]()
	// emailClientRef is the client ref for the email client used by this definition
	emailClientRef = types.NewClientRef[*EmailClient]()
	// healthCheckSchema is the operation schema for the health check operation
	healthCheckSchema, healthCheckOp = providerkit.OperationSchema[HealthCheck]()
	// sendEmailSchema is the operation schema for the generic send-email operation
	sendEmailSchema, SendEmailOp = providerkit.OperationSchema[SendEmailRequest]()
	// sendCampaignSchema is the operation schema for the send-campaign operation
	sendCampaignSchema, SendCampaignOp = providerkit.OperationSchema[SendCampaignRequest]()
)

// Tag key constants for email delivery tracking via provider webhooks
const (
	// TagCampaignTargetID identifies the campaign target for delivery event correlation
	TagCampaignTargetID = "campaign_target_id"
	// TagAssessmentResponseID identifies the assessment response for delivery event correlation
	TagAssessmentResponseID = "assessment_response_id"
	// TagIsTest marks a message as a test send to prevent cascade updates
	TagIsTest = "is_test"
)

// RuntimeEmailConfig is the complete config for runtime-provisioned email.
// Sourced from koanf/environment at startup
type RuntimeEmailConfig struct {
	// APIKey is the email provider API key
	APIKey string `json:"apiKey" koanf:"apiKey"`
	// Provider is the email service provider name (resend, sendgrid, postmark)
	Provider string `json:"provider" koanf:"provider" jsonschema:"required,enum=resend,enum=sendgrid,enum=postmark,description=Email service provider" default:"resend"`
	// FromEmail is the default sender email address
	FromEmail string `json:"fromEmail" koanf:"fromEmail" default:"support@mail.theopenlane.io"`
	// CompanyName is the display name of the sending company
	CompanyName string `json:"companyName" koanf:"companyName" default:"Openlane"`
	// CompanyAddress is the mailing address of the company
	CompanyAddress string `json:"companyAddress" koanf:"companyAddress" default:"5150 Broadway St San Antonio, TX 78209"`
	// Corporation is the legal corporation name
	Corporation string `json:"corporation" koanf:"corporation" default:"theopenlane, Inc."`
	// SupportEmail is the support contact email address
	SupportEmail string `json:"supportEmail" koanf:"supportEmail" default:"support@theopenlane.io"`
	// LogoURL is the company logo image URL
	LogoURL string `json:"logoURL" koanf:"logoURL" default:"https://www.theopenlane.io/cdn-cgi/imagedelivery/2gi-D0CFOlSOflWJG-LQaA/12e42452-e66e-4bae-0011-45a3f2cb6200/public"`
	// RootURL is the root application URL used to construct email action links
	RootURL string `json:"rootURL" koanf:"rootURL" default:"https://www.theopenlane.io"`
	// ProductURL is the product home URL
	ProductURL string `json:"productURL" koanf:"productURL" default:"https://console.theopenlane.io"`
	// DocsURL is the documentation URL
	DocsURL string `json:"docsURL" koanf:"docsURL" default:"https://docs.theopenlane.io"`
	// QuestionnaireEmail is an optional sender override for questionnaire auth emails
	QuestionnaireEmail string `json:"questionnaireEmail,omitempty" koanf:"questionnaireEmail" default:"no-reply@mail.theopenlane.io"`
	// Copyright is the copyright notice for email footers
	Copyright string `json:"copyright,omitempty" koanf:"copyright" default:"© theopenlane, Inc. All rights reserved."`
	// TroubleText is the fallback help text shown below action buttons; {ACTION} is replaced with button text at render time
	TroubleText string `json:"troubleText,omitempty" koanf:"troubleText" default:"If you're having trouble with the button '{ACTION}', copy and paste the URL below into your web browser"`
	// UnsubscribeURL is the unsubscribe link for email footers
	UnsubscribeURL string `json:"unsubscribeURL,omitempty" koanf:"unsubscribeURL" default:"https://console.theopenlane.io/unsubscribe"`
}

// Provisioned reports whether the runtime config has the minimum required fields
// to build a working email client (API key, provider, and from address)
func (c RuntimeEmailConfig) Provisioned() bool {
	return c.APIKey != "" && c.Provider != "" && c.FromEmail != ""
}

// EmailCredential is the credential schema for customer-provisioned email
type EmailCredential struct {
	// APIKey is the email provider API key
	APIKey string `json:"apiKey" jsonschema:"required,description=Email provider API key"`
	// Provider is the email service provider name
	Provider string `json:"provider" jsonschema:"required,enum=resend,enum=sendgrid,enum=postmark,description=Email service provider"`
}

// EmailUserInput is the installation-scoped configuration that customers provide
// when setting up their own email integration. These fields supply the branding
// and sender identity that the EmailClient carries for all customer-initiated sends
type EmailUserInput struct {
	// FromEmail is the default sender email address
	FromEmail string `json:"fromEmail" jsonschema:"required,description=Default sender email address"`
	// CompanyName is the display name used in email templates
	CompanyName string `json:"companyName" jsonschema:"required,description=Company display name for email templates"`
	// CompanyAddress is the mailing address for CAN-SPAM compliance
	CompanyAddress string `json:"companyAddress,omitempty" jsonschema:"description=Company mailing address"`
	// Corporation is the legal corporation name
	Corporation string `json:"corporation,omitempty" jsonschema:"description=Legal corporation name"`
	// SupportEmail is the support contact email address
	SupportEmail string `json:"supportEmail,omitempty" jsonschema:"description=Support contact email address"`
	// LogoURL is the company logo URL for email templates
	LogoURL string `json:"logoURL,omitempty" jsonschema:"description=Company logo URL for email templates"`
	// RootURL is the root application URL used to construct email action links
	RootURL string `json:"rootURL,omitempty" jsonschema:"description=Root application URL"`
	// ProductURL is the product home URL
	ProductURL string `json:"productURL,omitempty" jsonschema:"description=Product home URL"`
	// DocsURL is the documentation URL
	DocsURL string `json:"docsURL,omitempty" jsonschema:"description=Documentation URL"`
	// Copyright is the copyright notice for email footers
	Copyright string `json:"copyright,omitempty" jsonschema:"description=Copyright notice for email footers; auto-generated from corporation and year when empty"`
	// TroubleText is the fallback help text shown below action buttons
	TroubleText string `json:"troubleText,omitempty" jsonschema:"description=Help text shown below action buttons; {ACTION} is replaced with button text"`
	// UnsubscribeURL is the unsubscribe link for email footers
	UnsubscribeURL string `json:"unsubscribeURL,omitempty" jsonschema:"description=Unsubscribe URL for email footers; auto-generated from product URL when empty"`
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
		Copyright:      u.Copyright,
		TroubleText:    u.TroubleText,
		UnsubscribeURL: u.UnsubscribeURL,
	}
}
