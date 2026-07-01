package email

import (
	"html/template"

	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/internal/integrations/providerkit"
)

// SubprocessorEntry is a single changed vendor listed in a subprocessor notification: the vendor name
// and the kind of change. No logo, link, or other detail is rendered
type SubprocessorEntry struct {
	// Name is the subprocessor display name
	Name string `json:"name" jsonschema:"required,description=Subprocessor name"`
	// Change is the human-readable kind of change (Added, Updated, or Removed)
	Change string `json:"change,omitempty" jsonschema:"description=Kind of change: Added, Updated, or Removed"`
}

// SubprocessorNotificationRequest is the fixed-format system message notifying trust center subscribers
// about subprocessor changes. It is dispatched per recipient as a system email (no email template);
// content (subject, body, vendor list) and the trust center branding overlay are supplied per send by
// the caller so the vendor list always renders in a controlled, consistent layout
type SubprocessorNotificationRequest struct {
	// RecipientInfo carries the recipient email and the per-send unsubscribe token. Unlike the campaign
	// branded message (which is rendered internally), this system email is dispatched directly, so the
	// recipient fields must be part of the reflected schema for dispatch-input validation to accept them
	RecipientInfo
	// Subject is the email subject line
	Subject string `json:"subject" jsonschema:"required,description=Email subject line"`
	// Preheader is hidden preview text shown in the inbox list
	Preheader string `json:"preheader,omitempty" jsonschema:"description=Inbox preview text"`
	// Title is the headline shown above the body
	Title string `json:"title" jsonschema:"required,description=Email headline"`
	// Intros are body paragraphs rendered before the subprocessor table
	Intros []string `json:"intros,omitempty" jsonschema:"description=Body paragraphs rendered before the subprocessor table"`
	// Subprocessors is the list of changed vendors rendered as a formatted table
	Subprocessors []SubprocessorEntry `json:"subprocessors,omitempty" jsonschema:"description=Changed subprocessors rendered as a table"`
	// ButtonText is the call-to-action button label
	ButtonText string `json:"buttonText,omitempty" jsonschema:"description=Call-to-action button label"`
	// ButtonLink is the URL the call-to-action button navigates to
	ButtonLink string `json:"buttonLink,omitempty" jsonschema:"format=uri,description=Call-to-action button URL"`
	// UnsubscribeURL is the per-recipient trust center unsubscribe link shown in the footer, resolved
	// with the subscriber's token by the caller (this system email is dispatched directly and so does
	// not run the campaign's {{ .unsubscribeToken }} template interpolation)
	UnsubscribeURL string `json:"unsubscribeURL,omitempty" jsonschema:"description=Unsubscribe link shown in the footer"`
	// CompanyName overrides the footer company name with the trust center's branding when present
	CompanyName string `json:"companyName,omitempty" jsonschema:"description=Company display name shown in the footer"`
	// Corporation is the legal corporation name used in the footer copyright notice
	Corporation string `json:"corporation,omitempty" jsonschema:"description=Legal corporation name used in the footer copyright notice"`
	// LogoURL overrides the hero logo with the trust center's branding when present
	LogoURL string `json:"logoURL,omitempty" jsonschema:"format=uri,description=Hero logo URL override"`
	// PrimaryColor overrides the headline/emphasis color with the trust center's branding when present
	PrimaryColor string `json:"primaryColor,omitempty" jsonschema:"format=color,description=Primary headline color override (hex)"`
	// ButtonColor overrides the call-to-action button color with the trust center's branding when present
	ButtonColor string `json:"buttonColor,omitempty" jsonschema:"format=color,description=CTA button color override (hex)"`
	// BodyBackgroundColor overrides the outer page background with the trust center's branding when present
	BodyBackgroundColor string `json:"bodyBackgroundColor,omitempty" jsonschema:"format=color,description=Outer page background color override (hex)"`
	// CardBackgroundColor overrides the card background with the trust center's branding when present
	CardBackgroundColor string `json:"cardBackgroundColor,omitempty" jsonschema:"format=color,description=Card container background color override (hex)"`
	// TextColor overrides the body text color with the trust center's branding when present
	TextColor string `json:"textColor,omitempty" jsonschema:"format=color,description=Body text color override (hex)"`
}

// subprocessorNotificationSchema is the reflected JSON schema for the subprocessor notification input
// type and SubprocessorNotificationOp is the typed operation ref used for catalog dispatch
var (
	subprocessorNotificationSchema, SubprocessorNotificationOp = providerkit.OperationSchema[SubprocessorNotificationRequest]() //nolint:revive
)

var _ = RegisterEmailOperation(Operation[SubprocessorNotificationRequest]{
	Op: SubprocessorNotificationOp, Schema: subprocessorNotificationSchema, Theme: baseTheme,
	Description: "System notification listing subprocessor changes for a trust center's subscribers",
	Subject: func(_ RuntimeEmailConfig, req SubprocessorNotificationRequest) string {
		return req.Subject
	},
	Build: func(_ RuntimeEmailConfig, req SubprocessorNotificationRequest) render.ContentBody {
		body := render.ContentBody{
			Preheader: req.Preheader,
			Name:      req.FirstName,
			Title:     req.Title,
			Intros:    render.IntrosBlock{Paragraphs: req.Intros},
			Callout:   subprocessorCallout(req.Subprocessors),
		}

		if req.ButtonText != "" && req.ButtonLink != "" {
			body.Actions = []render.Action{{
				Button: render.Button{Text: req.ButtonText, Link: req.ButtonLink},
			}}
		}

		return body
	},
	Config: func(cfg RuntimeEmailConfig, req SubprocessorNotificationRequest) RuntimeEmailConfig {
		// start from the Openlane system treatment, then overlay only the trust center branding
		// values that are present so the format stays controlled while still honoring TS branding
		cfg = applySystemBranding(cfg)

		if req.CompanyName != "" {
			cfg.CompanyName = req.CompanyName
		}

		if req.Corporation != "" {
			cfg.Corporation = req.Corporation
		}

		if req.LogoURL != "" {
			cfg.LogoURL = req.LogoURL
			cfg.HeaderLogoURL = req.LogoURL
		}

		if req.PrimaryColor != "" {
			cfg.HeadingColor = req.PrimaryColor
		}

		if req.ButtonColor != "" {
			cfg.ButtonColor = req.ButtonColor
		}

		if req.BodyBackgroundColor != "" {
			cfg.BodyBackgroundColor = req.BodyBackgroundColor
		}

		if req.CardBackgroundColor != "" {
			cfg.CardBackgroundColor = req.CardBackgroundColor
		}

		if req.TextColor != "" {
			cfg.TextColor = req.TextColor
		}

		if req.UnsubscribeURL != "" {
			cfg.UnsubscribeURL = req.UnsubscribeURL
		}

		return cfg
	},
})

// subprocessorCallout renders the changed subprocessors as a callout list, one entry per vendor showing
// just the name and the kind of change (e.g. "Acme — Added"). render.Bold escapes the text, so no HTML is
// hand-assembled. The callout is rendered by the base theme
func subprocessorCallout(entries []SubprocessorEntry) *render.Callout {
	items := make([]template.HTML, 0, len(entries))
	for _, entry := range entries {
		items = append(items, render.Bold(entry.Name+" — "+entry.Change))
	}

	return &render.Callout{
		Title: "Subprocessors",
		Items: items,
	}
}
