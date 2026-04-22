package email

import (
	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/internal/integrations/definitions/email/themes"
	"github.com/theopenlane/core/internal/integrations/providerkit"
)

// BrandedHeroRequest is a customer-selectable catalog entry that renders the modern-hero
// theme: a full-width header image, large headline, body paragraphs, and an optional
// call-to-action. Suitable for launches, announcements, and welcome-shaped messages
type BrandedHeroRequest struct {
	RecipientInfo
	CampaignContext
	// Subject is the email subject line
	Subject string `json:"subject" jsonschema:"required,description=Email subject line"`
	// Preheader is hidden preview text shown in the inbox list
	Preheader string `json:"preheader,omitempty" jsonschema:"description=Inbox preview text"`
	// HeroImageURL is the URL of the full-width hero image rendered at the top of the panel
	HeroImageURL string `json:"heroImageURL,omitempty" jsonschema:"description=Full-width hero image URL"`
	// HeroImageAlt is the accessible alt text for the hero image
	HeroImageAlt string `json:"heroImageAlt,omitempty" jsonschema:"description=Hero image alt text"`
	// Title is the hero headline shown below the image
	Title string `json:"title" jsonschema:"required,description=Hero headline"`
	// Intros are body paragraphs rendered below the headline
	Intros []string `json:"intros,omitempty" jsonschema:"description=Body paragraphs rendered below the headline"`
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

// brandedHeroSchema is the reflected JSON schema for the hero message input type
// and BrandedHeroOp is the typed operation ref used for catalog dispatch
var (
	brandedHeroSchema, BrandedHeroOp = providerkit.OperationSchema[BrandedHeroRequest]()
)

// brandedHeroEmail is the customer-selectable catalog entry for hero-forward branded messages
var brandedHeroEmail = EmailOperation[BrandedHeroRequest]{
	Op:                 BrandedHeroOp,
	Schema:             brandedHeroSchema,
	Theme:              themes.ModernHero,
	Description:        "Customer-authored hero-image-forward message with headline, body paragraphs, and an optional call-to-action",
	CustomerSelectable: true,
	Subject: func(_ RuntimeEmailConfig, req BrandedHeroRequest) string {
		return req.Subject
	},
	Build: func(_ RuntimeEmailConfig, req BrandedHeroRequest) render.ContentBody {
		body := render.ContentBody{
			Preheader: req.Preheader,
			Name:      req.FirstName,
			Title:     req.Title,
			Intros:    render.IntrosBlock{Paragraphs: req.Intros},
			Outros:    render.OutrosBlock{Paragraphs: req.Outros},
			Style:     render.Style{PrimaryColor: req.PrimaryColor},
		}

		if req.HeroImageURL != "" {
			body.Icon = &render.ContentIcon{Src: req.HeroImageURL, Alt: req.HeroImageAlt}
		}

		if req.ButtonText != "" && req.ButtonLink != "" {
			body.Actions = []render.Action{{
				Button: render.Button{Text: req.ButtonText, Link: req.ButtonLink},
			}}
		}

		return body
	},
	Config: func(cfg RuntimeEmailConfig, req BrandedHeroRequest) RuntimeEmailConfig {
		if req.LogoURL != "" {
			cfg.LogoURL = req.LogoURL
		}

		return cfg
	},
}
