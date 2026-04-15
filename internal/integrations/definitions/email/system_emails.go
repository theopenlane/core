package email

import (
	"html/template"
	"strings"
	"time"

	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/internal/integrations/definitions/email/themes"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Theme instances shared across email operations
var (
	standardTheme      = themes.Standard
	trustCenterTheme   = themes.TrustCenter
	questionnaireTheme = themes.Questionnaire
)

// Icon image URLs hosted on the Openlane CDN
const (
	iconRocketURL    = "https://www.theopenlane.io/cdn-cgi/imagedelivery/2gi-D0CFOlSOflWJG-LQaA/b5d07352-e391-4ac6-41f8-e1ee9185e000/public"
	iconUserPlusURL  = "https://www.theopenlane.io/cdn-cgi/imagedelivery/2gi-D0CFOlSOflWJG-LQaA/a177b189-bf03-466a-e43a-542585eb1800/public"
	iconUserCheckURL = "https://www.theopenlane.io/cdn-cgi/imagedelivery/2gi-D0CFOlSOflWJG-LQaA/22704d4a-a811-44c0-8618-8309b03dfa00/public"
	iconBellURL      = "https://www.theopenlane.io/cdn-cgi/imagedelivery/2gi-D0CFOlSOflWJG-LQaA/23690f00-2ddb-4d22-e9a5-af470c93c100/public"
)

// Trust center / questionnaire button colors matching the original teal theme
const (
	tcButtonColor     = "#3fc2b4"
	tcButtonTextColor = "#ffffff"
)

// VerifyEmailRequest is the input for the email verification operation
type VerifyEmailRequest struct {
	RecipientInfo
	// Token is the email verification token appended to the verify URL
	Token string `json:"token" jsonschema:"required,description=Email verification token"`
}

// WelcomeRequest is the input for the welcome email operation
type WelcomeRequest struct {
	RecipientInfo
}

// InviteRequest is the input for the organization invite operation
type InviteRequest struct {
	RecipientInfo
	// InviterName is the name of the person who sent the invite
	InviterName string `json:"inviter_name" jsonschema:"required,description=Name of the person sending the invite"`
	// OrgName is the organization being invited to
	OrgName string `json:"org_name" jsonschema:"required,description=Organization name"`
	// Role is the invited role
	Role string `json:"role,omitempty" jsonschema:"description=Invited role"`
	// Token is the invite token appended to the invite URL
	Token string `json:"token" jsonschema:"required,description=Invite token"`
}

// InviteJoinedRequest is the input for the invite-accepted notification
type InviteJoinedRequest struct {
	RecipientInfo
	// OrgName is the organization the user joined
	OrgName string `json:"org_name" jsonschema:"required,description=Organization name"`
}

// PasswordResetEmailRequest is the input for the password reset request operation
type PasswordResetEmailRequest struct {
	RecipientInfo
	// Token is the password reset token appended to the reset URL
	Token string `json:"token" jsonschema:"required,description=Password reset token"`
}

// PasswordResetSuccessRequest is the input for the password reset confirmation operation
type PasswordResetSuccessRequest struct {
	RecipientInfo
}

// SubscribeRequest is the input for the subscription verification operation
type SubscribeRequest struct {
	RecipientInfo
	// OrgName is the display name of the subscribing organization
	OrgName string `json:"org_name" jsonschema:"required,description=Organization display name"`
	// Token is the subscriber verification token appended to the verify URL
	Token string `json:"token" jsonschema:"required,description=Subscriber verification token"`
}

// VerifyBillingRequest is the input for the billing email verification operation
type VerifyBillingRequest struct {
	RecipientInfo
	// Token is the billing verification token appended to the verify URL
	Token string `json:"token" jsonschema:"required,description=Billing verification token"`
}

// TrustCenterNDARequestEmail is the input for the trust center NDA signing request
type TrustCenterNDARequestEmail struct {
	RecipientInfo
	// OrgName is the organization requesting the NDA
	OrgName string `json:"org_name" jsonschema:"required,description=Organization name"`
	// NDAURL is the direct URL to the NDA signing page
	NDAURL string `json:"nda_url" jsonschema:"required,description=NDA signing URL"`
}

// TrustCenterNDASignedEmail is the input for the NDA signed confirmation
type TrustCenterNDASignedEmail struct {
	RecipientInfo
	// OrgName is the organization whose NDA was signed
	OrgName string `json:"org_name" jsonschema:"required,description=Organization name"`
	// TrustCenterURL is the URL to the trust center
	TrustCenterURL string `json:"trust_center_url" jsonschema:"required,description=Trust center URL"`
	// AttachmentFilename is the filename for the signed NDA attachment
	AttachmentFilename string `json:"attachment_filename,omitempty" jsonschema:"description=Signed NDA attachment filename"`
	// AttachmentData is the raw content of the signed NDA attachment
	AttachmentData []byte `json:"attachment_data,omitempty" jsonschema:"description=Signed NDA attachment content"`
}

// TrustCenterAuthEmail is the input for trust center access authentication
type TrustCenterAuthEmail struct {
	RecipientInfo
	// OrgName is the organization whose trust center is being accessed
	OrgName string `json:"org_name" jsonschema:"required,description=Organization name"`
	// AuthURL is the authentication URL for trust center access
	AuthURL string `json:"auth_url" jsonschema:"required,description=Trust center auth URL"`
}

// QuestionnaireAuthEmail is the input for questionnaire access authentication
type QuestionnaireAuthEmail struct {
	RecipientInfo
	// AssessmentName is the name of the assessment/questionnaire
	AssessmentName string `json:"assessment_name" jsonschema:"required,description=Assessment name"`
	// AuthURL is the authentication URL for questionnaire access
	AuthURL string `json:"auth_url" jsonschema:"required,description=Questionnaire auth URL"`
}

// BillingEmailChangedEmail is the input for the billing email change notification
type BillingEmailChangedEmail struct {
	RecipientInfo
	// OrgName is the organization whose billing email changed
	OrgName string `json:"org_name" jsonschema:"required,description=Organization name"`
	// OldBillingEmail is the previous billing email address
	OldBillingEmail string `json:"old_billing_email" jsonschema:"required,description=Previous billing email"`
	// NewBillingEmail is the new billing email address
	NewBillingEmail string `json:"new_billing_email" jsonschema:"required,description=New billing email"`
	// ChangedAt is the timestamp of the change
	ChangedAt time.Time `json:"changed_at" jsonschema:"required,description=Timestamp of the change"`
}

// --- Schema + operation ref vars ---

var (
	verifyEmailSchema, VerifyEmailOp                 = providerkit.OperationSchema[VerifyEmailRequest]()
	welcomeSchema, WelcomeOp                         = providerkit.OperationSchema[WelcomeRequest]()
	inviteSchema, InviteOp                           = providerkit.OperationSchema[InviteRequest]()
	inviteJoinedSchema, InviteJoinedOp               = providerkit.OperationSchema[InviteJoinedRequest]()
	resetRequestSchema, ResetRequestOp               = providerkit.OperationSchema[PasswordResetEmailRequest]()
	resetSuccessSchema, ResetSuccessOp               = providerkit.OperationSchema[PasswordResetSuccessRequest]()
	subscribeSchema, SubscribeOp                     = providerkit.OperationSchema[SubscribeRequest]()
	verifyBillingSchema, VerifyBillingOp             = providerkit.OperationSchema[VerifyBillingRequest]()
	tcNDARequestSchema, TCNDARequestOp               = providerkit.OperationSchema[TrustCenterNDARequestEmail]()
	tcNDASignedSchema, TCNDASignedOp                 = providerkit.OperationSchema[TrustCenterNDASignedEmail]()
	tcAuthSchema, TCAuthOp                           = providerkit.OperationSchema[TrustCenterAuthEmail]()
	questionnaireAuthSchema, QuestionnaireAuthOp     = providerkit.OperationSchema[QuestionnaireAuthEmail]()
	billingEmailChangedSchema, BillingEmailChangedOp = providerkit.OperationSchema[BillingEmailChangedEmail]()
)

// --- Email operation definitions ---

var verifyEmail = EmailOperation[VerifyEmailRequest]{
	Op: VerifyEmailOp, Schema: verifyEmailSchema, Theme: standardTheme,
	Subject: func(cfg RuntimeEmailConfig, _ VerifyEmailRequest) string {
		return "Please verify your email address to login to " + cfg.CompanyName
	},
	Content: func(cfg RuntimeEmailConfig, req VerifyEmailRequest) render.EmailContent {
		verifyURL := cfg.ProductURL + "/verify?token=" + req.Token

		return render.EmailContent{
			Body: render.ContentBody{
				Preheader: "Verify Your Email Address",
				Name:      req.FirstName,
				Title:     "Verify Your Email Address",
				Intros: []string{
					"Welcome to " + cfg.CompanyName + " — where compliance isn't just a checkbox.",
					"Before you get started, let's make sure it's really you. Click below to verify your email:",
				},
				Actions: []render.Action{{
					Button: render.Button{Text: "Confirm Email", Link: verifyURL},
				}},
				Outros: []string{
					"Or, if you're feeling old school, copy and paste this link into your browser: " + verifyURL,
					"This link expires in 7 days — but don't worry, if it does, you'll get a fresh one when you try to verify later.",
				},
			},
		}
	},
}

var welcomeEmail = EmailOperation[WelcomeRequest]{
	Op: WelcomeOp, Schema: welcomeSchema, Theme: standardTheme,
	Subject: func(cfg RuntimeEmailConfig, _ WelcomeRequest) string {
		return "Welcome to " + cfg.CompanyName + "!"
	},
	Content: func(cfg RuntimeEmailConfig, req WelcomeRequest) render.EmailContent {
		return render.EmailContent{
			Body: render.ContentBody{
				Preheader: "Welcome to " + cfg.CompanyName + "!",
				Icon:      &render.ContentIcon{Src: iconRocketURL, Alt: "Rocket"},
				Name:      req.FirstName,
				Title:     "Welcome to " + cfg.CompanyName + "!",
				Intros: []string{
					"We're thrilled to have you here. At " + cfg.CompanyName + ", we're working to develop a cutting-edge cybersecurity and compliance automation solution to help organizations of all sizes and industries secure their systems, navigate the increasingly complex web of privacy laws and regulations, ensure continuous compliance, manage risks, and get ahead of evolving cyber threats.",
				},
				ContentBlocks: []template.HTML{
					template.HTML(`<div style="text-align:left;margin-bottom:21px;background-color:rgb(240,253,249);padding:20px;border-radius:8px">` +
						`<p style="font-size:16px;font-weight:500;margin-bottom:15px;line-height:24px;margin:16px 0">Here's how to get started:</p>` +
						`<p style="font-size:16px;margin-bottom:12px;padding-left:20px;line-height:24px;margin:16px 0">1. Go through our onboarding process to get a personalized experience</p>` +
						`<p style="font-size:16px;margin-bottom:12px;padding-left:20px;line-height:24px;margin:16px 0">2. Setup a test program to see all the features our platform has to offer</p>` +
						`<p style="font-size:16px;margin-bottom:0px;padding-left:20px;line-height:24px;margin:16px 0">3. Checkout our <a href="` + cfg.DocsURL + `" style="color:rgb(63,118,255);text-decoration-line:none" target="_blank">quickstart guide</a> for helpful resources</p>` +
						`</div>`),
				},
				Actions: []render.Action{{
					Button: render.Button{Text: "Get Started", Link: cfg.ProductURL},
				}},
			},
		}
	},
}

var inviteEmail = EmailOperation[InviteRequest]{
	Op: InviteOp, Schema: inviteSchema, Theme: standardTheme,
	Subject: func(cfg RuntimeEmailConfig, req InviteRequest) string {
		return "Join Your Teammate " + req.InviterName + " on " + cfg.CompanyName + "!"
	},
	Content: func(cfg RuntimeEmailConfig, req InviteRequest) render.EmailContent {
		inviteURL := cfg.ProductURL + "/invite?token=" + req.Token

		intro := "You're in — let's build trust without the busywork. " + req.InviterName + " has invited you to collaborate in " + cfg.CompanyName + ", as part of the <b>" + req.OrgName + "</b> organization"
		if req.Role != "" {
			intro += " with the role of " + strings.ToUpper(req.Role)
		}

		intro += "."

		return render.EmailContent{
			Body: render.ContentBody{
				Preheader: "You've been invited to join " + cfg.CompanyName,
				Icon:      &render.ContentIcon{Src: iconUserPlusURL, Alt: "User-Plus"},
				Name:      req.FirstName,
				Title:     "You've been invited to join " + cfg.CompanyName + "!",
				IntrosUnsafe: []template.HTML{
					template.HTML(intro),
					"To get started (and verify your email), click the link below:",
				},
				Actions: []render.Action{{
					Button: render.Button{Text: "Accept Invite", Link: inviteURL},
				}},
				Outros: []string{
					"Or, if you're feeling old school, copy and paste this link into your browser: " + inviteURL,
					"This link expires in 7 days — but don't worry, if it does, you'll get a fresh one when you try to verify later.",
				},
			},
		}
	},
}

var inviteJoinedEmail = EmailOperation[InviteJoinedRequest]{
	Op: InviteJoinedOp, Schema: inviteJoinedSchema, Theme: standardTheme,
	Subject: func(cfg RuntimeEmailConfig, _ InviteJoinedRequest) string {
		return "You've been added to an Organization on " + cfg.CompanyName
	},
	Content: func(cfg RuntimeEmailConfig, req InviteJoinedRequest) render.EmailContent {
		return render.EmailContent{
			Body: render.ContentBody{
				Preheader: "You're in — welcome to " + req.OrgName + " on " + cfg.CompanyName + "!",
				Icon:      &render.ContentIcon{Src: iconUserCheckURL, Alt: "User-Check"},
				Name:      req.FirstName,
				Title:     "You're in — welcome to " + req.OrgName + " on " + cfg.CompanyName + "!",
				Intros: []string{
					"Welcome to " + cfg.CompanyName + " — we're excited to have you on board with " + req.OrgName + "!",
					"Ditch the spreadsheets, and embrace the pipeline with a faster, cleaner way to automate compliance, manage risk, and stay ahead of security threats — without the manual overhead.",
				},
				ContentBlocks: []template.HTML{
					template.HTML(`<div style="text-align:left;margin-bottom:21px;background-color:rgb(240,253,249);padding:20px;border-radius:8px">` +
						`<p style="font-size:16px;font-weight:500;margin-bottom:15px;line-height:24px;margin:16px 0">Here's how to get started:</p>` +
						`<p style="font-size:16px;margin-bottom:12px;padding-left:20px;line-height:24px;margin:16px 0">1. Complete the onboarding flow to tailor your experience</p>` +
						`<p style="font-size:16px;margin-bottom:12px;padding-left:20px;line-height:24px;margin:16px 0">2. Explore any active <a href="` + cfg.ProductURL + `/programs" style="color:rgb(63,118,255);text-decoration-line:none" target="_blank">programs</a> your team is running</p>` +
						`<p style="font-size:16px;margin-bottom:0px;padding-left:20px;line-height:24px;margin:16px 0">3. Checkout our <a href="` + cfg.DocsURL + `" style="color:rgb(63,118,255);text-decoration-line:none" target="_blank">quickstart guide</a> for helpful resources</p>` +
						`</div>`),
				},
				Outros: []string{
					"When you're ready, hop into the platform and start exploring:",
				},
				Actions: []render.Action{{
					Button: render.Button{Text: "Get Started", Link: cfg.ProductURL},
				}},
			},
		}
	},
}

var resetRequestEmail = EmailOperation[PasswordResetEmailRequest]{
	Op: ResetRequestOp, Schema: resetRequestSchema, Theme: standardTheme,
	Subject: func(cfg RuntimeEmailConfig, _ PasswordResetEmailRequest) string {
		return cfg.CompanyName + " Password Reset - Action Required"
	},
	Content: func(cfg RuntimeEmailConfig, req PasswordResetEmailRequest) render.EmailContent {
		resetURL := cfg.ProductURL + "/password-reset?token=" + req.Token

		return render.EmailContent{
			Body: render.ContentBody{
				Preheader: "Reset your " + cfg.CompanyName + " password",
				Name:      req.FirstName,
				Title:     "Reset your " + cfg.CompanyName + " password",
				Intros: []string{
					"We received a request to reset your " + cfg.CompanyName + " password.",
					"If that was you, no problem — just click the button below to set a new one:",
				},
				Actions: []render.Action{{
					Button: render.Button{Text: "Password Reset", Link: resetURL},
				}},
				Outros: []string{
					"No button? No problem — just paste this link in your browser: " + resetURL,
					"This link will expire in 15 minutes to keep things secure.",
					"If you didn't request a password reset, you can safely ignore this email.",
				},
			},
		}
	},
}

var resetSuccessEmail = EmailOperation[PasswordResetSuccessRequest]{
	Op: ResetSuccessOp, Schema: resetSuccessSchema, Theme: standardTheme,
	Subject: func(cfg RuntimeEmailConfig, _ PasswordResetSuccessRequest) string {
		return cfg.CompanyName + " Password Reset Confirmation"
	},
	Content: func(cfg RuntimeEmailConfig, req PasswordResetSuccessRequest) render.EmailContent {
		return render.EmailContent{
			Body: render.ContentBody{
				Preheader: "Your " + cfg.CompanyName + " password has been reset",
				Name:      req.FirstName,
				Title:     "Your " + cfg.CompanyName + " password has been reset",
				Intros: []string{
					"Your password was successfully updated. If you made this change, you're all set — no further action needed.",
					"If you didn't request a password reset, please contact our support team right away at " + cfg.SupportEmail + ". Keeping your account secure is our top priority.",
				},
			},
		}
	},
}

var subscribeEmail = EmailOperation[SubscribeRequest]{
	Op: SubscribeOp, Schema: subscribeSchema, Theme: standardTheme,
	Subject: func(cfg RuntimeEmailConfig, _ SubscribeRequest) string {
		return "You've been subscribed to " + cfg.CompanyName
	},
	Content: func(cfg RuntimeEmailConfig, req SubscribeRequest) render.EmailContent {
		verifyURL := cfg.ProductURL + "/subscribe/verify?token=" + req.Token

		return render.EmailContent{
			Body: render.ContentBody{
				Preheader: "You're In — Early Access Secured! Thanks for your interest in our beta program.",
				Icon:      &render.ContentIcon{Src: iconBellURL, Alt: "Notification Bell"},
				Name:      req.FirstName,
				Title:     "You're In — Early Access Secured!",
				Intros: []string{
					"We're thrilled to have you as part of our early community. Your interest means the world to us as we work to build a cutting-edge solution. We can't wait to share it with you!",
					"Please confirm your email address to ensure you receive all important updates about your early access.",
				},
				Actions: []render.Action{{
					Button: render.Button{Text: "Confirm Email", Link: verifyURL},
				}},
				ContentBlocks: []template.HTML{
					template.HTML(`<div style="text-align:left;margin-bottom:21px;background-color:rgb(240,253,249);padding:20px;border-radius:8px">` +
						`<p style="font-size:16px;line-height:24px;margin-bottom:15px;margin-top:16px;font-weight:500">What to Expect Next:</p>` +
						`<p style="font-size:16px;line-height:24px;margin-bottom:12px;margin-top:16px;padding-left:20px">` + "\u2705" + ` You'll hear from us soon – We'll email you as soon as your spot is ready.</p>` +
						`<p style="font-size:16px;line-height:24px;margin-bottom:12px;margin-top:16px;padding-left:20px">` + "\U0001F680" + ` Early access to beta features – Get a first look at everything we're building.</p>` +
						`<p style="font-size:16px;line-height:24px;margin-bottom:0px;margin-top:16px;padding-left:20px">` + "\U0001F4A1" + ` Help shape the future – Your feedback will directly influence the product.</p>` +
						`</div>`),
				},
				Outros: []string{
					"Thank you for being part of this journey — we're excited to have you on board!",
				},
			},
		}
	},
}

var verifyBillingEmail = EmailOperation[VerifyBillingRequest]{
	Op: VerifyBillingOp, Schema: verifyBillingSchema, Theme: standardTheme,
	Subject: func(cfg RuntimeEmailConfig, _ VerifyBillingRequest) string {
		return "Please verify the billing email for " + cfg.CompanyName + " to ensure your account stays up to date"
	},
	Content: func(cfg RuntimeEmailConfig, req VerifyBillingRequest) render.EmailContent {
		verifyURL := cfg.ProductURL + "/billing/verify?token=" + req.Token

		return render.EmailContent{
			Body: render.ContentBody{
				Preheader: "Verify Your Email Address",
				Name:      req.FirstName,
				Title:     "Verify Your Billing Email Address",
				Intros: []string{
					"You're receiving this because the billing contact for your " + cfg.CompanyName + " account was just updated.",
					"To help keep your account secure, please verify your email address by clicking the button below:",
				},
				Actions: []render.Action{{
					Button: render.Button{Text: "Confirm Billing Email", Link: verifyURL},
				}},
				Outros: []string{
					"Or, if you're feeling old school, copy and paste this link into your browser: " + verifyURL,
					"If you run into any issues, feel free to reach out at " + cfg.SupportEmail + ".",
				},
			},
		}
	},
}

var tcNDARequestEmail = EmailOperation[TrustCenterNDARequestEmail]{
	Op: TCNDARequestOp, Schema: tcNDARequestSchema, Theme: trustCenterTheme,
	Subject: func(_ RuntimeEmailConfig, req TrustCenterNDARequestEmail) string {
		return req.OrgName + " Trust Center NDA Request"
	},
	Content: func(_ RuntimeEmailConfig, req TrustCenterNDARequestEmail) render.EmailContent {
		return render.EmailContent{
			Body: render.ContentBody{
				Title: "You requested access to " + req.OrgName + "'s Trust Center",
				Intros: []string{
					"To continue, please review and sign the Non-Disclosure Agreement (NDA). Once signed, you'll be granted access to protected Trust Center documents.",
				},
				Actions: []render.Action{{
					Button: render.Button{Text: "Sign NDA", Link: req.NDAURL, Color: tcButtonColor, TextColor: tcButtonTextColor},
				}},
				Outros: []string{
					"If the button doesn't work, copy and paste this link into your browser: " + req.NDAURL,
				},
			},
		}
	},
}

var tcNDASignedEmail = EmailOperation[TrustCenterNDASignedEmail]{
	Op: TCNDASignedOp, Schema: tcNDASignedSchema, Theme: trustCenterTheme,
	Subject: func(_ RuntimeEmailConfig, req TrustCenterNDASignedEmail) string {
		return req.OrgName + " Trust Center NDA Signed"
	},
	Content: func(_ RuntimeEmailConfig, req TrustCenterNDASignedEmail) render.EmailContent {
		return render.EmailContent{
			Body: render.ContentBody{
				Title: "Your NDA with " + req.OrgName + " has been signed",
				Intros: []string{
					"Thank you for signing the Non-Disclosure Agreement (NDA). You now have access to " + req.OrgName + "'s protected Trust Center documents.",
				},
				Actions: []render.Action{{
					Button: render.Button{Text: "Visit Trust Center", Link: req.TrustCenterURL, Color: tcButtonColor, TextColor: tcButtonTextColor},
				}},
				Outros: []string{
					"If the button doesn't work, copy and paste this link into your browser: " + req.TrustCenterURL,
				},
			},
		}
	},
	MessageOptions: func(_ RuntimeEmailConfig, req TrustCenterNDASignedEmail) []newman.MessageOption {
		if req.AttachmentFilename == "" || len(req.AttachmentData) == 0 {
			return nil
		}

		return []newman.MessageOption{
			newman.WithAttachment(newman.NewAttachment(req.AttachmentFilename, req.AttachmentData)),
		}
	},
}

var tcAuthEmail = EmailOperation[TrustCenterAuthEmail]{
	Op: TCAuthOp, Schema: tcAuthSchema, Theme: trustCenterTheme,
	Subject: func(_ RuntimeEmailConfig, req TrustCenterAuthEmail) string {
		return "Access " + req.OrgName + "'s Trust Center"
	},
	Content: func(_ RuntimeEmailConfig, req TrustCenterAuthEmail) render.EmailContent {
		return render.EmailContent{
			Body: render.ContentBody{
				Title: "Access " + req.OrgName + "'s Trust Center",
				Intros: []string{
					"You've been granted access to " + req.OrgName + "'s Trust Center. Click the button below to authenticate and view the available resources.",
				},
				Actions: []render.Action{{
					Button: render.Button{Text: "Access Trust Center", Link: req.AuthURL, Color: tcButtonColor, TextColor: tcButtonTextColor},
				}},
				Outros: []string{
					"This authentication link provides secure, time-limited access and will expire after a short period for your security.",
					"If the button doesn't work, copy and paste this link into your browser: " + req.AuthURL,
				},
			},
		}
	},
}

var questionnaireAuthEmail = EmailOperation[QuestionnaireAuthEmail]{
	Op: QuestionnaireAuthOp, Schema: questionnaireAuthSchema, Theme: questionnaireTheme,
	Subject: func(cfg RuntimeEmailConfig, req QuestionnaireAuthEmail) string {
		return "Access " + req.AssessmentName + " Questionnaire from " + cfg.CompanyName
	},
	Content: func(cfg RuntimeEmailConfig, req QuestionnaireAuthEmail) render.EmailContent {
		return render.EmailContent{
			Body: render.ContentBody{
				Title: cfg.CompanyName + " sent you an assessment to complete",
				Intros: []string{
					cfg.CompanyName + " has shared a form (" + req.AssessmentName + ") for you to complete. Click the button below to access it.",
				},
				Actions: []render.Action{{
					Button: render.Button{Text: "Access Questionnaire", Link: req.AuthURL, Color: tcButtonColor, TextColor: tcButtonTextColor},
				}},
				Outros: []string{
					"This authentication link provides secure, time-limited access and will expire after a short period for your security.",
					"If the button doesn't work, copy and paste this link into your browser: " + req.AuthURL,
				},
			},
		}
	},
	MessageOptions: func(cfg RuntimeEmailConfig, _ QuestionnaireAuthEmail) []newman.MessageOption {
		if cfg.QuestionnaireEmail == "" {
			return nil
		}

		return []newman.MessageOption{
			newman.WithFrom(cfg.QuestionnaireEmail),
		}
	},
}

var billingChangedEmail = EmailOperation[BillingEmailChangedEmail]{
	Op: BillingEmailChangedOp, Schema: billingEmailChangedSchema, Theme: standardTheme,
	Subject: func(_ RuntimeEmailConfig, req BillingEmailChangedEmail) string {
		return "Billing Email Changed for " + req.OrgName
	},
	Content: func(cfg RuntimeEmailConfig, req BillingEmailChangedEmail) render.EmailContent {
		return render.EmailContent{
			Body: render.ContentBody{
				Preheader: "Billing Email Changed",
				Title:     "Billing Email Changed",
				Intros: []string{
					"This email is to confirm that the billing email for " + req.OrgName + " has been changed.",
				},
				ContentBlocks: []template.HTML{
					template.HTML(`<p style="font-size:16px;line-height:24px;margin-bottom:16px;margin-top:16px;text-align:center">` +
						`<strong>Previous email:</strong> ` + req.OldBillingEmail + `<br />` +
						`<strong>New email:</strong> ` + req.NewBillingEmail + `<br />` +
						`<strong>Time of action:</strong> ` + req.ChangedAt.Format("January 2, 2006 at 3:04 PM MST") +
						`</p>`),
				},
				OutrosUnsafe: []template.HTML{
					"If you made this change, no further action is required.",
					template.HTML(`If you did not make this change, please contact our support team immediately at <a href="mailto:` + cfg.SupportEmail + `" style="color:rgb(63,118,255);text-decoration-line:none" target="_blank">` + cfg.SupportEmail + `</a>.`),
				},
			},
		}
	},
}

// AllEmailOperations returns all system email operation registrations for wiring into the builder
func AllEmailOperations() []types.OperationRegistration {
	return []types.OperationRegistration{
		verifyEmail.Registration(),
		welcomeEmail.Registration(),
		inviteEmail.Registration(),
		inviteJoinedEmail.Registration(),
		resetRequestEmail.Registration(),
		resetSuccessEmail.Registration(),
		subscribeEmail.Registration(),
		verifyBillingEmail.Registration(),
		tcNDARequestEmail.Registration(),
		tcNDASignedEmail.Registration(),
		tcAuthEmail.Registration(),
		questionnaireAuthEmail.Registration(),
		billingChangedEmail.Registration(),
	}
}
