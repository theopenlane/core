package models

import (
	"mime/multipart"
	"net/textproto"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"

	"github.com/theopenlane/utils/rout"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/utils/passwd"
)

// =========
// Auth Data
// =========

type AuthData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Session      string `json:"session,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
}

// =========
// LOGIN
// =========

// LoginRequest holds the login payload for the /login route
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	OTPCode  string `json:"otp_code,omitempty"`
}

// LoginReply holds the response to LoginRequest
type LoginReply struct {
	rout.Reply
	AuthData
	Message string `json:"message"`
}

// Validate ensures the required fields are set on the LoginRequest request
func (r *LoginRequest) Validate() error {
	r.Username = strings.TrimSpace(r.Username)
	r.Password = strings.TrimSpace(r.Password)

	switch {
	case r.Username == "":
		return rout.NewMissingRequiredFieldError("username")
	case r.Password == "":
		return rout.NewMissingRequiredFieldError("password")
	}

	return nil
}

// ExampleLoginSuccessRequest is an example of a successful login request for OpenAPI documentation
var ExampleLoginSuccessRequest = LoginRequest{
	Username: "sfunky@theopenlane.io",
	Password: "mitb!",
	OTPCode:  "123456",
}

// ExampleLoginSuccessResponse is an example of a successful login response for OpenAPI documentation
var ExampleLoginSuccessResponse = LoginReply{
	Reply: rout.Reply{
		Success: true,
	},
	AuthData: AuthData{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		Session:      "session",
		TokenType:    "bearer",
	},
}

// =========
// REFRESH
// =========

// RefreshRequest holds the fields that should be included on a request to the `/refresh` endpoint
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshReply holds the fields that are sent on a response to the `/refresh` endpoint
type RefreshReply struct {
	rout.Reply
	Message string `json:"message,omitempty"`
	AuthData
}

// Validate ensures the required fields are set on the RefreshRequest request
func (r *RefreshRequest) Validate() error {
	if r.RefreshToken == "" {
		return rout.NewMissingRequiredFieldError("refresh_token")
	}

	return nil
}

// ExampleRefreshRequest is an example of a successful refresh request for OpenAPI documentation
var ExampleRefreshRequest = RefreshRequest{
	RefreshToken: "token",
}

// ExampleRefreshSuccessResponse is an example of a successful refresh response for OpenAPI documentation
var ExampleRefreshSuccessResponse = RefreshReply{
	Reply:   rout.Reply{Success: true},
	Message: "success",
	AuthData: AuthData{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		Session:      "session",
		TokenType:    "bearer",
	},
}

// =========
// REGISTER
// =========

// RegisterRequest holds the fields that should be included on a request to the `/register` endpoint
type RegisterRequest struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// RegisterReply holds the fields that are sent on a response to the `/register` endpoint
type RegisterReply struct {
	rout.Reply
	ID      string `json:"user_id"`
	Email   string `json:"email"`
	Message string `json:"message"`
	Token   string `json:"token"`
}

// Validate ensures the required fields are set on the RegisterRequest request
func (r *RegisterRequest) Validate() error {
	r.FirstName = strings.TrimSpace(r.FirstName)
	r.LastName = strings.TrimSpace(r.LastName)
	r.Email = strings.TrimSpace(r.Email)
	r.Password = strings.TrimSpace(r.Password)

	// Required for all requests
	switch {
	case r.Email == "":
		return rout.MissingField("email")
	case r.Password == "":
		return rout.MissingField("password")
	case passwd.Strength(r.Password) < passwd.Moderate:
		return rout.ErrPasswordTooWeak
	}

	return nil
}

// ExampleRegisterSuccessRequest is an example of a successful register request for OpenAPI documentation
var ExampleRegisterSuccessRequest = RegisterRequest{
	FirstName: "Sarah",
	LastName:  "Funk",
	Email:     "sfunky@theopenlane.io",
	Password:  "mitb!",
}

// ExampleRegisterSuccessResponse is an example of a successful register response for OpenAPI documentation
var ExampleRegisterSuccessResponse = RegisterReply{
	Reply:   rout.Reply{Success: true},
	ID:      "1234",
	Email:   "",
	Message: "Welcome to Openlane!",
	Token:   "",
}

// =========
// SWITCH ORGANIZATION
// =========

// SwitchOrganizationRequest contains the target organization ID being switched to for the /switch endpoint
type SwitchOrganizationRequest struct {
	TargetOrganizationID string `json:"target_organization_id"`
}

// SwitchOrganizationReply holds the new authentication and session information for the user for the new organization
type SwitchOrganizationReply struct {
	rout.Reply
	AuthData
}

// Validate ensures the required fields are set on the SwitchOrganizationRequest request
func (r *SwitchOrganizationRequest) Validate() error {
	if r.TargetOrganizationID == "" {
		return rout.NewMissingRequiredFieldError("target_organization_id")
	}

	return nil
}

// ExampleSwitchSuccessRequest is an example of a successful switch organization request for OpenAPI documentation
var ExampleSwitchSuccessRequest = SwitchOrganizationRequest{
	TargetOrganizationID: ulids.New().String(),
}

// ExampleSwitchSuccessReply is an example of a successful switch organization response for OpenAPI documentation
var ExampleSwitchSuccessReply = SwitchOrganizationReply{
	Reply: rout.Reply{
		Success: true,
	},
	AuthData: AuthData{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		Session:      "session",
		TokenType:    "bearer",
	},
}

// =========
// VERIFY EMAIL
// =========

// VerifyRequest holds the fields that should be included on a request to the `/verify` endpoint
type VerifyRequest struct {
	Token string `query:"token"`
}

// VerifyReply holds the fields that are sent on a response to the `/verify` endpoint
type VerifyReply struct {
	rout.Reply
	ID      string `json:"user_id"`
	Email   string `json:"email"`
	Token   string `json:"token"`
	Message string `json:"message,omitempty"`
	AuthData
}

// Validate ensures the required fields are set on the VerifyRequest request
func (r *VerifyRequest) Validate() error {
	if r.Token == "" {
		return rout.NewMissingRequiredFieldError("token")
	}

	return nil
}

// ExampleVerifySuccessRequest is an example of a successful verify request for OpenAPI documentation
var ExampleVerifySuccessRequest = VerifyRequest{
	Token: "token",
}

// ExampleVerifySuccessResponse is an example of a successful verify response for OpenAPI documentation
var ExampleVerifySuccessResponse = VerifyReply{
	Reply: rout.Reply{
		Success: true,
	},
	ID:      ulids.New().String(),
	Email:   "gregor.clegane@theopenlane.io",
	Token:   "token",
	Message: "Email has been verified",
	AuthData: AuthData{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		Session:      "session",
		TokenType:    "bearer",
	},
}

// =========
// RESEND EMAIL
// =========

// ResendRequest contains fields for a resend email verification request to the `/resend` endpoint
type ResendRequest struct {
	Email string `json:"email"`
}

// ResendReply holds the fields that are sent on a response to the `/resend` endpoint
type ResendReply struct {
	rout.Reply
	Message string `json:"message"`
}

// Validate ensures the required fields are set on the ResendRequest request
func (r *ResendRequest) Validate() error {
	if r.Email == "" {
		return rout.NewMissingRequiredFieldError("email")
	}

	return nil
}

// ExampleResendEmailSuccessRequest is an example of a successful resend email request for OpenAPI documentation
var ExampleResendEmailSuccessRequest = ResendRequest{
	Email: "cercei.lannister@theopenlane.io",
}

// ExampleResendEmailSuccessResponse is an example of a successful resend email response for OpenAPI documentation
var ExampleResendEmailSuccessResponse = ResendReply{
	Reply: rout.Reply{
		Success: true,
	},
	Message: "Email has been resent",
}

// =========
// FORGOT PASSWORD
// =========

// ForgotPasswordRequest contains fields for a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ForgotPasswordReply contains fields for a forgot password response
type ForgotPasswordReply struct {
	rout.Reply
	Message string `json:"message,omitempty"`
}

// Validate ensures the required fields are set on the ForgotPasswordRequest request
func (r *ForgotPasswordRequest) Validate() error {
	if r.Email == "" {
		return rout.NewMissingRequiredFieldError("email")
	}

	return nil
}

// ExampleForgotPasswordSuccessRequest is an example of a successful forgot password request for OpenAPI documentation
var ExampleForgotPasswordSuccessRequest = ForgotPasswordRequest{
	Email: "example@theopenlane.io",
}

// ExampleForgotPasswordSuccessResponse is an example of a successful forgot password response for OpenAPI documentation
var ExampleForgotPasswordSuccessResponse = ForgotPasswordReply{
	Reply: rout.Reply{
		Success: true,
	},
	Message: "We've received your request to have the password associated with this email reset. Please check your email.",
}

// =========
// RESET PASSWORD
// =========

// ResetPasswordRequest contains user input required to reset a user's password using /password-reset endpoint
type ResetPasswordRequest struct {
	Password string `json:"password"`
	Token    string `json:"token"`
}

// ResetPasswordReply is the response returned from a non-successful password reset request
// on success, no content is returned (204)
type ResetPasswordReply struct {
	rout.Reply
	Message string `json:"message"`
}

// Validate ensures the required fields are set on the ResetPasswordRequest request
func (r *ResetPasswordRequest) Validate() error {
	r.Password = strings.TrimSpace(r.Password)

	switch {
	case r.Token == "":
		return rout.NewMissingRequiredFieldError("token")
	case r.Password == "":
		return rout.NewMissingRequiredFieldError("password")
	case passwd.Strength(r.Password) < passwd.Moderate:
		return rout.ErrPasswordTooWeak
	}

	return nil
}

// ExampleResetPasswordSuccessRequest is an example of a successful reset password request for OpenAPI documentation
var ExampleResetPasswordSuccessRequest = ResetPasswordRequest{
	Password: "mitb!",
	Token:    "token",
}

// ExampleResetPasswordSuccessResponse is an example of a successful reset password response for OpenAPI documentation
var ExampleResetPasswordSuccessResponse = ResetPasswordReply{
	Reply: rout.Reply{
		Success: true,
	},
	Message: "Password has been reset",
}

// =========
// WEBAUTHN
// =========

// WebauthnRegistrationRequest is the request to begin a webauthn login
type WebauthnRegistrationRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// WebauthnRegistrationResponse is the response to begin a webauthn login
// this includes the credential creation options and the session token
type WebauthnBeginRegistrationResponse struct {
	Reply rout.Reply
	*protocol.CredentialCreation
	Session string `json:"session,omitempty"`
}

// WebauthnRegistrationResponse is the response after a successful webauthn registration
type WebauthnRegistrationResponse struct {
	rout.Reply
	Message   string `json:"message,omitempty"`
	TokenType string `json:"token_type"`
	AuthData
}

// WebauthnBeginLoginResponse is the response to begin a webauthn login
// this includes the credential assertion options and the session token
type WebauthnBeginLoginResponse struct {
	Reply rout.Reply
	*protocol.CredentialAssertion
	Session string `json:"session,omitempty"`
}

// WebauthnRegistrationResponse is the response after a successful webauthn login
type WebauthnLoginResponse struct {
	rout.Reply
	Message string `json:"message,omitempty"`
	AuthData
}

// =========
// SUBSCRIBER VERIFY
// =========

// VerifySubscribeRequest holds the fields that should be included on a request to the `/subscribe/verify` endpoint
type VerifySubscribeRequest struct {
	Token string `query:"token"`
}

// VerifySubscribeReply holds the fields that are sent on a response to the `/subscribe/verify` endpoint
type VerifySubscribeReply struct {
	rout.Reply
	Message string `json:"message,omitempty"`
}

// Validate ensures the required fields are set on the VerifySubscribeRequest request
func (r *VerifySubscribeRequest) Validate() error {
	if r.Token == "" {
		return rout.NewMissingRequiredFieldError("token")
	}

	return nil
}

// ExampleVerifySubscriptionSuccessRequest is an example of a successful verify subscription request for OpenAPI documentation
var ExampleVerifySubscriptionSuccessRequest = VerifySubscribeRequest{
	Token: "token",
}

// ExampleVerifySubscriptionResponse is an example of a successful verify subscription response for OpenAPI documentation
var ExampleVerifySubscriptionResponse = VerifySubscribeReply{
	Reply:   rout.Reply{Success: true},
	Message: "Subscription confirmed, looking forward to sending you updates!",
}

// =========
// PUBLISH EVENT
// =========

// PublishRequest is the request payload for the event publisher
type PublishRequest struct {
	Tags    map[string]string `json:"tags"`
	Topic   string            `json:"topic"`
	Message string            `json:"message"`
}

// PublishReply holds the fields that are sent on a response to the `/event/publish` endpoint
type PublishReply struct {
	rout.Reply
	Message string `json:"message"`
}

// Validate ensures the required fields are set on the PublishRequest request
func (r *PublishRequest) Validate() error {
	switch {
	case r.Message == "":
		return rout.NewMissingRequiredFieldError("message")
	case r.Tags == nil:
		return rout.NewMissingRequiredFieldError("tags")
	case len(r.Tags) == 0:
		return rout.NewMissingRequiredFieldError("tags")
	case r.Topic == "":
		return rout.NewMissingRequiredFieldError("topic")
	}

	return nil
}

// ExamplePublishSuccessRequest is an example of a successful publish request for OpenAPI documentation
var ExamplePublishSuccessRequest = PublishRequest{
	Tags:    map[string]string{"tag1": "meow", "tag2": "meowzer"},
	Topic:   "meow",
	Message: "hot diggity dog",
}

// ExamplePublishSuccessResponse is an example of a successful publish response for OpenAPI documentation
var ExamplePublishSuccessResponse = PublishReply{
	Reply: rout.Reply{
		Success: true,
	},
	Message: "success!",
}

// =========
// ORGANIZATION INVITE
// =========

// InviteRequest holds the fields that should be included on a request to the `/invite` endpoint
type InviteRequest struct {
	Token string `query:"token"`
}

// InviteReply holds the fields that are sent on a response to an accepted invitation
type InviteReply struct {
	rout.Reply
	ID          string `json:"user_id"`
	Email       string `json:"email"`
	Message     string `json:"message"`
	JoinedOrgID string `json:"joined_org_id"`
	Role        string `json:"role"`
	AuthData
}

// Validate ensures the required fields are set on the InviteRequest request
func (r *InviteRequest) Validate() error {
	if r.Token == "" {
		return rout.NewMissingRequiredFieldError("token")
	}

	return nil
}

// ExampleInviteRequest is an example of a successful invite request for OpenAPI documentation
var ExampleInviteRequest = InviteRequest{
	Token: "token",
}

// ExampleInviteResponse is an example of a successful invite response for OpenAPI documentation
var ExampleInviteResponse = InviteReply{
	Reply:       rout.Reply{Success: true},
	ID:          "1234",
	Email:       "",
	JoinedOrgID: "1234",
	Role:        "admin",
	Message:     "Welcome to your new organization!",
	AuthData: AuthData{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		Session:      "session",
		TokenType:    "bearer",
	},
}

// =========
// OAUTH
// =========

// OauthTokenRequest to authenticate an oauth user with the Server
type OauthTokenRequest struct {
	Name             string `json:"name"`
	Email            string `json:"email"`
	AuthProvider     string `json:"authProvider"`
	ExternalUserID   string `json:"externalUserId"`
	ExternalUserName string `json:"externalUserName"`
	ClientToken      string `json:"clientToken"`
	Image            string `json:"image,omitempty"`
}

// =========
// ACCOUNT/ACCESS
// =========

// AccountAccessRequest holds the fields that should be included on a request to the `/account/access` endpoint
type AccountAccessRequest struct {
	ObjectID    string `json:"objectId"`
	ObjectType  string `json:"objectType"`
	Relation    string `json:"relation"`
	SubjectType string `json:"subjectType,omitempty"`
}

// AccountAccessReply holds the fields that are sent on a response to the `/account/access` endpoint
type AccountAccessReply struct {
	rout.Reply
	Allowed bool `json:"allowed"`
}

// Validate ensures the required fields are set on the AccountAccessRequest
func (r *AccountAccessRequest) Validate() error {
	if r.ObjectID == "" {
		return rout.NewMissingRequiredFieldError("objectId")
	}

	if r.ObjectType == "" {
		return rout.NewMissingRequiredFieldError("objectType")
	}

	if r.Relation == "" {
		return rout.NewMissingRequiredFieldError("relation")
	}

	// Default to user if not set, only when using an API token should this be overwritten and set to service
	if r.SubjectType == "" {
		r.SubjectType = "user"
	}

	return nil
}

// ExampleAccountAccessRequest is an example of a successful `/account/access` request for OpenAPI documentation
var ExampleAccountAccessRequest = AccountAccessRequest{
	Relation:   "can_view",
	ObjectType: "organization",
	ObjectID:   "01J4EXD5MM60CX4YNYN0DEE3Y1",
}

// ExampleAccountAccessReply is an example of a successful `/account/access` response for OpenAPI documentation
var ExampleAccountAccessReply = AccountAccessReply{
	Reply:   rout.Reply{Success: true},
	Allowed: true,
}

// =========
// ACCOUNT/ROLES
// =========

// AccountRolesRequest holds the fields that should be included on a request to the `/account/roles` endpoint
type AccountRolesRequest struct {
	ObjectID    string   `json:"objectId"`
	ObjectType  string   `json:"objectType"`
	SubjectType string   `json:"subjectType,omitempty"`
	Relations   []string `json:"relations,omitempty"`
}

// AccountRolesReply holds the fields that are sent on a response to the `/account/roles` endpoint
type AccountRolesReply struct {
	rout.Reply
	Roles []string `json:"roles"`
}

// Validate ensures the required fields are set on the AccountAccessRequest
func (r *AccountRolesRequest) Validate() error {
	if r.ObjectID == "" {
		return rout.NewMissingRequiredFieldError("objectId")
	}

	if r.ObjectType == "" {
		return rout.NewMissingRequiredFieldError("objectType")
	}

	// Default to user if not set, only when using an API token should this be overwritten and set to service
	if r.SubjectType == "" {
		r.SubjectType = "user"
	}

	return nil
}

// ExampleAccountRolesRequest is an example of a successful `/account/roles` request for OpenAPI documentation
var ExampleAccountRolesRequest = AccountRolesRequest{
	ObjectType: "organization",
	ObjectID:   "01J4EXD5MM60CX4YNYN0DEE3Y1",
}

// ExampleAccountRolesReply is an example of a successful `/account/roles` response for OpenAPI documentation
var ExampleAccountRolesReply = AccountRolesReply{
	Reply: rout.Reply{Success: true},
	Roles: []string{"can_view", "can_edit", "audit_log_viewer"},
}

// =========
// ACCOUNT/ROLES/ORGANIZATION
// =========

// AccountRolesOrganizationRequest holds the fields that should be included on a request to the `/account/roles/organization` endpoint
type AccountRolesOrganizationRequest struct {
	ID string `param:"id"`
}

// AccountRolesOrganizationReply holds the fields that are sent on a response to the `/account/roles/organization` endpoint
type AccountRolesOrganizationReply struct {
	rout.Reply
	Roles          []string `json:"roles"`
	OrganizationID string   `json:"organization_id"`
}

// Validate ensures the required fields are set on the AccountAccessRequest
func (r *AccountRolesOrganizationRequest) Validate() error {
	if r.ID == "" {
		return rout.NewMissingRequiredFieldError("organization id")
	}

	return nil
}

// ExampleAccountRolesOrganizationRequest is an example of a successful `/account/roles/organization` request for OpenAPI documentation
var ExampleAccountRolesOrganizationRequest = AccountRolesOrganizationRequest{
	ID: "01J4HMNDSZCCQBTY93BF9CBF5D",
}

// ExampleAccountRolesOrganizationReply is an example of a successful `/account/roles/organization` response for OpenAPI documentation
var ExampleAccountRolesOrganizationReply = AccountRolesOrganizationReply{
	Reply:          rout.Reply{Success: true},
	Roles:          []string{"can_view", "can_edit", "audit_log_viewer"},
	OrganizationID: "01J4HMNDSZCCQBTY93BF9CBF5D",
}

// =========
// FILES
// =========

// UploadFilesRequest holds the fields that should be included on a request to the `/upload` endpoint
type UploadFilesRequest struct {
	UploadFile multipart.FileHeader `form:"uploadFile"`
}

// UploadFilesReply holds the fields that are sent on a response to the `/upload` endpoint
type UploadFilesReply struct {
	rout.Reply
	Message   string `json:"message,omitempty"`
	FileCount int64  `json:"file_count,omitempty"`
	Files     []File `json:"files,omitempty"`
}

// File holds the fields that are sent on a response to the `/upload` endpoint
type File struct {
	ID           string    `json:"id,omitempty"`
	Name         string    `json:"name,omitempty"`
	Size         int64     `json:"size,omitempty"`
	MimeType     string    `json:"mime_type,omitempty"`
	ContentType  string    `json:"content_type,omitempty"`
	PresignedURL string    `json:"presigned_url,omitempty"`
	MD5          []byte    `json:"md5,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

