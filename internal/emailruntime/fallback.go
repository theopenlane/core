package emailruntime

import (
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
)

// fallbackTemplate holds minimal template content used when no system template is seeded in the database
type fallbackTemplate struct {
	// SubjectTemplate is the plain-text subject Go template
	SubjectTemplate string
	// BodyTemplate is the minimal HTML body Go template
	BodyTemplate string
}

// toNotificationTemplate synthesizes a minimal NotificationTemplate from the fallback definition
func (f fallbackTemplate) toNotificationTemplate(key string) *generated.NotificationTemplate {
	return &generated.NotificationTemplate{
		Key:             key,
		Name:            key,
		Channel:         enums.ChannelEmail,
		Format:          enums.NotificationTemplateFormatHTML,
		SubjectTemplate: f.SubjectTemplate,
		BodyTemplate:    f.BodyTemplate,
		Active:          true,
		SystemOwned:     true,
	}
}

// systemFallbackTemplates defines minimal fallback templates for each system notification key.
// These are returned when no matching system-owned template record exists in the database
var systemFallbackTemplates = map[string]fallbackTemplate{
	TemplateKeyVerifyEmail: {
		SubjectTemplate: "Verify your email address",
		BodyTemplate:    `<p>Please verify your email address.{{if .URLS.Verify}} <a href="{{.URLS.Verify}}">Click here to verify</a>{{end}}</p>`,
	},
	TemplateKeyWelcome: {
		SubjectTemplate: "Welcome to {{.CompanyName}}",
		BodyTemplate:    `<p>Your account has been created successfully.</p>`,
	},
	TemplateKeyInvite: {
		SubjectTemplate: "You have been invited",
		BodyTemplate:    `<p>You have been invited to join an organization.{{if .URLS.Invite}} <a href="{{.URLS.Invite}}">Accept invitation</a>{{end}}</p>`,
	},
	TemplateKeyInviteJoined: {
		SubjectTemplate: "Invitation accepted",
		BodyTemplate:    `<p>Your invitation has been accepted.</p>`,
	},
	TemplateKeyPasswordResetRequest: {
		SubjectTemplate: "Reset your password",
		BodyTemplate:    `<p>A password reset was requested for your account.{{if .URLS.PasswordReset}} <a href="{{.URLS.PasswordReset}}">Reset password</a>{{end}}</p>`,
	},
	TemplateKeyPasswordResetSuccess: {
		SubjectTemplate: "Password reset successful",
		BodyTemplate:    `<p>Your password has been reset successfully.</p>`,
	},
	TemplateKeySubscribe: {
		SubjectTemplate: "Verify your subscription",
		BodyTemplate:    `<p>Please verify your subscription.{{if .URLS.VerifySubscriber}} <a href="{{.URLS.VerifySubscriber}}">Verify subscription</a>{{end}}</p>`,
	},
	TemplateKeyVerifyBilling: {
		SubjectTemplate: "Verify your billing email",
		BodyTemplate:    `<p>Please verify your billing email address.{{if .URLS.VerifyBilling}} <a href="{{.URLS.VerifyBilling}}">Verify billing email</a>{{end}}</p>`,
	},
	TemplateKeyTrustCenterNDARequest: {
		SubjectTemplate: "NDA signature requested",
		BodyTemplate:    `<p>An NDA signature has been requested.</p>`,
	},
	TemplateKeyTrustCenterNDASigned: {
		SubjectTemplate: "NDA signed",
		BodyTemplate:    `<p>The NDA has been signed.</p>`,
	},
	TemplateKeyTrustCenterAuth: {
		SubjectTemplate: "Trust center access",
		BodyTemplate:    `<p>Please access the trust center.</p>`,
	},
	TemplateKeyQuestionnaireAuth: {
		SubjectTemplate: "Questionnaire access",
		BodyTemplate:    `<p>Please access the questionnaire.</p>`,
	},
	TemplateKeyBillingEmailChanged: {
		SubjectTemplate: "Billing email updated",
		BodyTemplate:    `<p>Your billing email address has been updated.</p>`,
	},
}
