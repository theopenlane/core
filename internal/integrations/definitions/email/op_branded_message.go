package email

import (
	"encoding/json"

	"github.com/samber/lo"
	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/internal/integrations/providerkit"
)

// BrandedMessageRequest is a customer-selectable catalog entry providing a flexible,
// brand-themed email shape. Customers supply the subject, headline, body paragraphs,
// and optional call-to-action; the base theme handles layout, branding,
// and footer chrome
type BrandedMessageRequest struct {
	// RecipientInfo and CampaignContext are populated per-send (campaign target and
	// dispatch context), not authored in the template, so they are excluded from the
	// reflected catalog config schema via jsonschema:"-". Their JSON tags are untouched
	// so the dispatch payload still carries them
	RecipientInfo   `jsonschema:"-"`
	CampaignContext `jsonschema:"-"`
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
	ButtonLink string `json:"buttonLink,omitempty" jsonschema:"format=uri,description=Call-to-action button URL"`
	// Outros are fine-print paragraphs rendered after the call-to-action
	Outros []string `json:"outros,omitempty" jsonschema:"description=Fine-print paragraphs rendered after the call-to-action"`
	// LogoURL overrides the hero logo displayed in the email body for this send
	LogoURL string `json:"logoURL,omitempty" jsonschema:"format=uri,description=Hero logo URL override for this send"`
	// HeaderLogoURL overrides the small logo/icon in the top header bar for this send
	HeaderLogoURL string `json:"headerLogoURL,omitempty" jsonschema:"format=uri,description=Header bar logo/icon URL override for this send"`
	// PrimaryColor overrides the headline/emphasis color for this send
	PrimaryColor string `json:"primaryColor,omitempty" jsonschema:"format=color,description=Primary headline color override (hex)"`
	// BodyBackgroundColor overrides the outer page background color for this send
	BodyBackgroundColor string `json:"bodyBackgroundColor,omitempty" jsonschema:"format=color,description=Outer page background color override (hex)"`
	// CardBackgroundColor overrides the card container background color for this send
	CardBackgroundColor string `json:"cardBackgroundColor,omitempty" jsonschema:"format=color,description=Card container background color override (hex)"`
	// HeroBackgroundColor overrides the hero banner background color for this send
	HeroBackgroundColor string `json:"heroBackgroundColor,omitempty" jsonschema:"format=color,description=Hero banner background color override (hex)"`
	// ButtonColor overrides the call-to-action button background color for this send
	ButtonColor string `json:"buttonColor,omitempty" jsonschema:"format=color,description=CTA button background color override (hex)"`
	// ButtonTextColor overrides the call-to-action button text color for this send
	ButtonTextColor string `json:"buttonTextColor,omitempty" jsonschema:"format=color,description=CTA button text color override (hex)"`
	// TextColor overrides the body paragraph text color for this send
	TextColor string `json:"textColor,omitempty" jsonschema:"format=color,description=Body text color override (hex)"`
	// FooterTextColor overrides the muted text color for this send
	FooterTextColor string `json:"footerTextColor,omitempty" jsonschema:"format=color,description=Muted text color override (hex)"`
	// CompanyName is the company display name shown in the footer; customer-supplied
	// because a branded message carries the customer's own identity rather than Openlane's
	CompanyName string `json:"companyName,omitempty" jsonschema:"description=Company display name shown in the footer"`
	// Corporation is the legal corporation name used in the footer copyright notice
	Corporation string `json:"corporation,omitempty" jsonschema:"description=Legal corporation name used in the footer copyright notice"`
	// CompanyAddress is the mailing address shown in the footer
	CompanyAddress string `json:"companyAddress,omitempty" jsonschema:"description=Company mailing address shown in the footer"`
	// Copyright overrides the footer copyright line; when empty the template renders a notice from Corporation and the current year
	Copyright string `json:"copyright,omitempty" jsonschema:"description=Footer copyright line; when empty a notice is rendered from the corporation and current year"`
	// Tagline is a short descriptive footer line rendered above the social row
	Tagline string `json:"tagline,omitempty" jsonschema:"description=Short descriptive footer line rendered above the social row"`
	// TermsURL is the terms of service link shown in the footer
	TermsURL string `json:"termsURL,omitempty" jsonschema:"format=uri,description=Terms of service link shown in the footer"`
	// PrivacyURL is the privacy policy link shown in the footer
	PrivacyURL string `json:"privacyURL,omitempty" jsonschema:"format=uri,description=Privacy policy link shown in the footer"`
	// UnsubscribeURL is the unsubscribe link shown in the footer
	UnsubscribeURL string `json:"unsubscribeURL,omitempty" jsonschema:"format=uri,description=Unsubscribe link shown in the footer"`
	// Social is the ordered list of social footer entries rendered beneath the body
	Social []SocialLink `json:"social,omitempty" jsonschema:"description=Social footer links rendered beneath the body"`
}

