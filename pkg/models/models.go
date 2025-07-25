package models

import (
	"errors"
	"mime/multipart"
	"net/mail"
	"net/textproto"
	"slices"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"

	"github.com/theopenlane/utils/rout"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/pkg/enums"

	"github.com/theopenlane/utils/passwd"
)

var (
	errProviderRequired      = errors.New("provider parameter is required")
	errIntegrationIDRequired = errors.New("integration ID is required")
)

// =========
// Auth Data
// =========

type AuthData struct {
	AccessToken  string `json:"access_token" description:"The access token to be used for authentication"`
	RefreshToken string `json:"refresh_token,omitempty" description:"The refresh token to be used to refresh the access token after it expires"`
	Session      string `json:"session,omitempty" description:"The short-lived session token required for authentication"`
	TokenType    string `json:"token_type,omitempty" description:"The type of token being returned" example:"bearer"`
}

// =========
// LOGIN
// =========

// LoginRequest holds the login payload for the /login route
type LoginRequest struct {
	Username string `json:"username" description:"The email address associated with the existing account" example:"jsnow@example.com"`
	Password string `json:"password" description:"The password associated with the account" example:"Wint3rIsC0ming123!"`
}

// LoginReply holds the response to LoginRequest
type LoginReply struct {
	rout.Reply
	AuthData
	TFAEnabled bool   `json:"tfa_enabled,omitempty"`
	Message    string `json:"message"`
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

// AvailableAuthTypeReply holds the response to AvailableAuthTypeLoginRequest
type AvailableAuthTypeReply struct {
	rout.Reply
	Methods []enums.AuthProvider `json:"methods,omitempty"`
}

// AvailableAuthTypeLoginRequest holds the payload for checking the auth types available to a user
// passkeys? or both passkeys and credentials or just credentials
type AvailableAuthTypeLoginRequest struct {
	Username string `json:"username" description:"The email address associated with the existing account" example:"jsnow@example.com"`
}

// Validate ensures the required fields are set on the AvailableAuthTypeLoginRequest request
func (r *AvailableAuthTypeLoginRequest) Validate() error {
	r.Username = strings.TrimSpace(r.Username)

	if r.Username == "" {
		return rout.NewMissingRequiredFieldError("username")
	}

	if _, err := mail.ParseAddress(r.Username); err != nil {
		return rout.InvalidField("username")
	}

	return nil
}

// ExampleLoginSuccessRequest is an example of a successful login request for OpenAPI documentation
var ExampleLoginSuccessRequest = LoginRequest{
	Username: "sfunky@theopenlane.io",
	Password: "mitb!",
}

// ExampleAvailableAuthTypeRequest is an example of a successful available auth type check for OpenAPI documentation
var ExampleAvailableAuthTypeRequest = LoginRequest{
	Username: "sfunky@theopenlane.io",
}

// ExampleAvailableAuthTypeSuccessResponse is an example of a successful available auth methods check response for OpenAPI documentation
var ExampleAvailableAuthTypeSuccessResponse = AvailableAuthTypeReply{
	Reply: rout.Reply{
		Success: true,
	},
	Methods: []enums.AuthProvider{
		enums.AuthProviderCredentials,
		enums.AuthProviderWebauthn,
	},
}

// ExampleLoginSuccessResponse is an example of a successful login response for OpenAPI documentation
var ExampleLoginSuccessResponse = LoginReply{
	Reply: rout.Reply{
		Success: true,
	},
	TFAEnabled: true,
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
	RefreshToken string `json:"refresh_token" description:"The token to be used to refresh the access token after expiration"`
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
	FirstName string  `json:"first_name,omitempty" description:"The first name of the user" example:"Jon"`
	LastName  string  `json:"last_name,omitempty" description:"The last name of the user" example:"Snow"`
	Email     string  `json:"email,omitempty" description:"The email address of the user" example:"jsnow@example.com"`
	Password  string  `json:"password,omitempty" description:"The password to be used for authentication after registration" example:"Wint3rIsC0ming123!"`
	Token     *string `json:"token" description:"A newly invited user can use this to join a org as at the same time they are creating their account"`
}

// RegisterReply holds the fields that are sent on a response to the `/register` endpoint
type RegisterReply struct {
	rout.Reply
	ID      string `json:"user_id" description:"The ID of the user that was created" example:"01J4EXD5MM60CX4YNYN0DEE3Y1"`
	Email   string `json:"email" description:"The email address of the user" example:"jsnow@example.com"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty" exclude:"true"` // only used for requests against local development, excluded from OpenAPI documentation
}

// Validate ensures the required fields are set on the RegisterRequest request
func (r *RegisterRequest) Validate() error {
	r.FirstName = strings.TrimSpace(r.FirstName)
	r.LastName = strings.TrimSpace(r.LastName)
	r.Email = strings.TrimSpace(r.Email)
	r.Password = strings.TrimSpace(r.Password)

	if r.Token != nil {
		invitationToken := strings.TrimSpace(*r.Token)
		r.Token = &invitationToken
	}

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
}

// =========
// SWITCH ORGANIZATION
// =========

// SwitchOrganizationRequest contains the target organization ID being switched to for the /switch endpoint
type SwitchOrganizationRequest struct {
	TargetOrganizationID string `json:"target_organization_id" description:"The ID of the organization to switch to" example:"01J4EXD5MM60CX4YNYN0DEE3Y1"`
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
	Token string `query:"token" description:"The token to be used to verify the email address, token is sent via email"`
}

// VerifyReply holds the fields that are sent on a response to the `/verify` endpoint
type VerifyReply struct {
	rout.Reply
	ID      string `json:"user_id" description:"The ID of the user that was created" example:"01J4EXD5MM60CX4YNYN0DEE3Y1"`
	Email   string `json:"email" description:"The email address of the user" example:"jsnow@example.com"`
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
	Email string `json:"email" description:"The email address to resend the verification email to, must match the email address on the existing account"`
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
	Email string `json:"email" description:"The email address associated with the account to send the password reset email to" example:"jsnow@example.com"`
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
	Password string `json:"password" description:"The new password to be used for authentication"`
	Token    string `json:"token" description:"The token to be used to reset the password, token is sent via email"`
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
	Email string `json:"email" description:"The email address associated with the account" example:"jsnow@example.com"`
	Name  string `json:"name" description:"The name of the user" example:"Jon Snow"`
}

func (r *WebauthnRegistrationRequest) Validate() error {
	if r.Email == "" {
		return rout.NewMissingRequiredFieldError("email")
	}

	if r.Name == "" {
		return rout.NewMissingRequiredFieldError("name")
	}

	return nil
}

// WebauthnBeginRegistrationResponse is the response to begin a webauthn login
// this includes the credential creation options and the session token
type WebauthnBeginRegistrationResponse struct {
	Reply rout.Reply
	*protocol.CredentialCreation
	Session string `json:"session,omitempty"`
}

var ExampleWebauthnBeginRegistrationRequest = WebauthnRegistrationRequest{
	Email: "sarahisthebest@sarahsthebest.com",
	Name:  "Sarah Funk",
}

var ExampleWebauthnBeginRegistrationResponse = WebauthnBeginRegistrationResponse{
	Reply: rout.Reply{Success: true},
	CredentialCreation: &protocol.CredentialCreation{
		Response: protocol.PublicKeyCredentialCreationOptions{
			RelyingParty: protocol.RelyingPartyEntity{},
			User:         protocol.UserEntity{},
			Challenge:    protocol.URLEncodedBase64{},
		}},
	Session: "session",
}

// WebauthnRegistrationResponse is the response after a successful webauthn registration
type WebauthnRegistrationResponse struct {
	rout.Reply
	Message string `json:"message,omitempty"`
	AuthData
}

// WebauthnLoginRequest is the request to begin a webauthn login
type WebauthnLoginRequest struct {
	Email string `json:"email" description:"The email address associated with the account" example:"jsnow@example.com"`
}

func (r *WebauthnLoginRequest) Validate() error {
	if r.Email == "" {
		return rout.NewMissingRequiredFieldError("email")
	}

	return nil
}

// WebauthnBeginLoginResponse is the response to begin a webauthn login
// this includes the credential assertion options and the session token
type WebauthnBeginLoginResponse struct {
	Reply rout.Reply
	*protocol.CredentialAssertion
	Session string `json:"session,omitempty"`
}

// WebauthnLoginResponse is the response after a successful webauthn login
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
	Token string `query:"token" description:"The token to be used to verify the subscription, token is sent via email"`
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
// ORGANIZATION INVITE
// =========

// InviteRequest holds the fields that should be included on a request to the `/invite` endpoint
type InviteRequest struct {
	Token string `query:"token" description:"The token to be used to accept the invitation, token is sent via email"`
}

// InviteReply holds the fields that are sent on a response to an accepted invitation
type InviteReply struct {
	rout.Reply
	ID          string `json:"user_id" description:"The ID of the user that was created" example:"01J4EXD5MM60CX4YNYN0DEE3Y1"`
	Email       string `json:"email" description:"The email address of the user" example:"jsnow@example.com"`
	Message     string `json:"message"`
	JoinedOrgID string `json:"joined_org_id" description:"The ID of the organization the user joined" example:"01JJFVMGENQS9ZG3GVA50QVX5E"`
	Role        string `json:"role" description:"The role the user has in the organization" example:"admin"`
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
	Name             string `json:"name" description:"The name of the user" example:"Jon Snow"`
	Email            string `json:"email" description:"The email address of the user" example:"jsnow@example.com"`
	AuthProvider     string `json:"authProvider" description:"The provider used to authenticate the user, e.g. google, github, etc." example:"google"`
	ExternalUserID   string `json:"externalUserId" description:"The ID of the user from the external provider" example:"1234567890"`
	ExternalUserName string `json:"externalUserName" description:"The username of the user from the external provider" example:"jsnow"`
	ClientToken      string `json:"clientToken" description:"The token provided by the external provider"`
	Image            string `json:"image,omitempty" description:"The image URL of the user from the external provider"`
}

// =========
// ACCOUNT/ACCESS
// =========

// AccountAccessRequest holds the fields that should be included on a request to the `/account/access` endpoint
type AccountAccessRequest struct {
	ObjectID    string `json:"object_id" description:"The ID of the object to check access for" example:"01J4EXD5MM60CX4YNYN0DEE3Y1"`
	ObjectType  string `json:"object_type" description:"The type of object to check access for, e.g. organization, program, procedure, etc" example:"organization"`
	Relation    string `json:"relation" description:"The relation to check access for, e.g. can_view, can_edit" example:"can_view"`
	SubjectType string `json:"subject_type,omitempty" description:"The type of subject to check access for, e.g. service, user" example:"user"`
}

// AccountAccessReply holds the fields that are sent on a response to the `/account/access` endpoint
type AccountAccessReply struct {
	rout.Reply
	Allowed bool `json:"allowed"`
}

// Validate ensures the required fields are set on the AccountAccessRequest
func (r *AccountAccessRequest) Validate() error {
	if r.ObjectID == "" {
		return rout.NewMissingRequiredFieldError("object_id")
	}

	if r.ObjectType == "" {
		return rout.NewMissingRequiredFieldError("object_type")
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
	ObjectID    string   `json:"object_id" description:"The ID of the object to check roles for" example:"01J4EXD5MM60CX4YNYN0DEE3Y1"`
	ObjectType  string   `json:"object_type" description:"The type of object to check roles for, e.g. organization, program, procedure, etc" example:"organization"`
	SubjectType string   `json:"subject_type,omitempty" description:"The type of subject to check roles for, e.g. service, user" example:"user"`
	Relations   []string `json:"relations,omitempty" description:"The relations to check roles for, e.g. can_view, can_edit" example:"can_view"`
}

// AccountRolesReply holds the fields that are sent on a response to the `/account/roles` endpoint
type AccountRolesReply struct {
	rout.Reply
	Roles []string `json:"roles"`
}

// Validate ensures the required fields are set on the AccountAccessRequest
func (r *AccountRolesRequest) Validate() error {
	if r.ObjectID == "" {
		return rout.NewMissingRequiredFieldError("object_id")
	}

	if r.ObjectType == "" {
		return rout.NewMissingRequiredFieldError("object_type")
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
	ID string `param:"id" description:"The ID of the organization to check roles for" example:"01J4HMNDSZCCQBTY93BF9CBF5D"`
}

// AccountRolesOrganizationReply holds the fields that are sent on a response to the `/account/roles/organization` endpoint
type AccountRolesOrganizationReply struct {
	rout.Reply
	Roles          []string `json:"roles" description:"The roles the user has in the organization, e.g. can_view, can_edit" example:"can_view, can_edit"`
	OrganizationID string   `json:"organization_id" description:"The ID of the organization the user has roles in" example:"01J4HMNDSZCCQBTY93BF9CBF5D"`
}

// Validate ensures the required fields are set on the AccountRolesOrganizationRequest
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
// ACCOUNT/FEATURES
// =========

// AccountFeaturesRequest holds the fields that should be included on a request to the `/account/features` endpoint
type AccountFeaturesRequest struct {
	ID string `param:"id" description:"The ID of the organization to check roles for" example:"01J4HMNDSZCCQBTY93BF9CBF5D"`
}

// AccountFeaturesReply holds the fields that are sent on a response to the `/account/features` endpoint
type AccountFeaturesReply struct {
	rout.Reply
	Features       []string `json:"features" description:"The features the user has access to in the organization, e.g. policy-and-procedure-module, compliance-module" example:"policy-and-procedure-module, centralized-audit-documentation, risk-management, compliance-module"`
	OrganizationID string   `json:"organization_id" description:"The ID of the organization the user has features in" example:"01J4HMNDSZCCQBTY93BF9CBF5D"`
}

// Validate ensures the required fields are set on the AccountFeaturesRequest
func (r *AccountFeaturesRequest) Validate() error {
	if r.ID == "" {
		return rout.NewMissingRequiredFieldError("organization id")
	}

	return nil
}

// ExampleAccountFeaturesRequest is an example of a successful `/account/features` request for OpenAPI documentation
var ExampleAccountFeaturesRequest = AccountFeaturesRequest{
	ID: "01J4HMNDSZCCQBTY93BF9CBF5D",
}

// ExampleAccountFeaturesReply is an example of a successful `/account/features` response for OpenAPI documentation
var ExampleAccountFeaturesReply = AccountFeaturesReply{
	Reply:          rout.Reply{Success: true},
	Features:       []string{},
	OrganizationID: "01J4HMNDSZCCQBTY93BF9CBF5D",
}

// =========
// FILES
// =========

// UploadFilesRequest holds the fields that should be included on a request to the `/upload` endpoint
type UploadFilesRequest struct {
	UploadFile multipart.FileHeader `form:"uploadFile" description:"The file to be uploaded"`
}

// UploadFilesReply holds the fields that are sent on a response to the `/upload` endpoint
type UploadFilesReply struct {
	rout.Reply
	Message   string `json:"message,omitempty"`
	FileCount int64  `json:"file_count,omitempty" description:"The number of files uploaded"`
	Files     []File `json:"files,omitempty" description:"The files that were uploaded"`
}

// File holds the fields that are sent on a response to the `/upload` endpoint
type File struct {
	ID           string    `json:"id,omitempty" description:"The ID of the uploaded file"`
	Name         string    `json:"name,omitempty" description:"The name of the uploaded file"`
	Size         int64     `json:"size,omitempty" description:"The size of the uploaded file in bytes"`
	MimeType     string    `json:"mime_type,omitempty" description:"The mime type of the uploaded file"`
	ContentType  string    `json:"content_type,omitempty" description:"The content type of the uploaded file"`
	PresignedURL string    `json:"presigned_url,omitempty" description:"The presigned URL to download the file"`
	MD5          []byte    `json:"md5,omitempty" description:"The MD5 hash of the uploaded file"`
	CreatedAt    time.Time `json:"created_at,omitempty" description:"The time the file was uploaded"`
	UpdatedAt    time.Time `json:"updated_at,omitempty" description:"The time the file was last updated"`
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
// TFA VALIDATION
// =========

// TFARequest holds the payload for verifying the 2fa code (/2fa/validate)
type TFARequest struct {
	TOTPCode     string `json:"totp_code,omitempty" description:"The TOTP code to validate, always takes precedence over recovery code" example:"113371"`
	RecoveryCode string `json:"recovery_code,omitempty" description:"The recovery code to validate, only used if TOTP code is not provided" example:"8VM7AL91"`
}

// TFAReply holds the response to TFARequest
type TFAReply struct {
	rout.Reply
	Message string `json:"message"`
}

// Validate ensures the required fields are set on the TFARequest request
func (r *TFARequest) Validate() error {
	if r.TOTPCode == "" && r.RecoveryCode == "" {
		return rout.NewMissingRequiredFieldError("totp_code")
	}

	return nil
}

// ExampleLoginSuccessRequest is an example of a successful tfa validation request for OpenAPI documentation
var ExampleTFASuccessRequest = TFARequest{
	TOTPCode: "113371",
}

// ExampleLoginSuccessResponse is an example of a successful tfa validation response for OpenAPI documentation
var ExampleTFASSuccessResponse = TFAReply{
	Reply: rout.Reply{
		Success: true,
	},
}

// =========
// EXAMPLECSV REQUEST
// =========

// ExampleCSVRequest holds the payload for serving example CSV files
type ExampleCSVRequest struct {
	Filename string `json:"filename" description:"the file name to check for" example:"actionplan"`
}

// Validate ensures the required fields are set on the ExampleCSVRequest request
func (r *ExampleCSVRequest) Validate() error {
	if r.Filename == "" {
		return rout.NewMissingRequiredFieldError("filename")
	}

	return nil
}

// ExampleLoginSuccessRequest is an example of a successful tfa validation request for OpenAPI documentation
var ExampleExampleCSVRequest = ExampleCSVRequest{
	Filename: "actionplan",
}

// =========
// JOB RUNNERS
// =========

// JobRunnerRegistrationRequest is the request to register a new node
type JobRunnerRegistrationRequest struct {
	IPAddress string   `json:"ip_address" description:"The IP address of the node being registered"`
	Token     string   `json:"token" description:"Your agent registration token"`
	Name      string   `json:"name" description:"the name of your job runner node"`
	Tags      []string `json:"tags" description:"The tags for your runner node"`
}

// Validate ensures the required fields are set on the AgentNodeRegistrationRequest
func (r *JobRunnerRegistrationRequest) Validate() error {
	if r.IPAddress == "" {
		return rout.NewMissingRequiredFieldError("ip_address")
	}

	if r.Token == "" {
		return rout.NewMissingRequiredFieldError("token")
	}

	if r.Name == "" {
		return rout.NewMissingRequiredFieldError("name")
	}

	if len(r.Tags) == 0 {
		r.Tags = append(r.Tags, "self-hosted")
	}

	return nil
}

// JobRunnerRegistrationReply is the response to begin a webauthn login
// this includes the credential creation options and the session token
type JobRunnerRegistrationReply struct {
	Reply   rout.Reply
	Message string `json:"message"`
}

// ExampleJobRunnerRegistrationRequest is an example of a successful job runner
// registration request
var ExampleJobRunnerRegistrationRequest = JobRunnerRegistrationRequest{
	IPAddress: "192.168.0.1",
	Name:      "ubuntu-eu-west-2",
	Token:     "registration_tokenhere",
	Tags:      []string{"self-hosted", "eu-west-2", "gcp", "kubernetes"},
}

// ExampleJobRunnerRegistrationResponse is an example of a successful job runner
// registration response
var ExampleJobRunnerRegistrationResponse = JobRunnerRegistrationReply{
	Reply:   rout.Reply{Success: true},
	Message: "Job runner node registered",
}

// AcmeSolverRequest is the request to solve an acme challenge
type AcmeSolverRequest struct {
	Path string `param:"path" description:"The path to the acme challenge" example:"01J4HMNDSZCCQBTY93BF9CBF5D"`
}

// =========
// SSO
// =========

// WebfingerRequest represents the query parameters accepted by the
// `/.well-known/webfinger` endpoint
//
// The `resource` field must be provided and should be prefixed with
// `org:` for organization lookups or `acct:` for user lookups
type WebfingerRequest struct {
	Resource string `query:"resource" description:"resource identifier prefixed with org: or acct:" example:"acct:meowmeow@kitties.com"`
}

// Validate ensures a valid resource was provided on the WebfingerRequest
func (r *WebfingerRequest) Validate() error {
	r.Resource = strings.TrimSpace(r.Resource)
	switch {
	case r.Resource == "":
		return rout.NewMissingRequiredFieldError("resource")
	case !strings.HasPrefix(r.Resource, "org:") && !strings.HasPrefix(r.Resource, "acct:"):
		return rout.InvalidField("resource")
	}

	return nil
}

// ExampleWebfingerRequest is an example request for OpenAPI documentation
var ExampleWebfingerRequest = WebfingerRequest{
	Resource: "acct:sarah@funkyhous.info",
}

// SSOLoginRequest holds the query parameters for initiating an SSO login flow
type SSOLoginRequest struct {
	OrganizationID string `json:"organization_id" query:"organization_id" description:"organization id" example:"01J4EXD5MM60CX4YNYN0DEE3Y1"`
	ReturnURL      string `json:"return" query:"return" description:"return url after authentication" example:"https://app.mitb.com"`
}

// Validate ensures the required fields are set on the SSOLoginRequest
func (r *SSOLoginRequest) Validate() error {
	r.OrganizationID = strings.TrimSpace(r.OrganizationID)
	r.ReturnURL = strings.TrimSpace(r.ReturnURL)

	if r.OrganizationID == "" {
		return rout.NewMissingRequiredFieldError("organization_id")
	}

	return nil
}

// ExampleSSOLoginRequest is an example request for OpenAPI documentation
var ExampleSSOLoginRequest = SSOLoginRequest{
	OrganizationID: ulids.New().String(),
	ReturnURL:      "https://app.sitb.com",
}

// SSOCallbackRequest holds the query parameters for completing the SSO login flow
type SSOCallbackRequest struct {
	Code           string `json:"code" query:"code" description:"authorization code" example:"abc"`
	State          string `json:"state" query:"state" description:"state value" example:"state123"`
	OrganizationID string `json:"organization_id" query:"organization_id" description:"organization id" example:"01J4EXD5MM60CX4YNYN0DEE3Y1"`
}

// Validate ensures the required fields are set on the SSOCallbackRequest
func (r *SSOCallbackRequest) Validate() error {
	r.Code = strings.TrimSpace(r.Code)
	r.State = strings.TrimSpace(r.State)
	r.OrganizationID = strings.TrimSpace(r.OrganizationID)

	switch {
	case r.Code == "":
		return rout.NewMissingRequiredFieldError("code")
	case r.State == "":
		return rout.NewMissingRequiredFieldError("state")
	case r.OrganizationID == "":
		return rout.NewMissingRequiredFieldError("organization_id")
	}

	return nil
}

// ExampleSSOCallbackRequest is an example request for OpenAPI documentation
var ExampleSSOCallbackRequest = SSOCallbackRequest{
	Code:           "code",
	State:          "state",
	OrganizationID: ulids.New().String(),
}

// SSOTokenCallbackRequest holds the query parameters for completing token SSO authorization
type SSOTokenCallbackRequest struct {
	Code  string `json:"code" query:"code" description:"authorization code" example:"abc"`
	State string `json:"state" query:"state" description:"state value" example:"state123"`
}

// Validate ensures required fields are set on the SSOTokenCallbackRequest
func (r *SSOTokenCallbackRequest) Validate() error {
	r.Code = strings.TrimSpace(r.Code)
	r.State = strings.TrimSpace(r.State)

	switch {
	case r.Code == "":
		return rout.NewMissingRequiredFieldError("code")
	case r.State == "":
		return rout.NewMissingRequiredFieldError("state")
	}

	return nil
}

// ExampleSSOTokenCallbackRequest is an example request for OpenAPI documentation
var ExampleSSOTokenCallbackRequest = SSOTokenCallbackRequest{
	Code:  "code",
	State: "state",
}

// SSOStatusRequest is the request to check if SSO login is enforced for an organization
type SSOStatusRequest struct {
	Resource string `query:"resource" description:"organization or user email to check" example:"org:01J4EXD5MM60CX4YNYN0DEE3Y1"`
}

// Validate ensures the required fields are set on the SSOStatusRequest request
func (r *SSOStatusRequest) Validate() error {
	if r.Resource == "" {
		return rout.NewMissingRequiredFieldError("resource")
	}

	return nil
}

// SSOStatusReply is the response for SSOStatusRequest
type SSOStatusReply struct {
	rout.Reply
	Enforced       bool              `json:"enforced"`
	Provider       enums.SSOProvider `json:"provider,omitempty"`
	DiscoveryURL   string            `json:"discovery_url,omitempty"`
	OrganizationID string            `json:"organization_id,omitempty"`
}

// ExampleSSOStatusRequest is an example request for OpenAPI documentation
var ExampleSSOStatusRequest = SSOStatusRequest{
	Resource: "acct:mitb@theopenlane.io",
}

// ExampleSSOStatusReply is an example response for OpenAPI documentation
var ExampleSSOStatusReply = SSOStatusReply{
	Reply:          rout.Reply{Success: true},
	Enforced:       true,
	Provider:       enums.SSOProviderOkta,
	DiscoveryURL:   "https://id.example.com/.well-known/openid-configuration",
	OrganizationID: ulids.New().String(),
}

// SSOTokenAuthorizeRequest is the request for authorizing a token for SSO use
// with an organization
type SSOTokenAuthorizeRequest struct {
	OrganizationID string `json:"organization_id" query:"organization_id" description:"organization id" example:"01J4EXD5MM60CX4YNYN0DEE3Y1"`
	TokenID        string `json:"token_id" query:"token_id" description:"token id to authorize" example:"01JJFVMGENQS9ZG3GVA50QVX5E"`
	TokenType      string `json:"token_type" query:"token_type" description:"token type: api or personal" example:"api"`
}

// Validate ensures required fields are set on the SSOTokenAuthorizeRequest
func (r *SSOTokenAuthorizeRequest) Validate() error {
	r.OrganizationID = strings.TrimSpace(r.OrganizationID)
	r.TokenID = strings.TrimSpace(r.TokenID)
	r.TokenType = strings.TrimSpace(r.TokenType)

	switch {
	case r.OrganizationID == "":
		return rout.NewMissingRequiredFieldError("organization_id")
	case r.TokenID == "":
		return rout.NewMissingRequiredFieldError("token_id")
	case r.TokenType == "":
		return rout.NewMissingRequiredFieldError("token_type")
	}

	return nil
}

// SSOTokenAuthorizeReply is returned when a token has been successfully
// authorized for SSO
type SSOTokenAuthorizeReply struct {
	rout.Reply
	OrganizationID string `json:"organization_id"`
	TokenID        string `json:"token_id"`
	Message        string `json:"message,omitempty"`
}

// ExampleSSOTokenAuthorizeRequest is an example request for OpenAPI documentation
var ExampleSSOTokenAuthorizeRequest = SSOTokenAuthorizeRequest{
	OrganizationID: ulids.New().String(),
	TokenID:        ulids.New().String(),
	TokenType:      "api",
}

// ExampleSSOTokenAuthorizeReply is an example response for OpenAPI documentation
var ExampleSSOTokenAuthorizeReply = SSOTokenAuthorizeReply{
	Reply:          rout.Reply{Success: true},
	OrganizationID: ulids.New().String(),
	TokenID:        ulids.New().String(),
	Message:        "success",
}

// CreateTrustCenterAnonymousJWTResponse is the response to a request to create a trust center anonymous JWT
type CreateTrustCenterAnonymousJWTResponse struct {
	AuthData
}

// =========
// OAUTH INTEGRATIONS
// =========

// IntegrationToken represents stored OAuth tokens for an integration
type IntegrationToken struct {
	Provider         string     `json:"provider" description:"OAuth provider (github, slack, etc.)"`
	AccessToken      string     `json:"accessToken" description:"OAuth access token"`
	RefreshToken     string     `json:"refreshToken,omitempty" description:"OAuth refresh token"`
	ExpiresAt        *time.Time `json:"expiresAt,omitempty" description:"Token expiration time"`
	ProviderUserID   string     `json:"providerUserId,omitempty" description:"User ID from the OAuth provider"`
	ProviderUsername string     `json:"providerUsername,omitempty" description:"Username from the OAuth provider"`
	ProviderEmail    string     `json:"providerEmail,omitempty" description:"Email from the OAuth provider"`
}

// IsExpired returns true if the token has expired
func (t *IntegrationToken) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false // No expiry means never expires
	}
	return time.Now().After(*t.ExpiresAt)
}

// HasValidToken returns true if the token is valid and not expired
func (t *IntegrationToken) HasValidToken() bool {
	return t.Provider != "" && t.AccessToken != "" && !t.IsExpired()
}

// IntegrationTokenResponse is the response for getting integration tokens
type IntegrationTokenResponse struct {
	rout.Reply
	Provider  string            `json:"provider"`
	Token     *IntegrationToken `json:"token"`
	ExpiresAt *time.Time        `json:"expiresAt,omitempty"`
}

// ListIntegrationsResponse is the response for listing integrations
type ListIntegrationsResponse struct {
	rout.Reply
	Integrations any `json:"integrations"` // Will be []*ent.Integration
}

// DeleteIntegrationResponse is the response for deleting an integration
type DeleteIntegrationResponse struct {
	rout.Reply
	Message string `json:"message"`
}

// IntegrationStatusResponse is the response for checking integration status
type IntegrationStatusResponse struct {
	rout.Reply
	Provider     string `json:"provider"`
	Connected    bool   `json:"connected"`
	Status       string `json:"status,omitempty"` // "connected", "expired", "invalid"
	TokenValid   bool   `json:"tokenValid,omitempty"`
	TokenExpired bool   `json:"tokenExpired,omitempty"`
	Message      string `json:"message"`
	Integration  any    `json:"integration,omitempty"` // Will be *ent.Integration
}

// GetIntegrationTokenRequest is the request for getting integration tokens
type GetIntegrationTokenRequest struct {
	Provider string `param:"provider" description:"OAuth provider (github, slack, etc.)" example:"github"`
}

// DeleteIntegrationRequest is the request for deleting an integration
type DeleteIntegrationRequest struct {
	ID string `param:"id" description:"Integration ID" example:"01J4HMNDSZCCQBTY93BF9CBF5D"`
}

// RefreshIntegrationTokenRequest is the request for refreshing integration tokens
type RefreshIntegrationTokenRequest struct {
	Provider string `param:"provider" description:"OAuth provider (github, slack, etc.)" example:"github"`
}

// GetIntegrationStatusRequest is the request for checking integration status
type GetIntegrationStatusRequest struct {
	Provider string `param:"provider" description:"OAuth provider (github, slack, etc.)" example:"github"`
}

// Validate validates the GetIntegrationTokenRequest
func (r *GetIntegrationTokenRequest) Validate() error {
	r.Provider = strings.TrimSpace(r.Provider)
	if r.Provider == "" {
		return errProviderRequired
	}
	return nil
}

// Validate validates the DeleteIntegrationRequest
func (r *DeleteIntegrationRequest) Validate() error {
	r.ID = strings.TrimSpace(r.ID)
	if r.ID == "" {
		return errIntegrationIDRequired
	}
	return nil
}

// Validate validates the RefreshIntegrationTokenRequest
func (r *RefreshIntegrationTokenRequest) Validate() error {
	r.Provider = strings.TrimSpace(r.Provider)
	if r.Provider == "" {
		return errProviderRequired
	}
	return nil
}

// Validate validates the GetIntegrationStatusRequest
func (r *GetIntegrationStatusRequest) Validate() error {
	r.Provider = strings.TrimSpace(r.Provider)
	if r.Provider == "" {
		return errProviderRequired
	}
	return nil
}

// =========
// OAUTH INTEGRATION REQUESTS/RESPONSES
// =========

// OAuthFlowRequest represents the initial OAuth flow request
type OAuthFlowRequest struct {
	Provider    string   `json:"provider" description:"OAuth provider (github, slack, etc.)" example:"github"`
	RedirectURI string   `json:"redirectUri,omitempty" description:"Custom redirect URI after OAuth flow" example:"https://app.example.com/integrations"`
	Scopes      []string `json:"scopes,omitempty" description:"Additional OAuth scopes to request" example:"repo,gist"`
}

// Validate ensures the required fields are set on the OAuthFlowRequest
func (r *OAuthFlowRequest) Validate() error {
	r.Provider = strings.TrimSpace(strings.ToLower(r.Provider))

	if r.Provider == "" {
		return rout.NewMissingRequiredFieldError("provider")
	}

	// Validate supported providers
	supportedProviders := []string{"github", "slack"}
	validProvider := slices.Contains(supportedProviders, r.Provider)
	if !validProvider {
		return rout.InvalidField("provider")
	}

	// Clean up scopes
	cleanScopes := make([]string, 0, len(r.Scopes))
	for _, scope := range r.Scopes {
		if trimmed := strings.TrimSpace(scope); trimmed != "" {
			cleanScopes = append(cleanScopes, trimmed)
		}
	}
	r.Scopes = cleanScopes

	return nil
}

// OAuthFlowResponse contains the OAuth authorization URL
type OAuthFlowResponse struct {
	rout.Reply
	AuthURL       string `json:"authUrl" description:"URL to redirect user to for OAuth authorization" example:"https://github.com/login/oauth/authorize?client_id=..."`
	State         string `json:"state,omitempty" description:"OAuth state parameter for security" example:"eyJvcmdJRCI6IjAxSE..."`
	Message       string `json:"message,omitempty" description:"Optional message (e.g., for authentication required)" example:"Authentication required. Please login first."`
	RequiresLogin bool   `json:"requiresLogin,omitempty" description:"Whether user needs to login before OAuth flow" example:"true"`
}

// OAuthCallbackRequest represents the OAuth callback data
type OAuthCallbackRequest struct {
	Provider string `json:"provider,omitempty" query:"provider" description:"OAuth provider (extracted from state)" example:"github"`
	Code     string `json:"code" query:"code" description:"OAuth authorization code" example:"4/0AQl..."`
	State    string `json:"state" query:"state" description:"OAuth state parameter" example:"eyJvcmdJRCI6IjAxSE..."`
}

// Validate ensures the required fields are set on the OAuthCallbackRequest
func (r *OAuthCallbackRequest) Validate() error {
	r.Provider = strings.TrimSpace(strings.ToLower(r.Provider))
	r.Code = strings.TrimSpace(r.Code)
	r.State = strings.TrimSpace(r.State)

	switch {
	case r.Code == "":
		return rout.NewMissingRequiredFieldError("code")
	case r.State == "":
		return rout.NewMissingRequiredFieldError("state")
	}

	return nil
}

// OAuthCallbackResponse contains the result of OAuth callback processing
type OAuthCallbackResponse struct {
	rout.Reply
	Success     bool   `json:"success" description:"Whether the OAuth callback was processed successfully" example:"true"`
	Integration any    `json:"integration,omitempty" description:"The created/updated integration object"`
	Message     string `json:"message" description:"Success or error message" example:"Successfully connected GitHub integration"`
}

// ExampleOAuthFlowRequest is an example OAuth flow request for OpenAPI documentation
var ExampleOAuthFlowRequest = OAuthFlowRequest{
	Provider:    "github",
	RedirectURI: "https://app.example.com/integrations",
	Scopes:      []string{"repo", "gist"},
}

// ExampleOAuthCallbackRequest is an example OAuth callback request for OpenAPI documentation
var ExampleOAuthCallbackRequest = OAuthCallbackRequest{
	Provider: "github",
	Code:     "4/0AQlEz8xY...",
	State:    "eyJvcmdJRCI6IjAxSE...",
}

// ExampleOAuthFlowResponse is an example OAuth flow response for OpenAPI documentation
var ExampleOAuthFlowResponse = OAuthFlowResponse{
	Reply:         rout.Reply{Success: true},
	AuthURL:       "https://github.com/login/oauth/authorize?client_id=...&state=eyJvcmdJRCI6IjAxSE...",
	State:         "eyJvcmdJRCI6IjAxSE...",
	Message:       "",
	RequiresLogin: false,
}

// ExampleOAuthCallbackResponse is an example OAuth callback response for OpenAPI documentation
var ExampleOAuthCallbackResponse = OAuthCallbackResponse{
	Reply:   rout.Reply{Success: true},
	Success: true,
	Message: "Successfully connected GitHub integration",
}
