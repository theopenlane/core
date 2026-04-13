package email

import (
	"strings"
	"time"

	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/internal/integrations/definitions/email/themes"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// Theme instances shared across email operations
var (
	standardTheme      = themes.StandardTheme{}
	trustCenterTheme   = themes.TrustCenterTheme{}
	questionnaireTheme = themes.QuestionnaireTheme{}
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
	verifyEmailSchema, verifyEmailOp                 = providerkit.OperationSchema[VerifyEmailRequest]()
	welcomeSchema, welcomeOp                         = providerkit.OperationSchema[WelcomeRequest]()
	inviteSchema, inviteOp                           = providerkit.OperationSchema[InviteRequest]()
	inviteJoinedSchema, inviteJoinedOp               = providerkit.OperationSchema[InviteJoinedRequest]()
	resetRequestSchema, resetRequestOp               = providerkit.OperationSchema[PasswordResetEmailRequest]()
	resetSuccessSchema, resetSuccessOp               = providerkit.OperationSchema[PasswordResetSuccessRequest]()
	subscribeSchema, subscribeOp                     = providerkit.OperationSchema[SubscribeRequest]()
	verifyBillingSchema, verifyBillingOp             = providerkit.OperationSchema[VerifyBillingRequest]()
	tcNDARequestSchema, tcNDARequestOp               = providerkit.OperationSchema[TrustCenterNDARequestEmail]()
	tcNDASignedSchema, tcNDASignedOp                 = providerkit.OperationSchema[TrustCenterNDASignedEmail]()
	tcAuthSchema, tcAuthOp                           = providerkit.OperationSchema[TrustCenterAuthEmail]()
	questionnaireAuthSchema, questionnaireAuthOp     = providerkit.OperationSchema[QuestionnaireAuthEmail]()
	billingEmailChangedSchema, billingEmailChangedOp = providerkit.OperationSchema[BillingEmailChangedEmail]()
)

// --- Operation accessors (for emission sites) ---

// VerifyEmailOp returns the operation accessor for verify email sends
func VerifyEmailOp() OperationAccessor {
	return newAccessor(verifyEmailOp.Name())
}

// WelcomeOp returns the operation accessor for welcome email sends
func WelcomeOp() OperationAccessor {
	return newAccessor(welcomeOp.Name())
}

// InviteOp returns the operation accessor for invite email sends
func InviteOp() OperationAccessor {
	return newAccessor(inviteOp.Name())
}

// InviteJoinedOp returns the operation accessor for invite-joined email sends
func InviteJoinedOp() OperationAccessor {
	return newAccessor(inviteJoinedOp.Name())
}

// ResetRequestOp returns the operation accessor for password reset request sends
func ResetRequestOp() OperationAccessor {
	return newAccessor(resetRequestOp.Name())
}

// ResetSuccessOp returns the operation accessor for password reset success sends
func ResetSuccessOp() OperationAccessor {
	return newAccessor(resetSuccessOp.Name())
}

// SubscribeOp returns the operation accessor for subscribe email sends
func SubscribeOp() OperationAccessor {
	return newAccessor(subscribeOp.Name())
}

// VerifyBillingOp returns the operation accessor for verify billing sends
func VerifyBillingOp() OperationAccessor {
	return newAccessor(verifyBillingOp.Name())
}

// TCNDARequestOp returns the operation accessor for trust center NDA request sends
func TCNDARequestOp() OperationAccessor {
	return newAccessor(tcNDARequestOp.Name())
}

// TCNDASignedOp returns the operation accessor for trust center NDA signed sends
func TCNDASignedOp() OperationAccessor {
	return newAccessor(tcNDASignedOp.Name())
}

// TCAuthOp returns the operation accessor for trust center auth sends
func TCAuthOp() OperationAccessor {
	return newAccessor(tcAuthOp.Name())
}

// QuestionnaireAuthOp returns the operation accessor for questionnaire auth sends
func QuestionnaireAuthOp() OperationAccessor {
	return newAccessor(questionnaireAuthOp.Name())
}

// BillingEmailChangedOp returns the operation accessor for billing email changed sends
func BillingEmailChangedOp() OperationAccessor {
	return newAccessor(billingEmailChangedOp.Name())
}

// OperationAccessor provides the operation name and topic for emission sites
type OperationAccessor struct {
	name  string
	topic gala.TopicName
}

// Name returns the operation name
func (a OperationAccessor) Name() string {
	return a.name
}

// Topic returns the gala topic for this operation
func (a OperationAccessor) Topic() gala.TopicName {
	return a.topic
}

// newAccessor creates a new OperationAccessor for the given operation name
func newAccessor(name string) OperationAccessor {
	return OperationAccessor{
		name:  name,
		topic: definitionID.OperationTopic(name),
	}
}

// --- Email operation definitions ---

var verifyEmail = EmailOperation[VerifyEmailRequest]{
	Op: verifyEmailOp, Schema: verifyEmailSchema, Theme: standardTheme,
	Subject: func(cfg RuntimeEmailConfig, _ VerifyEmailRequest) string {
		return "Please verify your email address to login to " + cfg.CompanyName
	},
	Content: func(cfg RuntimeEmailConfig, req VerifyEmailRequest) render.EmailContent {
		verifyURL := cfg.ProductURL + "/verify?token=" + req.Token

		return render.EmailContent{
			Body: render.ContentBody{
				Name:  req.FirstName,
				Title: "Verify your email address",
				Intros: []string{
					"Thank you for registering for the " + cfg.CompanyName + " platform - in order to ensure the security of your account, please verify your email address by clicking the button below, or copy and paste the linked URL into your browser:",
				},
				Actions: []render.Action{{
					Button: render.Button{Text: "Verify Email", Link: verifyURL},
				}},
				Outros: []string{
					verifyURL,
					"If you are having trouble verifying your email address, please contact us at " + cfg.SupportEmail + ".",
				},
			},
		}
	},
}

var welcomeEmail = EmailOperation[WelcomeRequest]{
	Op: welcomeOp, Schema: welcomeSchema, Theme: standardTheme,
	Subject: func(cfg RuntimeEmailConfig, _ WelcomeRequest) string {
		return "Welcome to " + cfg.CompanyName + "!"
	},
	Content: func(cfg RuntimeEmailConfig, req WelcomeRequest) render.EmailContent {
		return render.EmailContent{
			Body: render.ContentBody{
				Name:  req.FirstName,
				Title: "Welcome to " + cfg.CompanyName + "!",
				Intros: []string{
					"Welcome to the " + cfg.CompanyName + " platform - you can now log in to your account.",
					"We've created a personal Organization just for you to help you get started - you can create additional Organizations for your businesses, or just jump right in to see all the amazing features we've cooked up for you.",
					"Check out the starter guide at " + cfg.DocsURL + "/getting-started for more information or our end-to-end examples at " + cfg.DocsURL + "/examples for ideas and inspiration.",
				},
				Actions: []render.Action{{
					Button: render.Button{Text: "Login", Link: cfg.ProductURL},
				}},
				Outros: []string{
					"If you have any questions, please reach out to us at " + cfg.SupportEmail + ".",
				},
			},
		}
	},
}

var inviteEmail = EmailOperation[InviteRequest]{
	Op: inviteOp, Schema: inviteSchema, Theme: standardTheme,
	Subject: func(cfg RuntimeEmailConfig, req InviteRequest) string {
		return "Join Your Teammate " + req.InviterName + " on " + cfg.CompanyName + "!"
	},
	Content: func(cfg RuntimeEmailConfig, req InviteRequest) render.EmailContent {
		inviteURL := cfg.ProductURL + "/invite?token=" + req.Token

		return render.EmailContent{
			Body: render.ContentBody{
				Name:  req.FirstName,
				Title: "Join your team on " + cfg.CompanyName + "!",
				Intros: []string{
					req.InviterName + " has invited you to use " + cfg.CompanyName + " with them, in an Organization called " + req.OrgName + " with role of " + strings.ToUpper(req.Role) + ".",
					"Accept the invitation by clicking on the following button:",
				},
				Actions: []render.Action{{
					Button: render.Button{Text: "Join Now", Link: inviteURL},
				}},
				Outros: []string{
					"Or you can copy and paste the following URL into your browser: " + inviteURL,
					"If you have any questions, please contact " + cfg.SupportEmail + ".",
				},
			},
		}
	},
}

var inviteJoinedEmail = EmailOperation[InviteJoinedRequest]{
	Op: inviteJoinedOp, Schema: inviteJoinedSchema, Theme: standardTheme,
	Subject: func(cfg RuntimeEmailConfig, _ InviteJoinedRequest) string {
		return "You've been added to an Organization on " + cfg.CompanyName
	},
	Content: func(cfg RuntimeEmailConfig, req InviteJoinedRequest) render.EmailContent {
		return render.EmailContent{
			Body: render.ContentBody{
				Name:  req.FirstName,
				Title: "Collaborate with your team on " + cfg.CompanyName,
				Intros: []string{
					"You've been successfully added to an additional Organization " + req.OrgName + ".",
				},
			},
		}
	},
}

var resetRequestEmail = EmailOperation[PasswordResetEmailRequest]{
	Op: resetRequestOp, Schema: resetRequestSchema, Theme: standardTheme,
	Subject: func(cfg RuntimeEmailConfig, _ PasswordResetEmailRequest) string {
		return cfg.CompanyName + " Password Reset - Action Required"
	},
	Content: func(cfg RuntimeEmailConfig, req PasswordResetEmailRequest) render.EmailContent {
		resetURL := cfg.ProductURL + "/password-reset?token=" + req.Token

		return render.EmailContent{
			Body: render.ContentBody{
				Name:  req.FirstName,
				Title: "Reset your password",
				Intros: []string{
					"We received a password reset request for your " + cfg.CompanyName + " account. If you requested a new password, please click on the button below which links to a page where you can securely set a new password.",
				},
				Actions: []render.Action{{
					Button: render.Button{Text: "Reset Password", Link: resetURL},
				}},
				Outros: []string{
					"Or you can copy and paste the following URL into your browser: " + resetURL,
					"For your security, this link will expire after 15 minutes.",
					"If you did not request a new password, please ignore this email and no action is required on your part. If you have concerns, please contact " + cfg.SupportEmail + " to report an issue - the security of your account is important to us.",
				},
			},
		}
	},
}

var resetSuccessEmail = EmailOperation[PasswordResetSuccessRequest]{
	Op: resetSuccessOp, Schema: resetSuccessSchema, Theme: standardTheme,
	Subject: func(cfg RuntimeEmailConfig, _ PasswordResetSuccessRequest) string {
		return cfg.CompanyName + " Password Reset Confirmation"
	},
	Content: func(cfg RuntimeEmailConfig, req PasswordResetSuccessRequest) render.EmailContent {
		return render.EmailContent{
			Body: render.ContentBody{
				Name:  req.FirstName,
				Title: "Your " + cfg.CompanyName + " Password Has Been Reset",
				Intros: []string{
					"Your " + cfg.CompanyName + " password has been successfully reset - no further action is required on your part if you submitted the password reset.",
					"If you did not request a password reset, please contact our Customer Support team immediately at " + cfg.SupportEmail + " - your account security is important to us.",
				},
			},
		}
	},
}

var subscribeEmail = EmailOperation[SubscribeRequest]{
	Op: subscribeOp, Schema: subscribeSchema, Theme: standardTheme,
	Subject: func(cfg RuntimeEmailConfig, _ SubscribeRequest) string {
		return "You've been subscribed to " + cfg.CompanyName
	},
	Content: func(cfg RuntimeEmailConfig, req SubscribeRequest) render.EmailContent {
		verifyURL := cfg.ProductURL + "/subscribe/verify?token=" + req.Token

		return render.EmailContent{
			Body: render.ContentBody{
				Name:  req.FirstName,
				Title: "Thank you for subscribing",
				Intros: []string{
					"Thank you for subscribing to " + req.OrgName + " - in order to confirm the subscription of future emails, please verify your email address by clicking the button below, or copy and paste the linked URL into your browser:",
				},
				Actions: []render.Action{{
					Button: render.Button{Text: "Verify Email", Link: verifyURL},
				}},
				Outros: []string{
					verifyURL,
					"If you are having trouble verifying your email address, please contact us at " + cfg.SupportEmail + ".",
				},
			},
		}
	},
}

var verifyBillingEmail = EmailOperation[VerifyBillingRequest]{
	Op: verifyBillingOp, Schema: verifyBillingSchema, Theme: standardTheme,
	Subject: func(cfg RuntimeEmailConfig, _ VerifyBillingRequest) string {
		return "Please verify the billing email for " + cfg.CompanyName + " to ensure your account stays up to date"
	},
	Content: func(cfg RuntimeEmailConfig, req VerifyBillingRequest) render.EmailContent {
		verifyURL := cfg.ProductURL + "/billing/verify?token=" + req.Token

		return render.EmailContent{
			Body: render.ContentBody{
				Name:  req.FirstName,
				Title: "Verify your billing contact",
				Intros: []string{
					"This email has been sent to you because the billing contact for your " + cfg.CompanyName + " account has changed. In order to ensure the security of your account, please verify your email address by clicking the button below, or copy and paste the linked URL into your browser:",
				},
				Actions: []render.Action{{
					Button: render.Button{Text: "Verify Email", Link: verifyURL},
				}},
				Outros: []string{
					verifyURL,
					"If you are having trouble verifying your email address, please contact us at " + cfg.SupportEmail + ".",
				},
			},
		}
	},
}

var tcNDARequestEmail = EmailOperation[TrustCenterNDARequestEmail]{
	Op: tcNDARequestOp, Schema: tcNDARequestSchema, Theme: trustCenterTheme,
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
					Button: render.Button{Text: "Sign NDA", Link: req.NDAURL},
				}},
				Outros: []string{
					"If the button doesn't work, copy and paste this link into your browser: " + req.NDAURL,
				},
			},
		}
	},
}