// brandedMessageSchema is the reflected JSON schema for the branded message input type
// and BrandedMessageOp is the typed operation ref used for catalog dispatch
var (
	brandedMessageSchema, BrandedMessageOp = providerkit.OperationSchema[BrandedMessageRequest]() //nolint:revive
)

// brandedMessageExample is a representative input used to render the catalog preview
// and to seed the form preview with demo values for fields the author has not yet filled.
// Branding fields carry literal demo values because no installation config is guaranteed;
// content fields use {{ .firstName }} to demonstrate per-recipient interpolation
var brandedMessageExample = BrandedMessageRequest{
	RecipientInfo: RecipientInfo{
		Email:     "jordan.avery@example.com",
		FirstName: "Jordan",
		LastName:  "Avery",
	},
	Subject:         "A note from Acme Security",
	Preheader:       "A quick update from the Acme Security team",
	Title:           "Hi {{ .firstName }}, welcome aboard",
	Intros:          []string{"Thanks for partnering with Acme Security — we're glad to have you.", "Review your onboarding checklist below to get started."},
	ButtonText:      "View Onboarding",
	ButtonLink:      "https://example.com/onboarding",
	Outros:          []string{"Questions? Just reply to this email and our team will help out."},
	LogoURL:         "https://www.theopenlane.io/cdn-cgi/imagedelivery/2gi-D0CFOlSOflWJG-LQaA/12e42452-e66e-4bae-0011-45a3f2cb6200/w=240,fit=contain",
	HeaderLogoURL:   "https://www.theopenlane.io/cdn-cgi/imagedelivery/2gi-D0CFOlSOflWJG-LQaA/12e42452-e66e-4bae-0011-45a3f2cb6200/w=240,fit=contain",
	CompanyName:     "Acme Security",
	Corporation:     "Acme Security, Inc.",
	CompanyAddress:  "123 Market Street, San Francisco, CA 94105",
	Tagline:         "Security and compliance, simplified",
	PrimaryColor:    "#0f3d3a",
	ButtonColor:     "#3fc2b4",
	ButtonTextColor: "#0f3d3a",
	Social: []SocialLink{
		{Platform: "GitHub", IconURL: "https://www.theopenlane.io/cdn-cgi/imagedelivery/2gi-D0CFOlSOflWJG-LQaA/39a11dc8-8e01-44ed-8557-0b78ae050a00/w=36", URL: "https://github.com/example"},
		{Platform: "LinkedIn", IconURL: "https://www.theopenlane.io/cdn-cgi/imagedelivery/2gi-D0CFOlSOflWJG-LQaA/e9c20fd9-c8a6-4f79-9267-41f491746c00/w=36", URL: "https://linkedin.com/company/example"},
	},
}

// brandedMessageUISchema carries only UI hints that JSON Schema cannot express. Field
// types and widgets are driven by the reflected config schema instead: color fields carry
// format=color and URL fields carry format=uri, field order follows struct order, and
// per-send fields (recipient, campaign) are excluded from the schema via jsonschema:"-".
// The one irreducible hint left is rendering the body-paragraph lists as multi-line textareas
var brandedMessageUISchema = json.RawMessage(`{
  "intros": {"items": {"ui:widget": "textarea"}},
  "outros": {"items": {"ui:widget": "textarea"}}
}`)

var _ = RegisterEmailOperation(Operation[BrandedMessageRequest]{
	Op:                 BrandedMessageOp,
	Schema:             brandedMessageSchema,
	Theme:              baseTheme,
	Description:        "Customer-authored branded message with headline, body paragraphs, and an optional call-to-action",
	CustomerSelectable: lo.ToPtr(true),
	Example:            brandedMessageExample,
	UISchema:           brandedMessageUISchema,
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
		// A branded message carries the customer's own identity, and the customer
		// may have no installation config to inherit. Identity, logo, and footer
		// fields are therefore sourced entirely from the request. Assigning directly
		// (rather than guarding on non-empty) also strips any Openlane system defaults
		// so they never leak into a customer's message. Functional fields (FromEmail,
		// SupportEmail, QuestionnaireEmail, APIKey, Provider, RootURL, ProductURL,
		// DocsURL, TroubleText) are preserved — they back delivery, not branding.
		cfg.CompanyName = req.CompanyName
		cfg.CompanyAddress = req.CompanyAddress
		cfg.Corporation = req.Corporation
		cfg.Copyright = req.Copyright
		cfg.Tagline = req.Tagline
		cfg.TermsURL = req.TermsURL
		cfg.PrivacyURL = req.PrivacyURL
		cfg.UnsubscribeURL = req.UnsubscribeURL
		cfg.Social = req.Social
		cfg.LogoURL = req.LogoURL
		cfg.HeaderLogoURL = req.HeaderLogoURL
		cfg.HeaderText = ""

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
