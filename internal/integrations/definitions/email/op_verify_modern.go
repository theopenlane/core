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
	// HeaderLogo optionally overrides the upper-left header logo; empty falls back to Config.LogoURL
	HeaderLogo string `json:"headerLogo,omitempty" jsonschema:"description=Override URL for the upper-left header logo"`
}

// verifyEmailModernSchema is the reflected JSON schema for the modern verify input type
// and VerifyEmailModernOp is the typed operation ref used for catalog dispatch
var (
	verifyEmailModernSchema, VerifyEmailModernOp = providerkit.OperationSchema[VerifyEmailModernRequest]()
)

// verifyEmailModernEmail renders the same content as verifyEmail using the modern-message
// theme so the two layouts can be compared side-by-side
var verifyEmailModernEmail = EmailOperation[VerifyEmailModernRequest]{
	Op: VerifyEmailModernOp, Schema: verifyEmailModernSchema, Theme: modernMessageTheme,
	Description: "Modern-themed variant of the email verification message for side-by-side theme comparison",
	Subject: func(cfg RuntimeEmailConfig, _ VerifyEmailModernRequest) string {
		return "Please verify your email address to login to " + cfg.CompanyName
	},
	Build: func(cfg RuntimeEmailConfig, req VerifyEmailModernRequest) render.ContentBody {
		verifyURL := cfg.ProductURL + "/verify?token=" + req.Token

		body := render.ContentBody{
			Preheader: "Verify Your Email Address",
			Name:      req.FirstName,
			Title:     "Verify Your Email Address",
			Intros: render.IntrosBlock{
				Paragraphs: []string{
					"We're almost there!",
					"Thank you for signing up for" + cfg.CompanyName,
					"To verify your account, we just need to confirm your email address.",
				},
			},
			Actions: []render.Action{{
				Button: render.Button{Text: "Confirm Email", Link: verifyURL},
			}},
			Outros: render.OutrosBlock{
				Paragraphs: []string{
					"If for some reason the button doesn't work, copy and paste this link into your browser: " + verifyURL,
				},
			},
		}

		if req.HeaderLogo != "" {
			body.Header.Logo = &render.ContentIcon{Src: req.HeaderLogo, Alt: cfg.CompanyName}
		}

		return body
	},
}