var tcNDASignedEmail = EmailOperation[TrustCenterNDASignedEmail]{
	Op: tcNDASignedOp, Schema: tcNDASignedSchema, Theme: trustCenterTheme,
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
					Button: render.Button{Text: "Visit Trust Center", Link: req.TrustCenterURL},
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
	Op: tcAuthOp, Schema: tcAuthSchema, Theme: trustCenterTheme,
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
					Button: render.Button{Text: "Access Trust Center", Link: req.AuthURL},
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
	Op: questionnaireAuthOp, Schema: questionnaireAuthSchema, Theme: questionnaireTheme,
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
					Button: render.Button{Text: "Access Questionnaire", Link: req.AuthURL},
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
	Op: billingEmailChangedOp, Schema: billingEmailChangedSchema, Theme: standardTheme,
	Subject: func(_ RuntimeEmailConfig, req BillingEmailChangedEmail) string {
		return "Billing Email Changed for " + req.OrgName
	},
	Content: func(cfg RuntimeEmailConfig, req BillingEmailChangedEmail) render.EmailContent {
		return render.EmailContent{
			Body: render.ContentBody{
				Title: "Billing Email Changed",
				Intros: []string{
					"This email is to confirm that the billing email for " + req.OrgName + " has been changed.",
				},
				Dictionary: []render.Cell{
					{Key: "Previous email", Value: req.OldBillingEmail},
					{Key: "New email", Value: req.NewBillingEmail},
					{Key: "Time of action", Value: req.ChangedAt.Format("January 2, 2006 at 3:04 PM MST")},
				},
				Outros: []string{
					"If you made this change, no further action is required.",
					"If you did not make this change, please contact our support team immediately at " + cfg.SupportEmail + ".",
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
