package email

import (
	"html/template"

	"github.com/samber/lo"
	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// subprocessorNotificationSubject and the related copy back the subprocessor change email when the trust center branding carries no company name; sends with a company name use the company-specific copy composed in the operation
const (
	subprocessorNotificationSubject    = "Subprocessor update"
	subprocessorNotificationPreheader  = "Review the latest changes to our subprocessor list"
	subprocessorNotificationTitle      = "We've updated our subprocessors"
	subprocessorNotificationIntro      = "The subprocessors we use have changed. You can review the full list anytime in our trust center."
	subprocessorNotificationButtonText = "View subprocessors"
)

// SubprocessorEntry is a single changed vendor listed in a subprocessor notification: the vendor name and the kind of change
type SubprocessorEntry struct {
	// Name is the subprocessor display name
	Name string `json:"name" jsonschema:"required,description=Subprocessor name"`
	// Change is the human-readable kind of change (Added, Updated, or Removed)
	Change string `json:"change,omitempty" jsonschema:"description=Kind of change: Added, Updated, or Removed"`
}

// SubprocessorNotificationRequest is the fixed-format system message notifying trust center subscribers about subprocessor changes
type SubprocessorNotificationRequest struct {
	// RecipientInfo carries the recipient email and the per-send unsubscribe token
	RecipientInfo
	// TrustCenterBranding is the trust center's visual identity overlay; empty values fall back to the Openlane system branding
	TrustCenterBranding
	// Subprocessors is the list of changed vendors rendered as a formatted list
	Subprocessors []SubprocessorEntry `json:"subprocessors,omitempty" jsonschema:"description=Changed subprocessors rendered as a list"`
	// TrustCenterURL is the public trust center link the call-to-action button navigates to
	TrustCenterURL string `json:"trustCenterURL,omitempty" jsonschema:"format=uri,description=Public trust center URL for the call-to-action button"`
	// UnsubscribeURL is the per-recipient trust center unsubscribe link shown in the footer, resolved with the subscriber's token by the caller since this system email is dispatched directly and does not run the campaign's {{ .unsubscribeToken }} template interpolation
	UnsubscribeURL string `json:"unsubscribeURL,omitempty" jsonschema:"description=Unsubscribe link shown in the footer"`
}

// SubprocessorNotificationOp is the typed operation ref used for catalog dispatch
var SubprocessorNotificationOp = types.OperationRefOf[SubprocessorNotificationRequest]() //nolint:revive

var _ = RegisterEmailOperation(Operation[SubprocessorNotificationRequest]{
	Op: SubprocessorNotificationOp, Theme: baseTheme,
	Description: "System notification listing subprocessor changes for a trust center's subscribers",
	Subject: func(_ RuntimeEmailConfig, req SubprocessorNotificationRequest) string {
		if req.CompanyName != "" {
			return req.CompanyName + " Trust Center Subprocessor Update"
		}

		return subprocessorNotificationSubject
	},
	Build: func(_ RuntimeEmailConfig, req SubprocessorNotificationRequest) render.ContentBody {
		preheader := subprocessorNotificationPreheader
		title := subprocessorNotificationTitle
		intro := subprocessorNotificationIntro

		if req.CompanyName != "" {
			preheader = "Review the latest changes to the " + req.CompanyName + " subprocessor list"
			title = req.CompanyName + " has updated its subprocessors"
			intro = "The subprocessors " + req.CompanyName + " uses have changed. You can review the full list anytime in the " + req.CompanyName + " trust center."
		}

		body := render.ContentBody{
			Preheader: preheader,
			Name:      req.FirstName,
			Title:     title,
			Intros:    render.IntrosBlock{Paragraphs: []string{intro}},
			Callout:   subprocessorCallout(req.Subprocessors),
		}

		if req.TrustCenterURL != "" {
			body.Actions = []render.Action{{
				Button: render.Button{Text: subprocessorNotificationButtonText, Link: req.TrustCenterURL},
			}}
		}

		return body
	},
	Config: func(cfg RuntimeEmailConfig, req SubprocessorNotificationRequest) RuntimeEmailConfig {
		return trustCenterEmailConfig(cfg, req.TrustCenterBranding, req.UnsubscribeURL)
	},
})

// subprocessorCallout renders the changed subprocessors as a callout list, one entry per vendor showing just the name and the kind of change (e.g. "Acme — Added")
func subprocessorCallout(entries []SubprocessorEntry) *render.Callout {
	return &render.Callout{
		Title: "Subprocessors",
		Items: lo.Map(entries, func(entry SubprocessorEntry, _ int) template.HTML {
			return render.Bold(entry.Name + " — " + entry.Change)
		}),
	}
}