// ExampleUploadFileRequest is an example of a successful upload request for OpenAPI documentation
var ExampleUploadFileRequest = UploadFilesRequest{
	UploadFile: multipart.FileHeader{
		Filename: "file.txt",
		Size:     1024, // nolint:mnd
		Header: textproto.MIMEHeader{
			"Content-Type": []string{"text/plain"},
		},
	},
}

// ExampleUploadFilesSuccessResponse is an example of a successful upload response for OpenAPI documentation
var ExampleUploadFilesSuccessResponse = UploadFilesReply{
	Reply: rout.Reply{
		Success: true,
	},
	Message:   "file(s) uploaded successfully",
	FileCount: 1,
	Files: []File{
		{
			ID:           "1234",
			Name:         "file.txt",
			Size:         1024, // nolint:mnd
			MimeType:     "text/plain",
			ContentType:  "text/plain",
			PresignedURL: "https://example.com/file.txt",
			MD5:          []byte("1234"),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	},
}

// =========
// ENTITLEMENTS
// =========

// EntitlementsRequest holds the fields that should be included on a request to the `/entitlements` endpoint
type EntitlementsRequest struct {
	OrganizationID string `json:"organization_id"`
}

// EntitlementsReply holds the fields that are sent on a response to the `/entitlements` endpoint
type EntitlementsReply struct {
	rout.Reply
	SessionID string `json:"entitlements"`
}
