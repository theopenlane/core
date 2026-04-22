package email

import (
	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/internal/integrations/definitions/email/themes"
	"github.com/theopenlane/core/internal/integrations/providerkit"
)

// BrandedMessageRequest is a customer-selectable catalog entry providing a flexible,
// brand-themed email shape. Customers supply the subject, headline, body paragraphs,
// and optional call-to-action; the modern-message theme handles layout, branding,
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
	// LogoURL overrides the tenant logo shown in the header for this send
	LogoURL string `json:"logoURL,omitempty" jsonschema:"description=Logo image URL override for this send"`
	// PrimaryColor overrides the headline/emphasis color for this send
	PrimaryColor string `json:"primaryColor,omitempty" jsonschema:"description=Primary headline color override (hex)"`
}

// brandedMessageSchema is the reflected JSON schema for the branded message input type
// and BrandedMessageOp is the typed operation ref used for catalog dispatch
var (
	brandedMessageSchema, BrandedMessageOp = providerkit.OperationSchema[BrandedMessageRequest]()
)

// brandedMessageEmail is the customer-selectable catalog entry for flexible branded messages
var brandedMessageEmail = EmailOperation[BrandedMessageRequest]{
	Op:                 BrandedMessageOp,
	Schema:             brandedMessageSchema,
	Theme:              themes.ModernMessage,
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
			Style:     render.Style{PrimaryColor: req.PrimaryColor},
		}

		if req.ButtonText != "" && req.ButtonLink != "" {
			body.Actions = []render.Action{{
				Button: render.Button{Text: req.ButtonText, Link: req.ButtonLink},
			}}
		}

		return body
	},
	Config: func(cfg RuntimeEmailConfig, req BrandedMessageRequest) RuntimeEmailConfig {
		if req.LogoURL != "" {
			cfg.LogoURL = req.LogoURL
		}

		return cfg
	},
}
