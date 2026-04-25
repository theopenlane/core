package email

import (
	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/internal/integrations/providerkit"
)

// VerifyEmailModernRequest is the input for the modern-themed variant of the email
// verification operation. It mirrors VerifyEmailRequest field-for-field; a distinct
// type is required because OperationSchema keys on the input type name
type VerifyEmailModernRequest struct {
	RecipientInfo
	// Token is the email verification token appended to the verify URL
	Token string `json:"token" jsonschema:"required,description=Email verification token"`
}

// verifyEmailModernSchema is the reflected JSON schema for the modern verify input type
// and VerifyEmailModernOp is the typed operation ref used for catalog dispatch
var (
	verifyEmailModernSchema, VerifyEmailModernOp = providerkit.OperationSchema[VerifyEmailModernRequest]()
)

// verifyEmailModernEmail renders the email verification message using the
// openlane-modern card theme with centered logo, icon circle, and CTA
var verifyEmailModernEmail = EmailOperation[VerifyEmailModernRequest]{
	Op: VerifyEmailModernOp, Schema: verifyEmailModernSchema, Theme: openlaneModernTheme,
	Description: "Openlane-modern themed email verification message with centered card layout",
	Subject: func(cfg RuntimeEmailConfig, _ VerifyEmailModernRequest) string {
		return "Please verify your email address to login to " + cfg.CompanyName
	},
	Build: func(cfg RuntimeEmailConfig, req VerifyEmailModernRequest) render.ContentBody {
		verifyURL := tokenURL(cfg.ProductURL, "/verify", req.Token)

		return render.ContentBody{
			Preheader: "Verify Your Email Address",
			Icon:      &render.ContentIcon{Src: iconRocketURL, Alt: "Verify"},
			Name:      req.FirstName,
			Title:     "Verify Your Email Address",
			Intros: render.IntrosBlock{
				Paragraphs: []string{
					"We're almost there!",
					"Thank you for signing up for " + cfg.CompanyName,
					"To verify your account, we just need to confirm your email address.",
				},
			},
			Actions: []render.Action{{
				Button: render.Button{Text: "Confirm Email", Link: verifyURL},
			}},
		}
	},
}
