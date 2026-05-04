package email

import (
	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/internal/integrations/providerkit"
)

// BrandedMessageRequest is a customer-selectable catalog entry providing a flexible,
// brand-themed email shape. Customers supply the subject, headline, body paragraphs,
// and optional call-to-action; the base theme handles layout, branding,
// and footer chrome
type BrandedMessageRequest struct {
	RecipientInfo
	CampaignContext
	// Subject is the email subject line
	Subject string `json:"subject" jsonschema:"required,description=Email subject line"`
	// Preheader is hidden preview text shown in the inbox list
	Preheader string `json:"preheader,omitempty" jsonschema:"description=Inbox preview text"`
	// Title is the headline shown above the body paragraphs
	Title string `json:"title" jsonschema:"required,description=Email headline"`
	// Intros are body paragraphs rendered before the optional call-to-action button
	Intros []string `json:"intros,omitempty" jsonschema:"description=Body paragraphs rendered before the call-to-action"`
	// ButtonText is the optional call-to-action button label
	ButtonText string `json:"buttonText,omitempty" jsonschema:"description=Call-to-action button label"`
	// ButtonLink is the URL the call-to-action button navigates to
	ButtonLink string `json:"buttonLink,omitempty" jsonschema:"description=Call-to-action button URL"`
	// Outros are fine-print paragraphs rendered after the call-to-action
	Outros []string `json:"outros,omitempty" jsonschema:"description=Fine-print paragraphs rendered after the call-to-action"`
	// LogoURL overrides the hero logo displayed in the email body for this send
	LogoURL string `json:"logoURL,omitempty" jsonschema:"description=Hero logo URL override for this send"`
	// HeaderLogoURL overrides the small logo/icon in the top header bar for this send
	HeaderLogoURL string `json:"headerLogoURL,omitempty" jsonschema:"description=Header bar logo/icon URL override for this send"`
	// PrimaryColor overrides the headline/emphasis color for this send
	PrimaryColor string `json:"primaryColor,omitempty" jsonschema:"description=Primary headline color override (hex)"`
	// BodyBackgroundColor overrides the outer page background color for this send
	BodyBackgroundColor string `json:"bodyBackgroundColor,omitempty" jsonschema:"description=Outer page background color override (hex)"`
	// CardBackgroundColor overrides the card container background color for this send
	CardBackgroundColor string `json:"cardBackgroundColor,omitempty" jsonschema:"description=Card container background color override (hex)"`
	// HeroBackgroundColor overrides the hero banner background color for this send
	HeroBackgroundColor string `json:"heroBackgroundColor,omitempty" jsonschema:"description=Hero banner background color override (hex)"`
	// ButtonColor overrides the call-to-action button background color for this send
	ButtonColor string `json:"buttonColor,omitempty" jsonschema:"description=CTA button background color override (hex)"`
	// ButtonTextColor overrides the call-to-action button text color for this send
	ButtonTextColor string `json:"buttonTextColor,omitempty" jsonschema:"description=CTA button text color override (hex)"`
	// TextColor overrides the body paragraph text color for this send
	TextColor string `json:"textColor,omitempty" jsonschema:"description=Body text color override (hex)"`
	// FooterTextColor overrides the muted text color for this send
	FooterTextColor string `json:"footerTextColor,omitempty" jsonschema:"description=Muted text color override (hex)"`
}

// brandedMessageSchema is the reflected JSON schema for the branded message input type
// and BrandedMessageOp is the typed operation ref used for catalog dispatch
var (
	brandedMessageSchema, BrandedMessageOp = providerkit.OperationSchema[BrandedMessageRequest]() //nolint:revive
)

var _ = RegisterEmailOperation(Operation[BrandedMessageRequest]{
	Op:                 BrandedMessageOp,
	Schema:             brandedMessageSchema,
	Theme:              baseTheme,
	Description:        "Customer-authored branded message with headline, body paragraphs, and an optional call-to-action",
	CustomerSelectable: true,
	Subject: func(_ RuntimeEmailConfig, req BrandedMessageRequest) string {
		return req.Subject
	},
	Build: func(_ RuntimeEmailConfig, req BrandedMessageRequest) render.ContentBody {
		body := render.ContentBody{
			Preheader: req.Preheader,
			Name:      req.FirstName,
			Title:     req.Title,
			Intros:    render.IntrosBlock{Paragraphs: req.Intros},
			Outros:    render.OutrosBlock{Paragraphs: req.Outros},
		}

		if req.ButtonText != "" && req.ButtonLink != "" {
			body.Actions = []render.Action{{
				Button: render.Button{Text: req.ButtonText, Link: req.ButtonLink},
			}}
		}

		return body
	},
	Config: func(cfg RuntimeEmailConfig, req BrandedMessageRequest) RuntimeEmailConfig {
		// Zero display/branding fields so customer campaigns render a blank
		// canvas rather than inheriting system defaults. Functional fields
		// (FromEmail, SupportEmail, QuestionnaireEmail, APIKey, Provider,
		// RootURL, ProductURL, DocsURL, TroubleText) are preserved — they
		// come from the customer's own integration config.
		cfg.CompanyName = ""
		cfg.CompanyAddress = ""
		cfg.Corporation = ""
		cfg.LogoURL = ""
		cfg.HeaderLogoURL = ""
		cfg.Copyright = ""
		cfg.Tagline = ""
		cfg.Social = nil
		cfg.TermsURL = ""
		cfg.PrivacyURL = ""
		cfg.UnsubscribeURL = ""
		cfg.HeaderText = ""

		if req.LogoURL != "" {
			cfg.LogoURL = req.LogoURL
		}

		if req.HeaderLogoURL != "" {
			cfg.HeaderLogoURL = req.HeaderLogoURL
		}

		if req.PrimaryColor != "" {
			cfg.HeadingColor = req.PrimaryColor
		}

		if req.BodyBackgroundColor != "" {
			cfg.BodyBackgroundColor = req.BodyBackgroundColor
		}

		if req.CardBackgroundColor != "" {
			cfg.CardBackgroundColor = req.CardBackgroundColor
		}

		if req.HeroBackgroundColor != "" {
			cfg.HeroBackgroundColor = req.HeroBackgroundColor
		}

		if req.ButtonColor != "" {
			cfg.ButtonColor = req.ButtonColor
		}

		if req.ButtonTextColor != "" {
			cfg.ButtonTextColor = req.ButtonTextColor
		}

		if req.TextColor != "" {
			cfg.TextColor = req.TextColor
		}

		if req.FooterTextColor != "" {
			cfg.FooterTextColor = req.FooterTextColor
		}

		return cfg
	},
})
