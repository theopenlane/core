package email

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

var (
	// DefinitionID is the stable identifier for the email integration definition
	DefinitionID = types.NewDefinitionRef("def_01EMAILINT00000000000000001")
	// runtimeEmailSchema is the JSON schema and typed ref for the runtime email config
	runtimeEmailSchema, runtimeEmailRef = providerkit.RuntimeSchema[RuntimeEmailConfig]()
	// emailCredentialSchema is the JSON schema and typed credential ref for customer-provisioned email
	emailCredentialSchema, emailCredentialRef = providerkit.CredentialSchema[Credential]()
	// emailClientRef is the client ref for the email client used by this definition
	emailClientRef = types.NewClientRef[*Client]()
	// healthCheckSchema is the operation schema for the health check operation
	healthCheckSchema, healthCheckOp = providerkit.OperationSchema[HealthCheck]()
	// sendEmailSchema is the operation schema for the generic send-email operation
	sendEmailSchema, SendEmailOp = providerkit.OperationSchema[SendEmailRequest]() //nolint:revive
	// sendBrandedCampaignSchema is the operation schema for the branded campaign dispatch operation
	sendBrandedCampaignSchema, SendCampaignOp = providerkit.OperationSchema[SendBrandedCampaignRequest]() //nolint:revive
	// sendQuestionnaireCampaignSchema is the operation schema for the questionnaire campaign dispatch operation
	sendQuestionnaireCampaignSchema, SendQuestionnaireCampaignOp = providerkit.OperationSchema[SendQuestionnaireCampaignRequest]() //nolint:revive
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
// Operational fields (API key, provider, URLs) are sourced from koanf/environment
// at startup. Branding and presentation fields carry struct-tag defaults only and
// are overridable per-send via UserInput or per-operation Config functions
type RuntimeEmailConfig struct {
	// APIKey is the email provider API key
	APIKey string `json:"apiKey" koanf:"apiKey" jsonschema:"required,description=Email provider API key"`
	// Provider is the email service provider name (resend, sendgrid, postmark)
	Provider string `json:"provider" koanf:"provider" jsonschema:"required,enum=resend,enum=sendgrid,enum=postmark,description=Email service provider" default:"resend"`
	// FromEmail is the default sender email address
	FromEmail string `json:"fromEmail" koanf:"fromEmail" jsonschema:"required,description=Sender email address" default:"support@mail.theopenlane.io"`
	// SupportEmail is the support contact email address
	SupportEmail string `json:"supportEmail" koanf:"supportEmail" jsonschema:"description=Support contact email address" default:"support@theopenlane.io"`
	// QuestionnaireEmail is an optional sender override for questionnaire auth emails
	QuestionnaireEmail string `json:"questionnaireEmail,omitempty" koanf:"questionnaireEmail" jsonschema:"description=Sender override for questionnaire auth emails" default:"support@mail.theopenlane.io"`
	// RootURL is the root application URL used to construct email action links
	RootURL string `json:"rootURL" koanf:"rootURL" jsonschema:"description=Root application URL used to construct email action links" default:"https://www.theopenlane.io"`
	// ProductURL is the product home URL
	ProductURL string `json:"productURL" koanf:"productURL" jsonschema:"description=Product home URL" default:"https://console.theopenlane.io"`
	// DocsURL is the documentation URL
	DocsURL string `json:"docsURL" koanf:"docsURL" jsonschema:"description=Documentation URL" default:"https://docs.theopenlane.io"`
	// CompanyName is the display name of the sending company
	CompanyName string `json:"companyName" jsonschema:"description=Company display name" default:"Openlane"`
	// CompanyAddress is the mailing address of the company
	CompanyAddress string `json:"companyAddress" jsonschema:"description=Company mailing address" default:"5150 Broadway St San Antonio, TX 78209"`
	// Corporation is the legal corporation name
	Corporation string `json:"corporation" jsonschema:"description=Legal corporation name" default:"theopenlane, Inc."`
	// LogoURL is the hero logo image URL displayed prominently in the email body
	LogoURL string `json:"logoURL,omitempty" jsonschema:"description=Hero logo URL displayed in the email body"`
	// HeaderLogoURL is the small logo/icon displayed in the top header bar
	HeaderLogoURL string `json:"headerLogoURL,omitempty" jsonschema:"description=Small logo or icon displayed in the top header bar" default:"https://www.theopenlane.io/cdn-cgi/imagedelivery/2gi-D0CFOlSOflWJG-LQaA/12e42452-e66e-4bae-0011-45a3f2cb6200/public"`
	// Copyright is an optional copyright override for email footers; when empty the template renders © {year} {corporation}
	Copyright string `json:"copyright,omitempty" jsonschema:"description=Copyright override for email footers; when empty the template renders a dynamic notice from Corporation and the current year"`
	// TroubleText is the fallback help text shown below action buttons; {ACTION} is replaced with button text at render time
	TroubleText string `json:"troubleText,omitempty" jsonschema:"description=Fallback help text shown below action buttons; {ACTION} is replaced with the button text at render time" default:"If you're having trouble with the button '{ACTION}', copy and paste the URL below into your web browser"`
	// TermsURL is the terms of service link for email footers
	TermsURL string `json:"termsURL,omitempty" jsonschema:"description=Terms of service link for email footers" default:"https://www.theopenlane.io/legal/terms-of-service"`
	// PrivacyURL is the privacy policy link for email footers
	PrivacyURL string `json:"privacyURL,omitempty" jsonschema:"description=Privacy policy link for email footers" default:"https://www.theopenlane.io/legal/privacy"`
	// UnsubscribeURL is an optional unsubscribe link override for email footers; when empty the template constructs one from ProductURL and the recipient email
	UnsubscribeURL string `json:"unsubscribeURL,omitempty" jsonschema:"description=Unsubscribe link override for email footers; when empty the template constructs one from ProductURL and the recipient email"`
	// HeaderText is the optional text displayed in the upper-right corner of the modern theme header row
	HeaderText string `json:"headerText,omitempty" jsonschema:"description=Text displayed in the upper-right corner of the modern theme header"`
	// CardStyle controls the card visual style; elevated adds rounded corners and a drop shadow
	CardStyle string `json:"cardStyle,omitempty" jsonschema:"enum=flat,enum=elevated,description=Card visual style" default:"elevated"`
	// BodyBackgroundColor is the outer page background color
	BodyBackgroundColor string `json:"bodyBackgroundColor,omitempty" jsonschema:"description=Outer page background color" default:"#e8eaed"`
	// CardBackgroundColor is the card container background color
	CardBackgroundColor string `json:"cardBackgroundColor,omitempty" jsonschema:"description=Card container background color" default:"#ffffff"`
	// HeroBackgroundColor is the hero banner section background color
	HeroBackgroundColor string `json:"heroBackgroundColor,omitempty" jsonschema:"description=Hero banner section background color" default:"#f3f4f6"`
	// ButtonColor is the call-to-action button background color
	ButtonColor string `json:"buttonColor,omitempty" jsonschema:"description=Call-to-action button background color" default:"#14171e"`
	// ButtonTextColor is the call-to-action button text color
	ButtonTextColor string `json:"buttonTextColor,omitempty" jsonschema:"description=Call-to-action button text color" default:"#ffffff"`
	// HeadingColor is the heading and title text color
	HeadingColor string `json:"headingColor,omitempty" jsonschema:"description=Heading and title text color" default:"#14171e"`
	// TextColor is the body paragraph text color
	TextColor string `json:"textColor,omitempty" jsonschema:"description=Body paragraph text color" default:"#43454b"`
	// FooterTextColor is the muted text color for headers, footers, and secondary content
	FooterTextColor string `json:"footerTextColor,omitempty" jsonschema:"description=Muted text color for headers footers and secondary content" default:"#7b7d81"`
	// Tagline is a short descriptive footer line rendered in modern themes above the social row
	Tagline string `json:"tagline,omitempty" jsonschema:"description=Short descriptive footer line rendered above the social row in modern themes"`
	// Social is the ordered list of social footer entries rendered by modern themes
	Social []SocialLink `json:"social,omitempty" jsonschema:"-"`
}

// SocialLink is a single social media footer entry: platform label, icon image URL, and destination URL
type SocialLink struct {
	// Platform is the display label for the social network (e.g. X, LinkedIn)
	Platform string `json:"platform" koanf:"platform" jsonschema:"required,description=Display label for the social network"`
	// IconURL is the publicly reachable URL of the icon image
	IconURL string `json:"iconURL" koanf:"iconURL" jsonschema:"required,description=Publicly reachable icon image URL"`
	// URL is the destination the icon links to
	URL string `json:"url" koanf:"url" jsonschema:"required,description=Destination URL the icon links to"`
}

// Provisioned reports whether the runtime config has the minimum required fields
// to build a working email client (API key, provider, and from address)
func (c RuntimeEmailConfig) Provisioned() bool {
	return c.APIKey != "" && c.Provider != "" && c.FromEmail != ""
}

// Credential is the credential schema for customer-provisioned email
type Credential struct {
	// APIKey is the email provider API key
	APIKey string `json:"apiKey" jsonschema:"required,description=Email provider API key"`
	// Provider is the email service provider name
	Provider string `json:"provider" jsonschema:"required,enum=resend,enum=sendgrid,enum=postmark,description=Email service provider"`
}

// UserInput is the installation-scoped configuration that customers provide
// when setting up their own email integration. These fields supply the branding
// and sender identity that the Client carries for all customer-initiated sends
type UserInput struct {
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
	// LogoURL is the hero logo URL displayed prominently in the email body
	LogoURL string `json:"logoURL,omitempty" jsonschema:"description=Hero logo URL displayed in the email body"`
	// HeaderLogoURL is the small logo/icon displayed in the top header bar
	HeaderLogoURL string `json:"headerLogoURL,omitempty" jsonschema:"description=Small logo or icon displayed in the top header bar"`
	// RootURL is the root application URL used to construct email action links
	RootURL string `json:"rootURL,omitempty" jsonschema:"description=Root application URL"`
	// ProductURL is the product home URL
	ProductURL string `json:"productURL,omitempty" jsonschema:"description=Product home URL"`
	// DocsURL is the documentation URL
	DocsURL string `json:"docsURL,omitempty" jsonschema:"description=Documentation URL"`
	// Copyright is the copyright notice for email footers
	Copyright string `json:"copyright,omitempty" jsonschema:"description=Copyright notice for email footers; auto-generated from corporation and year when empty"`
	// TroubleText is the fallback help text shown below action buttons
	TroubleText string `json:"troubleText,omitempty" jsonschema:"description=Help text shown below action buttons"`
	// UnsubscribeURL is the unsubscribe link for email footers
	UnsubscribeURL string `json:"unsubscribeURL,omitempty" jsonschema:"description=Unsubscribe URL for email footers"`
	// Tagline is a short descriptive footer line rendered in modern themes above the social row
	Tagline string `json:"tagline,omitempty" jsonschema:"description=Short descriptive footer line rendered above the social row in modern themes"`
	// Social is the ordered list of social footer entries rendered by modern themes
	Social []SocialLink `json:"social,omitempty" jsonschema:"-"`
	// CardStyle controls the card visual style; elevated adds rounded corners and a drop shadow
	CardStyle string `json:"cardStyle,omitempty" jsonschema:"enum=flat,enum=elevated,description=Card visual style"`
	// BodyBackgroundColor is the outer page background color
	BodyBackgroundColor string `json:"bodyBackgroundColor,omitempty" jsonschema:"description=Outer page background color"`
	// CardBackgroundColor is the card container background color
	CardBackgroundColor string `json:"cardBackgroundColor,omitempty" jsonschema:"description=Card container background color"`
	// HeroBackgroundColor is the hero banner section background color
	HeroBackgroundColor string `json:"heroBackgroundColor,omitempty" jsonschema:"description=Hero banner section background color"`
	// ButtonColor is the call-to-action button background color
	ButtonColor string `json:"buttonColor,omitempty" jsonschema:"description=Call-to-action button background color"`
	// ButtonTextColor is the call-to-action button text color
	ButtonTextColor string `json:"buttonTextColor,omitempty" jsonschema:"description=Call-to-action button text color"`
	// HeadingColor is the heading and title text color
	HeadingColor string `json:"headingColor,omitempty" jsonschema:"description=Heading and title text color"`
	// TextColor is the body paragraph text color
	TextColor string `json:"textColor,omitempty" jsonschema:"description=Body paragraph text color"`
	// FooterTextColor is the muted text color for headers, footers, and secondary content
	FooterTextColor string `json:"footerTextColor,omitempty" jsonschema:"description=Muted text color for headers footers and secondary content"`
}

// ToRuntimeConfig converts customer user input to a RuntimeEmailConfig
func (u UserInput) ToRuntimeConfig() RuntimeEmailConfig {
	var cfg RuntimeEmailConfig

	_ = jsonx.RoundTrip(u, &cfg)

	return cfg
}
