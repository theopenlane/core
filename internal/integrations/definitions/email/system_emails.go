package email

import (
	"html"
	"html/template"
	"net/url"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/internal/integrations/definitions/email/themes"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Theme instances shared across email operations
var (
	standardTheme       = themes.Standard
	trustCenterTheme    = themes.TrustCenter
	questionnaireTheme  = themes.Questionnaire
	modernMessageTheme  = themes.ModernMessage
	openlaneModernTheme = themes.OpenlaneModern
)

// Image URLs hosted on the Openlane CDN. The logo URLs use CF Images flexible-variant
// size tags (w=NN,fit=contain) so they render cleanly at the template's slot widths
// without crushing the banner-shaped source image
const (
	iconRocketURL    = "https://www.theopenlane.io/cdn-cgi/imagedelivery/2gi-D0CFOlSOflWJG-LQaA/b5d07352-e391-4ac6-41f8-e1ee9185e000/public"
	iconUserPlusURL  = "https://www.theopenlane.io/cdn-cgi/imagedelivery/2gi-D0CFOlSOflWJG-LQaA/a177b189-bf03-466a-e43a-542585eb1800/public"
	iconUserCheckURL = "https://www.theopenlane.io/cdn-cgi/imagedelivery/2gi-D0CFOlSOflWJG-LQaA/22704d4a-a811-44c0-8618-8309b03dfa00/public"
	iconBellURL      = "https://www.theopenlane.io/cdn-cgi/imagedelivery/2gi-D0CFOlSOflWJG-LQaA/23690f00-2ddb-4d22-e9a5-af470c93c100/public"
	iconMarkURL      = "https://www.theopenlane.io/cdn-cgi/imagedelivery/2gi-D0CFOlSOflWJG-LQaA/13c70376-9501-4abd-a577-41a2b843a000/public"
	iconBWMarkURL    = "https://www.theopenlane.io/cdn-cgi/imagedelivery/2gi-D0CFOlSOflWJG-LQaA/2c681874-6b4a-41de-73ea-58e550a95800/public"
)

// Trust center / questionnaire button colors matching the original teal theme
const (
	tcButtonColor     = "#3fc2b4"
	tcButtonTextColor = "#ffffff"
)

// tokenURL constructs a product URL with a query-encoded token parameter
func tokenURL(base, path, token string) string {
	return base + path + "?token=" + url.QueryEscape(token)
}

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

// TrustCenterNDARequestEmail is the input for the trust center NDA signing request.
// Callers may pass either the pre-built NDAURL or the RequestID + TrustCenterID pair;
// when NDAURL is empty the operation resolves it from RequestID and TrustCenterID.
// OrgName is resolved from TrustCenterID when empty
type TrustCenterNDARequestEmail struct {
	RecipientInfo
	// OrgName is the organization requesting the NDA; resolved from TrustCenterID when empty
	OrgName string `json:"org_name,omitempty" jsonschema:"description=Organization name"`
	// RequestID is the NDA request identifier used for JWT subject construction
	RequestID string `json:"requestId,omitempty" jsonschema:"description=NDA request ID for token generation"`
	// TrustCenterID is the trust center identifier used for URL construction
	TrustCenterID string `json:"trustCenterId,omitempty" jsonschema:"description=Trust center ID for URL construction"`
	// NDAURL is the direct URL to the NDA signing page; when empty the operation constructs it from RequestID and TrustCenterID
	NDAURL string `json:"ndaUrl,omitempty" jsonschema:"description=NDA signing URL"`
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

// TrustCenterAuthEmail is the input for trust center access authentication.
// Callers may pass either the pre-built AuthURL or the RequestID + TrustCenterID pair;
// when AuthURL is empty the operation resolves it from RequestID and TrustCenterID.
// OrgName is resolved from TrustCenterID when empty
type TrustCenterAuthEmail struct {
	RecipientInfo
	// OrgName is the organization whose trust center is being accessed; resolved from TrustCenterID when empty
	OrgName string `json:"org_name,omitempty" jsonschema:"description=Organization name"`
	// RequestID is the NDA request identifier used for JWT subject construction
	RequestID string `json:"requestId,omitempty" jsonschema:"description=NDA request ID for token generation"`
	// TrustCenterID is the trust center identifier used for URL construction
	TrustCenterID string `json:"trustCenterId,omitempty" jsonschema:"description=Trust center ID for URL construction"`
	// AuthURL is the authentication URL for trust center access; when empty the operation constructs it from RequestID and TrustCenterID
	AuthURL string `json:"authUrl,omitempty" jsonschema:"description=Trust center auth URL"`
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
	Op: VerifyEmailOp, Schema: verifyEmailSchema, Theme: modernMessageTheme,
	Description: "System email prompting a new user to verify their email address",
	Subject: func(cfg RuntimeEmailConfig, _ VerifyEmailRequest) string {
		return "Please verify your email address to login to " + cfg.CompanyName
	},
	Build: func(cfg RuntimeEmailConfig, req VerifyEmailRequest) render.ContentBody {
		verifyURL := tokenURL(cfg.ProductURL, "/verify", req.Token)

		return render.ContentBody{
			Preheader: "Verify Your Email Address",
			Header:    render.HeaderBlock{Logo: &render.ContentIcon{Src: iconMarkURL, Alt: cfg.CompanyName}},
			Name:      req.FirstName,
			Title:     "Verify Your Email Address",
			Intros: render.IntrosBlock{
				Paragraphs: []string{
					"Welcome to " + cfg.CompanyName + " - where compliance isn't just a checkbox.",
					"Before you get started, let's make sure it's really you. Click below to verify your email:",
				},
			},
			Actions: []render.Action{{
				Button: render.Button{Text: "Confirm Email", Link: verifyURL},
			}},
			Outros: render.OutrosBlock{
				Paragraphs: []string{
					"This link expires in 7 days - but don't worry, if it does, you'll get a fresh one when you try to verify later.",
				},
			},
		}
	},
}

var welcomeEmail = EmailOperation[WelcomeRequest]{
	Op: WelcomeOp, Schema: welcomeSchema, Theme: modernMessageTheme,
	Description: "System welcome email delivered after account signup",
	Subject: func(cfg RuntimeEmailConfig, _ WelcomeRequest) string {
		return "Welcome to " + cfg.CompanyName + "!"
	},
	Build: func(cfg RuntimeEmailConfig, req WelcomeRequest) render.ContentBody {
		return render.ContentBody{
			Preheader: "Welcome to " + cfg.CompanyName + "!",
			Header:    render.HeaderBlock{Logo: &render.ContentIcon{Src: iconMarkURL, Alt: cfg.CompanyName}},
			Icon:      &render.ContentIcon{Src: iconRocketURL, Alt: "Rocket"},
			Name:      req.FirstName,
			Title:     "Welcome to " + cfg.CompanyName + "!",
			Intros: render.IntrosBlock{
				Paragraphs: []string{
					"We're thrilled to have you here. At " + cfg.CompanyName + ", we're working to develop a cutting-edge cybersecurity and compliance automation solution to help organizations of all sizes and industries secure their systems, navigate the increasingly complex web of privacy laws and regulations, ensure continuous compliance, manage risks, and get ahead of evolving cyber threats.",
				},
			},
			Callout: &render.Callout{
				Title: "Here's how to get started:",
				Items: []template.HTML{
					"Go through our onboarding process to get a personalized experience",
					"Setup a test program to see all the features our platform has to offer",
					"Checkout our " + render.Link(cfg.DocsURL+"/docs/platform/quickstartguide", "quickstart guide") + " for helpful resources",
				},
			},
			Actions: []render.Action{{
				Button: render.Button{Text: "Get Started", Link: cfg.ProductURL},
			}},
		}
	},
}

var inviteEmail = EmailOperation[InviteRequest]{
	Op: InviteOp, Schema: inviteSchema, Theme: modernMessageTheme,
	Description: "System email inviting a user to join an organization",
	Subject: func(cfg RuntimeEmailConfig, req InviteRequest) string {
		return "Join Your Teammate " + req.InviterName + " on " + cfg.CompanyName + "!"
	},
	Build: func(cfg RuntimeEmailConfig, req InviteRequest) render.ContentBody {
		inviteURL := tokenURL(cfg.ProductURL, "/invite", req.Token)
		inviteIntro := template.HTML("You're in - let's build trust without the busywork. "+html.EscapeString(req.InviterName)+" has invited you to collaborate in "+html.EscapeString(cfg.CompanyName)+", as part of the ") +
			render.Bold(req.OrgName) + " organization"
		if req.Role != "" {
			inviteIntro += template.HTML(" with the role of " + html.EscapeString(strings.ToUpper(req.Role)))
		}
		inviteIntro += "."

		return render.ContentBody{
			Preheader: "You've been invited to join " + cfg.CompanyName,
			Header:    render.HeaderBlock{Logo: &render.ContentIcon{Src: iconMarkURL, Alt: cfg.CompanyName}},
			Icon:      &render.ContentIcon{Src: iconUserPlusURL, Alt: "User-Plus"},
			Name:      req.FirstName,
			Title:     "You've been invited to join " + cfg.CompanyName + "!",
			Intros: render.IntrosBlock{
				Unsafe: []template.HTML{
					inviteIntro,
					"To get started (and verify your email), click the link below:",
				},
			},
			Actions: []render.Action{{
				Button: render.Button{Text: "Accept Invite", Link: inviteURL},
			}},
			Outros: render.OutrosBlock{
				Paragraphs: []string{
					"This link expires in 7 days - but don't worry, if it does, you'll get a fresh one when you try to verify later.",
				},
			},
		}
	},
}

var inviteJoinedEmail = EmailOperation[InviteJoinedRequest]{
	Op: InviteJoinedOp, Schema: inviteJoinedSchema, Theme: modernMessageTheme,
	Description: "System notification confirming an invited user has joined an organization",
	Subject: func(cfg RuntimeEmailConfig, _ InviteJoinedRequest) string {
		return "You've been added to an Organization on " + cfg.CompanyName
	},
	Build: func(cfg RuntimeEmailConfig, req InviteJoinedRequest) render.ContentBody {
		return render.ContentBody{
			Preheader: "You're in - welcome to " + req.OrgName + " on " + cfg.CompanyName + "!",
			Header:    render.HeaderBlock{Logo: &render.ContentIcon{Src: iconMarkURL, Alt: cfg.CompanyName}},
			Icon:      &render.ContentIcon{Src: iconUserCheckURL, Alt: "User-Check"},
			Name:      req.FirstName,
			Title:     "You're in - welcome to " + req.OrgName + " on " + cfg.CompanyName + "!",
			Intros: render.IntrosBlock{
				Paragraphs: []string{
					"Welcome to " + cfg.CompanyName + " - we're excited to have you on board with " + req.OrgName + "!",
					"Ditch the spreadsheets, and embrace the pipeline with a faster, cleaner way to automate compliance, manage risk, and stay ahead of security threats - without the manual overhead.",
				},
			},
			Callout: &render.Callout{
				Title:   "Here's how to get started:",
				Ordered: true,
				Items: []template.HTML{
					"Complete the onboarding flow to tailor your experience",
					"Explore any active " + render.Link(cfg.ProductURL+"/programs", "programs") + " your team is running",
					"Checkout our " + render.Link(cfg.DocsURL, "quickstart guide") + " for helpful resources",
				},
			},
			Outros: render.OutrosBlock{
				Paragraphs: []string{
					"When you're ready, hop into the platform and start exploring:",
				},
			},
			Actions: []render.Action{{
				Button: render.Button{Text: "Get Started", Link: cfg.ProductURL},
			}},
		}
	},
}

var resetRequestEmail = EmailOperation[PasswordResetEmailRequest]{
	Op: ResetRequestOp, Schema: resetRequestSchema, Theme: modernMessageTheme,
	Description: "System email delivering a password reset link to a user",
	Subject: func(cfg RuntimeEmailConfig, _ PasswordResetEmailRequest) string {
		return cfg.CompanyName + " Password Reset - Action Required"
	},
	Build: func(cfg RuntimeEmailConfig, req PasswordResetEmailRequest) render.ContentBody {
		resetURL := tokenURL(cfg.ProductURL, "/password-reset", req.Token)

		return render.ContentBody{
			Preheader: "Reset your " + cfg.CompanyName + " password",
			Header:    render.HeaderBlock{Logo: &render.ContentIcon{Src: iconMarkURL, Alt: cfg.CompanyName}},
			Name:      req.FirstName,
			Title:     "Reset your " + cfg.CompanyName + " password",
			Intros: render.IntrosBlock{
				Paragraphs: []string{
					"We received a request to reset your " + cfg.CompanyName + " password.",
					"If that was you, no problem - just click the button below to set a new one:",
				},
			},
			Actions: []render.Action{{
				Button: render.Button{Text: "Password Reset", Link: resetURL},
			}},
			Outros: render.OutrosBlock{
				Paragraphs: []string{
					"This link will expire in 15 minutes to keep things secure.",
					"If you didn't request a password reset, you can safely ignore this email.",
				},
			},
		}
	},
}

var resetSuccessEmail = EmailOperation[PasswordResetSuccessRequest]{
	Op: ResetSuccessOp, Schema: resetSuccessSchema, Theme: modernMessageTheme,
	Description: "System email confirming a successful password reset",
	Subject: func(cfg RuntimeEmailConfig, _ PasswordResetSuccessRequest) string {
		return cfg.CompanyName + " Password Reset Confirmation"
	},
	Build: func(cfg RuntimeEmailConfig, req PasswordResetSuccessRequest) render.ContentBody {
		return render.ContentBody{
			Preheader: "Your " + cfg.CompanyName + " password has been reset",
			Header:    render.HeaderBlock{Logo: &render.ContentIcon{Src: iconMarkURL, Alt: cfg.CompanyName}},
			Name:      req.FirstName,
			Title:     "Your " + cfg.CompanyName + " password has been reset",
			Intros: render.IntrosBlock{
				Paragraphs: []string{
					"Your password was successfully updated. If you made this change, you're all set - no further action needed.",
					"If you didn't request a password reset, please contact our support team right away at " + cfg.SupportEmail + ". Keeping your account secure is our top priority.",
				},
			},
		}
	},
}

var subscribeEmail = EmailOperation[SubscribeRequest]{
	Op: SubscribeOp, Schema: subscribeSchema, Theme: modernMessageTheme,
	Description: "System email confirming a subscriber's early access signup",
	Subject: func(cfg RuntimeEmailConfig, _ SubscribeRequest) string {
		return "You've been subscribed to " + cfg.CompanyName
	},
	Build: func(cfg RuntimeEmailConfig, req SubscribeRequest) render.ContentBody {
		verifyURL := tokenURL(cfg.ProductURL, "/subscribe/verify", req.Token)

		return render.ContentBody{
			Preheader: "You're In - Early Access Secured! Thanks for your interest in our beta program.",
			Header:    render.HeaderBlock{Logo: &render.ContentIcon{Src: iconMarkURL, Alt: cfg.CompanyName}},
			Icon:      &render.ContentIcon{Src: iconBellURL, Alt: "Notification Bell"},
			Name:      req.FirstName,
			Title:     "You're In - Early Access Secured!",
			Intros: render.IntrosBlock{
				Paragraphs: []string{
					"We're thrilled to have you as part of our early community. Your interest means the world to us as we work to build a cutting-edge solution. We can't wait to share it with you!",
					"Please confirm your email address to ensure you receive all important updates about your early access.",
				},
			},
			Actions: []render.Action{{
				Button: render.Button{Text: "Confirm Email", Link: verifyURL},
			}},
			Callout: &render.Callout{
				Title: "What to Expect Next:",
				Items: []template.HTML{
					"You'll hear from us soon - We'll email you as soon as your spot is ready.",
					"Early access to beta features - Get a first look at everything we're building.",
					"Help shape the future - Your feedback will directly influence the product.",
				},
			},
			Outros: render.OutrosBlock{
				Paragraphs: []string{
					"Thank you for being part of this journey - we're excited to have you on board!",
				},
			},
		}
	},
}

var verifyBillingEmail = EmailOperation[VerifyBillingRequest]{
	Op: VerifyBillingOp, Schema: verifyBillingSchema, Theme: modernMessageTheme,
	Description: "System email prompting verification of the billing email on file",
	Subject: func(cfg RuntimeEmailConfig, _ VerifyBillingRequest) string {
		return "Please verify the billing email for " + cfg.CompanyName + " to ensure your account stays up to date"
	},
	Build: func(cfg RuntimeEmailConfig, req VerifyBillingRequest) render.ContentBody {
		verifyURL := tokenURL(cfg.ProductURL, "/billing/verify", req.Token)

		return render.ContentBody{
			Preheader: "Verify Your Email Address",
			Header:    render.HeaderBlock{Logo: &render.ContentIcon{Src: iconMarkURL, Alt: cfg.CompanyName}},
			Name:      req.FirstName,
			Title:     "Verify Your Billing Email Address",
			Intros: render.IntrosBlock{
				Paragraphs: []string{
					"You're receiving this because the billing contact for your " + cfg.CompanyName + " account was just updated.",
					"To help keep your account secure, please verify your email address by clicking the button below:",
				},
			},
			Actions: []render.Action{{
				Button: render.Button{Text: "Confirm Billing Email", Link: verifyURL},
			}},
			Outros: render.OutrosBlock{
				Paragraphs: []string{
					"If you run into any issues, feel free to reach out at " + cfg.SupportEmail + ".",
				},
			},
		}
	},
}

var tcNDARequestEmail = EmailOperation[TrustCenterNDARequestEmail]{
	Op: TCNDARequestOp, Schema: tcNDARequestSchema, Theme: modernMessageTheme,
	Description: "System email requesting an NDA signature before granting trust center access",
	PreHook:     resolveTrustCenterNDARequestFields,
	Subject: func(_ RuntimeEmailConfig, req TrustCenterNDARequestEmail) string {
		return req.OrgName + " Trust Center NDA Request"
	},
	Build: func(cfg RuntimeEmailConfig, req TrustCenterNDARequestEmail) render.ContentBody {
		return render.ContentBody{
			Header: render.HeaderBlock{Logo: &render.ContentIcon{Src: iconMarkURL, Alt: cfg.CompanyName}},
			Title:  "You requested access to " + req.OrgName + "'s Trust Center",
			Intros: render.IntrosBlock{
				Paragraphs: []string{
					"To continue, please review and sign the Non-Disclosure Agreement (NDA). Once signed, you'll be granted access to protected Trust Center documents.",
				},
			},
			Actions: []render.Action{{
				Button: render.Button{Text: "Sign NDA", Link: req.NDAURL, Color: tcButtonColor, TextColor: tcButtonTextColor},
			}},
		}
	},
}

var tcNDASignedEmail = EmailOperation[TrustCenterNDASignedEmail]{
	Op: TCNDASignedOp, Schema: tcNDASignedSchema, Theme: modernMessageTheme,
	Description: "System email confirming a signed NDA and attaching the signed copy",
	Subject: func(_ RuntimeEmailConfig, req TrustCenterNDASignedEmail) string {
		return req.OrgName + " Trust Center NDA Signed"
	},
	Build: func(cfg RuntimeEmailConfig, req TrustCenterNDASignedEmail) render.ContentBody {
		return render.ContentBody{
			Header: render.HeaderBlock{Logo: &render.ContentIcon{Src: iconMarkURL, Alt: cfg.CompanyName}},
			Title:  "Your NDA with " + req.OrgName + " has been signed",
			Intros: render.IntrosBlock{
				Paragraphs: []string{
					"Thank you for signing the Non-Disclosure Agreement (NDA). You now have access to " + req.OrgName + "'s protected Trust Center documents.",
				},
			},
			Actions: []render.Action{{
				Button: render.Button{Text: "Visit Trust Center", Link: req.TrustCenterURL, Color: tcButtonColor, TextColor: tcButtonTextColor},
			}},
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
	Op: TCAuthOp, Schema: tcAuthSchema, Theme: modernMessageTheme,
	Description: "System email delivering a time-limited authentication link to a trust center",
	PreHook:     resolveTrustCenterAuthFields,
	Subject: func(_ RuntimeEmailConfig, req TrustCenterAuthEmail) string {
		return "Access " + req.OrgName + "'s Trust Center"
	},
	Build: func(cfg RuntimeEmailConfig, req TrustCenterAuthEmail) render.ContentBody {
		return render.ContentBody{
			Header: render.HeaderBlock{Logo: &render.ContentIcon{Src: iconMarkURL, Alt: cfg.CompanyName}},
			Title:  "Access " + req.OrgName + "'s Trust Center",
			Intros: render.IntrosBlock{
				Paragraphs: []string{
					"You've been granted access to " + req.OrgName + "'s Trust Center. Click the button below to authenticate and view the available resources.",
				},
			},
			Actions: []render.Action{{
				Button: render.Button{Text: "Access Trust Center", Link: req.AuthURL, Color: tcButtonColor, TextColor: tcButtonTextColor},
			}},
			Outros: render.OutrosBlock{
				Paragraphs: []string{
					"This authentication link provides secure, time-limited access and will expire after a short period for your security.",
				},
			},
		}
	},
}

var questionnaireAuthEmail = EmailOperation[QuestionnaireAuthEmail]{
	Op: QuestionnaireAuthOp, Schema: questionnaireAuthSchema, Theme: modernMessageTheme,
	Description: "System email delivering a time-limited authentication link to a questionnaire",
	Subject: func(cfg RuntimeEmailConfig, req QuestionnaireAuthEmail) string {
		return "Access " + req.AssessmentName + " Questionnaire from " + cfg.CompanyName
	},
	Build: func(cfg RuntimeEmailConfig, req QuestionnaireAuthEmail) render.ContentBody {
		return render.ContentBody{
			Header: render.HeaderBlock{Logo: &render.ContentIcon{Src: iconMarkURL, Alt: cfg.CompanyName}},
			Title:  cfg.CompanyName + " sent you an assessment to complete",
			Intros: render.IntrosBlock{
				Paragraphs: []string{
					cfg.CompanyName + " has shared a form (" + req.AssessmentName + ") for you to complete. Click the button below to access it.",
				},
			},
			Actions: []render.Action{{
				Button: render.Button{Text: "Access Questionnaire", Link: req.AuthURL, Color: tcButtonColor, TextColor: tcButtonTextColor},
			}},
			Outros: render.OutrosBlock{
				Paragraphs: []string{
					"This authentication link provides secure, time-limited access and will expire after a short period for your security.",
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
	Op: BillingEmailChangedOp, Schema: billingEmailChangedSchema, Theme: modernMessageTheme,
	Description: "System notification confirming a change to the billing email on file",
	Subject: func(_ RuntimeEmailConfig, req BillingEmailChangedEmail) string {
		return "Billing Email Changed for " + req.OrgName
	},
	Build: func(cfg RuntimeEmailConfig, req BillingEmailChangedEmail) render.ContentBody {
		return render.ContentBody{
			Preheader: "Billing Email Changed",
			Header:    render.HeaderBlock{Logo: &render.ContentIcon{Src: iconMarkURL, Alt: cfg.CompanyName}},
			Title:     "Billing Email Changed",
			Intros: render.IntrosBlock{
				Paragraphs: []string{
					"This email is to confirm that the billing email for " + req.OrgName + " has been changed.",
				},
			},
			Dictionary: render.Dictionary{
				Cells: []render.Cell{
					{Key: "Previous email", Value: req.OldBillingEmail},
					{Key: "New email", Value: req.NewBillingEmail},
					{Key: "Time of action", Value: req.ChangedAt.Format("January 2, 2006 at 3:04 PM MST")},
				},
			},
			Outros: render.OutrosBlock{
				Unsafe: []template.HTML{
					"If you made this change, no further action is required.",
					template.HTML(`If you did not make this change, please contact our support team immediately at <a href="mailto:` + cfg.SupportEmail + `" style="color:rgb(63,118,255);text-decoration-line:none" target="_blank">` + cfg.SupportEmail + `</a>.`),
				},
			},
		}
	},
}

var dispatchers = []EmailDispatcher{
	verifyEmail,
	welcomeEmail,
	inviteEmail,
	inviteJoinedEmail,
	resetRequestEmail,
	resetSuccessEmail,
	subscribeEmail,
	verifyBillingEmail,
	tcNDARequestEmail,
	tcNDASignedEmail,
	tcAuthEmail,
	questionnaireAuthEmail,
	billingChangedEmail,
	brandedMessageEmail,
}

var dispatcherIndex = lo.SliceToMap(dispatchers, func(d EmailDispatcher) (string, EmailDispatcher) {
	return d.Name(), d
})

// AllEmailOperations returns all system email operation registrations for wiring into the builder
func AllEmailOperations() []types.OperationRegistration {
	return lo.Map(dispatchers, func(d EmailDispatcher, _ int) types.OperationRegistration {
		return d.Registration()
	})
}

// CustomerSelectableDispatchers returns the dispatchers marked as customer-selectable
func CustomerSelectableDispatchers() []EmailDispatcher {
	return lo.Filter(dispatchers, func(d EmailDispatcher, _ int) bool {
		return d.Registration().CustomerSelectable
	})
}

// DispatcherByKey resolves a registered email dispatcher by its catalog key
func DispatcherByKey(key string) (EmailDispatcher, bool) {
	d, ok := dispatcherIndex[key]
	return d, ok
}
