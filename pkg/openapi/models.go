package openapi

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/mail"
	"net/textproto"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/invopop/jsonschema"
	"github.com/samber/lo"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/pkg/catalog"
	"github.com/theopenlane/core/pkg/catalog/gencatalog"
	"github.com/theopenlane/core/pkg/enums"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"

	"github.com/theopenlane/utils/passwd"
)

const (
	exampleFindingsCount = 5
)

// ExampleProvider interface allows response models to provide their own examples
// This eliminates the need for separate Example* variables and static switch statements
type ExampleProvider interface {
	ExampleResponse() any
}

var (
	errProviderRequired      = errors.New("provider parameter is required")
	errIntegrationIDRequired = errors.New("integration ID is required")
)

const (
	exampleUserULID        = "01K9MJ23ND309PAN0ZV1ECFYT7"
	exampleUserAltULID     = "01K9MJ23ND309PAN0ZQFK6N2R3"
	exampleOrgULID         = "01K9MJ3PD7XKJSCT9ZWYGW9CVE"
	exampleOrgAltULID      = "01K9MJ23ND309PAN0ZTX6ESG47"
	exampleJoinedOrgULID   = "01K9MJ23ND309PAN0ZWTN4C0PJ"
	exampleSessionULID     = "01K9MJ23ND309PAN0Z6GN2BH90"
	exampleTokenULID       = "01K9MJ23ND309PAN0ZA9XQ3MFH" // #nosec G101 -- example token placeholder used only in docs
	exampleIntegrationULID = "01K9MJ23ND309PAN0ZNKQ6HB5S"
)

func exampleULID(key string) string {
	switch key {
	case "user":
		return exampleUserULID
	case "user_alt":
		return exampleUserAltULID
	case "organization":
		return exampleOrgULID
	case "organization_alt":
		return exampleOrgAltULID
	case "joined_org":
		return exampleJoinedOrgULID
	case "session":
		return exampleSessionULID
	case "token":
		return exampleTokenULID
	case "integration":
		return exampleIntegrationULID
	default:
		return exampleUserULID
	}
}

var exampleBaseTime = time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)

func exampleTime(offset time.Duration) time.Time {
	return exampleBaseTime.Add(offset)
}

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

// LoginRequest contains credentials for user authentication
type LoginRequest struct {
	Username string `json:"username" description:"The email address associated with the existing account" example:"jsnow@example.com"`
	Password string `json:"password" description:"The password associated with the account" example:"Wint3rIsC0ming123!"`
}

// LoginReply contains authentication tokens and user information after successful login
type LoginReply struct {
	rout.Reply
	AuthData
	TFAEnabled       bool   `json:"tfa_enabled,omitempty"`
	TFASetupRequired bool   `json:"tfa_required,omitempty"`
	Message          string `json:"message"`
}

// ExampleResponse returns an example LoginReply for OpenAPI documentation
func (r *LoginReply) ExampleResponse() any {
	return LoginReply{
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
		Message: "Login successful",
	}
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

// ExampleResponse returns an example AvailableAuthTypeReply for OpenAPI documentation
func (r *AvailableAuthTypeReply) ExampleResponse() any {
	return AvailableAuthTypeReply{
		Reply:   rout.Reply{Success: true},
		Methods: []enums.AuthProvider{enums.AuthProviderCredentials, enums.AuthProviderWebauthn},
	}
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

// RefreshRequest contains the refresh token used to obtain new access tokens
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" description:"The token to be used to refresh the access token after expiration"`
}

// RefreshReply contains new authentication tokens after successful refresh
type RefreshReply struct {
	rout.Reply
	Message string `json:"message,omitempty"`
	AuthData
}

// ExampleResponse returns an example RefreshReply for OpenAPI documentation
func (r *RefreshReply) ExampleResponse() any {
	return RefreshReply{
		Reply: rout.Reply{
			Success: true,
		},
		Message: "Token refreshed successfully",
		AuthData: AuthData{
			AccessToken:  "new_access_token",
			RefreshToken: "new_refresh_token",
			Session:      "session",
			TokenType:    "bearer",
		},
	}
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
// USERINFO
// =========

// UserInfoReply contains user information for authenticated requests
type UserInfoReply struct {
	rout.Reply
	ID              string  `json:"id" description:"The ID of the user" example:"01J4EXD5MM60CX4YNYN0DEE3Y1"`
	Email           string  `json:"email" description:"The email address of the user" example:"jsnow@example.com"`
	FirstName       string  `json:"first_name,omitempty" description:"The first name of the user" example:"Jon"`
	LastName        string  `json:"last_name,omitempty" description:"The last name of the user" example:"Snow"`
	DisplayName     string  `json:"display_name,omitempty" description:"The display name of the user" example:"Jon Snow"`
	AvatarRemoteURL *string `json:"avatar_remote_url,omitempty" description:"URL of the user's remote avatar" example:"https://example.com/avatar.jpg"`
	LastSeen        *string `json:"last_seen,omitempty" description:"The time the user was last seen" example:"2023-01-01T00:00:00Z"`
	Role            string  `json:"role,omitempty" description:"The user's role" example:"ADMIN"`
	Sub             string  `json:"sub" description:"The subject of the user JWT" example:"user123"`
}

// ExampleUserInfoSuccessResponse is an example of a successful userinfo response for OpenAPI documentation
var ExampleUserInfoSuccessResponse = UserInfoReply{
	Reply:           rout.Reply{Success: true},
	ID:              "01J4EXD5MM60CX4YNYN0DEE3Y1",
	Email:           "jsnow@example.com",
	FirstName:       "Jon",
	LastName:        "Snow",
	DisplayName:     "Jon Snow",
	AvatarRemoteURL: stringPtr("https://example.com/avatar.jpg"),
	LastSeen:        stringPtr("2023-01-01T00:00:00Z"),
	Role:            "ADMIN",
	Sub:             "user123",
}

// Helper function for string pointer
func stringPtr(s string) *string {
	return &s
}

// =========
// REGISTER
// =========

// RegisterRequest contains user registration information for creating new accounts
type RegisterRequest struct {
	FirstName string  `json:"first_name,omitempty" description:"The first name of the user" example:"Jon"`
	LastName  string  `json:"last_name,omitempty" description:"The last name of the user" example:"Snow"`
	Email     string  `json:"email,omitempty" description:"The email address of the user" example:"jsnow@example.com"`
	Password  string  `json:"password,omitempty" description:"The password to be used for authentication after registration" example:"Wint3rIsC0ming123!"`
	Token     *string `json:"token" description:"A newly invited user can use this to join a org as at the same time they are creating their account"`
}

// RegisterReply contains authentication tokens and user information after successful registration
type RegisterReply struct {
	rout.Reply
	ID      string `json:"user_id" description:"The ID of the user that was created" example:"01J4EXD5MM60CX4YNYN0DEE3Y1"`
	Email   string `json:"email" description:"The email address of the user" example:"jsnow@example.com"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty" exclude:"true"` // only used for requests against local development, excluded from OpenAPI documentation
}

// ExampleResponse returns an example RegisterReply for OpenAPI documentation
func (r *RegisterReply) ExampleResponse() any {
	return RegisterReply{
		Reply:   rout.Reply{Success: true},
		ID:      exampleULID("user"),
		Email:   "jsnow@example.com",
		Message: "User registered successfully",
	}
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
	Token:     stringPtr("invite_token_example"),
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
	NeedsSSO    bool   `json:"needs_sso,omitempty"`
	NeedsTFA    bool   `json:"needs_tfa,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

// ExampleResponse returns an example SwitchOrganizationReply for OpenAPI documentation
func (r *SwitchOrganizationReply) ExampleResponse() any {
	return SwitchOrganizationReply{
		Reply: rout.Reply{Success: true},
		AuthData: AuthData{
			AccessToken:  "new_access_token",
			RefreshToken: "new_refresh_token",
			Session:      "new_session",
			TokenType:    "bearer",
		},
	}
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
	TargetOrganizationID: exampleULID("organization"),
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

// VerifyRequest contains email verification token
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

// ExampleResponse returns an example VerifyReply for OpenAPI documentation
func (r *VerifyReply) ExampleResponse() any {
	return VerifyReply{
		Reply:   rout.Reply{Success: true},
		ID:      exampleULID("user"),
		Email:   "jsnow@example.com",
		Message: "Email verified successfully",
		AuthData: AuthData{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			Session:      "session",
			TokenType:    "bearer",
		},
	}
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
	ID:      exampleULID("user_alt"),
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
// FILEDOWNLOAD
// =========

type File = storagetypes.File

type FileDownload struct {
	ID    string `param:"id" description:"the file ID" example:"01J4HMNDSZCCQBTY93BF9CBF5D"`
	Token string `query:"token" description:"The token to be used to verify the email address, token is sent via email"`
}

// Validate ensures the required fields are set on the VerifyRequest request
func (r *FileDownload) Validate() error {
	if r.Token == "" {
		return rout.NewMissingRequiredFieldError("token")
	}

	return nil
}

// ExampleVerifySuccessRequest is an example of a successful verify request for OpenAPI documentation
var ExampleFileDownloadRequest = FileDownload{
	Token: "token",
}

// UploadFilesReply holds the fields that are sent on a response to the `/upload` endpoint
type FileDownloadReply struct {
	rout.Reply
	Message string `json:"message,omitempty"`
	File    File   `json:"file" description:"The files that were uploaded"`
}

// ExampleResponse returns an example UploadFilesReply for OpenAPI documentation
func (r *FileDownloadReply) ExampleResponse() any {
	return FileDownloadReply{
		Reply:   rout.Reply{Success: true},
		Message: "Files uploaded successfully",
		File:    File{OriginalName: "example1.pdf"}, // nolint:mnd
	}
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

// ExampleResponse returns an example ResendReply for OpenAPI documentation
func (r *ResendReply) ExampleResponse() any {
	return ResendReply{
		Reply:   rout.Reply{Success: true},
		Message: "Verification email resent successfully",
	}
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

// ExampleResponse returns an example ForgotPasswordReply for OpenAPI documentation
func (r *ForgotPasswordReply) ExampleResponse() any {
	return ForgotPasswordReply{
		Reply:   rout.Reply{Success: true},
		Message: "Password reset email sent successfully",
	}
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

// ExampleResponse returns an example ResetPasswordReply for OpenAPI documentation
func (r *ResetPasswordReply) ExampleResponse() any {
	return ResetPasswordReply{
		Reply:   rout.Reply{Success: true},
		Message: "Password reset successfully",
	}
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
	Name  string `json:"name,omitempty" description:"The name of the user" example:"Jon Snow"`
}

func (r *WebauthnRegistrationRequest) Validate() error {
	if r.Email == "" {
		return rout.NewMissingRequiredFieldError("email")
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

// ExampleResponse returns an example WebauthnBeginRegistrationResponse for OpenAPI documentation
func (r *WebauthnBeginRegistrationResponse) ExampleResponse() any {
	return WebauthnBeginRegistrationResponse{
		Reply:   rout.Reply{Success: true},
		Session: "registration_session_example",
	}
}

var ExampleWebauthnBeginRegistrationRequest = WebauthnRegistrationRequest{
	Email: "sarahisthebest@sarahsthebest.com",
	Name:  "Sarah Funk",
}

var ExampleWebauthnBeginRegistrationResponse = WebauthnBeginRegistrationResponse{
	Reply: rout.Reply{Success: true},
	CredentialCreation: &protocol.CredentialCreation{
		Response: protocol.PublicKeyCredentialCreationOptions{
			RelyingParty: protocol.RelyingPartyEntity{
				CredentialEntity: protocol.CredentialEntity{
					Name: "Openlane",
				},
				ID: "theopenlane.io",
			},
			User: protocol.UserEntity{
				CredentialEntity: protocol.CredentialEntity{
					Name: "Sarah Funk",
				},
				DisplayName: "sarahfunk",
				ID:          []byte("user-id-12345"),
			},
			Challenge: protocol.URLEncodedBase64("cmFuZG9tLWNoYWxsZW5nZS1zdHJpbmc="),
			Timeout:   60000, // nolint:mnd
		}},
	Session: "example-session-id",
}

// WebauthnRegistrationFinishRequest is the request to finish webauthn registration
// This represents the credential creation response from the browser's WebAuthn API
type WebauthnRegistrationFinishRequest struct {
	ID                      string                 `json:"id" description:"The credential ID"`
	RawID                   string                 `json:"rawId" description:"The raw credential ID"`
	Type                    string                 `json:"type" description:"The credential type, should be 'public-key'"`
	AuthenticatorAttachment string                 `json:"authenticatorAttachment,omitempty" description:"How the authenticator is attached"`
	ClientExtensionResults  map[string]interface{} `json:"clientExtensionResults,omitempty" description:"Extension results"`
	Response                struct {
		AttestationObject  string   `json:"attestationObject" description:"The attestation object"`
		ClientDataJSON     string   `json:"clientDataJSON" description:"The client data JSON"`
		PublicKey          string   `json:"publicKey,omitempty" description:"The public key"`
		PublicKeyAlgorithm int      `json:"publicKeyAlgorithm,omitempty" description:"The public key algorithm"`
		Transports         []string `json:"transports,omitempty" description:"Available transports"`
		AuthenticatorData  string   `json:"authenticatorData,omitempty" description:"The authenticator data"`
	} `json:"response" description:"The authenticator response"`
}

// ExampleWebauthnRegistrationFinishRequest is an example WebAuthn registration finish request for OpenAPI documentation
var ExampleWebauthnRegistrationFinishRequest = WebauthnRegistrationFinishRequest{
	ID:                      "JBqvfKoo_U-McTi9NxkpDTncmL2Lg6fczz6PD7WesCHQPg",
	RawID:                   "JBqvfKoo_U-McTi9NxkpDTncmL2Lg6fczz6PD7WesCHQPg",
	Type:                    "public-key",
	AuthenticatorAttachment: "platform",
	ClientExtensionResults:  map[string]interface{}{},
	Response: struct {
		AttestationObject  string   `json:"attestationObject" description:"The attestation object"`
		ClientDataJSON     string   `json:"clientDataJSON" description:"The client data JSON"`
		PublicKey          string   `json:"publicKey,omitempty" description:"The public key"`
		PublicKeyAlgorithm int      `json:"publicKeyAlgorithm,omitempty" description:"The public key algorithm"`
		Transports         []string `json:"transports,omitempty" description:"Available transports"`
		AuthenticatorData  string   `json:"authenticatorData,omitempty" description:"The authenticator data"`
	}{
		AttestationObject:  "o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YVimSZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2NdAAAAALraVWanqkAfvZZFYZpVEg0AIiQar3yqKP1PjHE4vTcZKQ053Ji9i4On3M8-jw-1nrAh0D6lAQIDJiABIVggldWfMY_HYjHCZuSgBcDj-Zqcnipy1SJVNlhvmZBxvpciWCDh1UJNz9Uyr6jqeQhApJ3krQCvDNoeXaH0ffa9KapYdw",
		ClientDataJSON:     "eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoiTWZlN1l6aS0zUU9rMDM4VHh3dVVvaTBaaURIZEdaOGlGNVhXc09UTTVnbyIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6MzAwMSIsImNyb3NzT3JpZ2luIjpmYWxzZX0",
		PublicKey:          "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEldWfMY_HYjHCZuSgBcDj-Zqcnipy1SJVNlhvmZBxvpfh1UJNz9Uyr6jqeQhApJ3krQCvDNoeXaH0ffa9KapYdw",
		PublicKeyAlgorithm: -7,
		Transports:         []string{"internal", "hybrid"},
		AuthenticatorData:  "SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2NdAAAAALraVWanqkAfvZZFYZpVEg0AIiQar3yqKP1PjHE4vTcZKQ053Ji9i4On3M8-jw-1nrAh0D6lAQIDJiABIVggldWfMY_HYjHCZuSgBcDj-Zqcnipy1SJVNlhvmZBxvpciWCDh1UJNz9Uyr6jqeQhApJ3krQCvDNoeXaH0ffa9KapYdw",
	},
}

// WebauthnRegistrationResponse is the response after a successful webauthn registration
type WebauthnRegistrationResponse struct {
	rout.Reply
	Message string `json:"message,omitempty"`
	AuthData
}

// ExampleResponse returns an example WebauthnRegistrationResponse for OpenAPI documentation
func (r *WebauthnRegistrationResponse) ExampleResponse() any {
	return WebauthnRegistrationResponse{
		Reply:   rout.Reply{Success: true},
		Message: "WebAuthn registration successful",
		AuthData: AuthData{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			Session:      "session",
			TokenType:    "bearer",
		},
	}
}

// WebauthnLoginRequest is the request to begin a webauthn login
type WebauthnLoginRequest struct {
	Email string `json:"email,omitempty" description:"The email address associated with the account" example:"jsnow@example.com"`
}

// ExampleWebauthnLoginRequest is an example WebAuthn login request for OpenAPI documentation
var ExampleWebauthnLoginRequest = WebauthnLoginRequest{
	Email: "",
}

// Validate ensures the required fields are set on the WebauthnLoginRequest request
func (r *WebauthnLoginRequest) Validate() error {
	// email is not required so there is not validation required here
	return nil
}

// WebauthnBeginLoginResponse is the response to begin a webauthn login
// this includes the credential assertion options and the session token
type WebauthnBeginLoginResponse struct {
	Reply rout.Reply
	*protocol.CredentialAssertion
	Session string `json:"session,omitempty"`
}

// WebauthnLoginFinishRequest is the request to finish webauthn login
// This represents the credential assertion response from the browser's WebAuthn API
type WebauthnLoginFinishRequest struct {
	ID                      string                 `json:"id" description:"The credential ID"`
	RawID                   string                 `json:"rawId" description:"The raw credential ID"`
	Type                    string                 `json:"type" description:"The credential type, should be 'public-key'"`
	AuthenticatorAttachment string                 `json:"authenticatorAttachment,omitempty" description:"How the authenticator is attached"`
	ClientExtensionResults  map[string]interface{} `json:"clientExtensionResults,omitempty" description:"Extension results"`
	Response                struct {
		AuthenticatorData string `json:"authenticatorData" description:"The authenticator data"`
		ClientDataJSON    string `json:"clientDataJSON" description:"The client data JSON"`
		Signature         string `json:"signature" description:"The assertion signature"`
		UserHandle        string `json:"userHandle,omitempty" description:"The user handle"`
	} `json:"response" description:"The authenticator response"`
}

// WebauthnLoginResponse is the response after a successful webauthn login
type WebauthnLoginResponse struct {
	rout.Reply
	Message string `json:"message,omitempty"`
	AuthData
}

// ExampleResponse returns an example WebauthnBeginLoginResponse for OpenAPI documentation
func (r *WebauthnBeginLoginResponse) ExampleResponse() any {
	return WebauthnBeginLoginResponse{
		Reply:   rout.Reply{Success: true},
		Session: "session123",
	}
}

// ExampleWebauthnBeginLoginResponse is an example WebAuthn begin login response for OpenAPI documentation
var ExampleWebauthnBeginLoginResponse = WebauthnBeginLoginResponse{
	Reply:   rout.Reply{Success: true},
	Session: "session123",
}

// ExampleResponse returns an example WebauthnLoginResponse for OpenAPI documentation
func (r *WebauthnLoginResponse) ExampleResponse() any {
	return WebauthnLoginResponse{
		Reply:   rout.Reply{Success: true},
		Message: "Authentication successful",
		AuthData: AuthData{
			AccessToken:  "access_token_here",
			RefreshToken: "refresh_token_here",
		},
	}
}

// ExampleWebauthnLoginFinishRequest is an example WebAuthn login finish request for OpenAPI documentation
var ExampleWebauthnLoginFinishRequest = WebauthnLoginFinishRequest{
	ID:                      "JBqvfKoo_U-McTi9NxkpDTncmL2Lg6fczz6PD7WesCHQPg",
	RawID:                   "JBqvfKoo_U-McTi9NxkpDTncmL2Lg6fczz6PD7WesCHQPg",
	Type:                    "public-key",
	AuthenticatorAttachment: "platform",
	ClientExtensionResults:  map[string]interface{}{},
	Response: struct {
		AuthenticatorData string `json:"authenticatorData" description:"The authenticator data"`
		ClientDataJSON    string `json:"clientDataJSON" description:"The client data JSON"`
		Signature         string `json:"signature" description:"The assertion signature"`
		UserHandle        string `json:"userHandle,omitempty" description:"The user handle"`
	}{
		AuthenticatorData: "SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2NdAAAAALraVWanqkAfvZZFYZpVEg0AIiQar3yqKP1PjHE4vTcZKQ053Ji9i4On3M8-jw-1nrAh0D6lAQIDJiABIVggldWfMY_HYjHCZuSgBcDj-Zqcnipy1SJVNlhvmZBxvpciWCDh1UJNz9Uyr6jqeQhApJ3krQCvDNoeXaH0ffa9KapYdw",
		ClientDataJSON:    "eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoiTWZlN1l6aS0zUU9rMDM4VHh3dVVvaTBaaURIZEdaOGlGNVhXc09UTTVnbyIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6MzAwMSIsImNyb3NzT3JpZ2luIjpmYWxzZX0",
		Signature:         "MEUCIQDKIueQAhZmGtPTmzp7QQRjZU_XLUqHdGj3QKRMOxRNbwIgF1hkJJ5y7cA3RGZe9x4n9vXq_L9x8eR1r9cE4w1uJ_A",
		UserHandle:        "dXNlci1pZC0xMjM0NQ",
	},
}

// ExampleWebauthnLoginResponse is an example WebAuthn login response for OpenAPI documentation
var ExampleWebauthnLoginResponse = WebauthnLoginResponse{
	Reply:   rout.Reply{Success: true},
	Message: "Authentication successful",
	AuthData: AuthData{
		AccessToken:  "access_token_here",
		RefreshToken: "refresh_token_here",
	},
}

// ExampleWebauthnRegistrationResponse is an example WebAuthn registration response for OpenAPI documentation
var ExampleWebauthnRegistrationResponse = WebauthnRegistrationResponse{
	Reply:   rout.Reply{Success: true},
	Message: "Registration successful",
	AuthData: AuthData{
		AccessToken:  "access_token_here",
		RefreshToken: "refresh_token_here",
	},
}

// =========
// SUBSCRIBER VERIFY
// =========

// VerifySubscribeRequest contains subscription verification information
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

// InviteRequest contains invitation token for organization membership
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
	NeedsSSO    bool   `json:"needs_sso,omitempty"`
	AuthData
}

// ExampleResponse returns an example InviteReply for OpenAPI documentation
func (r *InviteReply) ExampleResponse() any {
	return InviteReply{
		Reply:       rout.Reply{Success: true},
		ID:          exampleULID("user"),
		Email:       "jsnow@example.com",
		Message:     "Invitation accepted successfully",
		JoinedOrgID: exampleULID("joined_org"),
		Role:        "admin",
		AuthData: AuthData{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			Session:      "session",
			TokenType:    "bearer",
		},
	}
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
	OrgID            string `json:"org_id,omitempty" description:"the organization id for the sso connection"`
}

// ExampleOauthTokenRequest is an example OAuth token request for OpenAPI documentation
var ExampleOauthTokenRequest = OauthTokenRequest{
	Name:             "Jon Snow",
	Email:            "jsnow@example.com",
	AuthProvider:     "google",
	ExternalUserID:   "1234567890",
	ExternalUserName: "jsnow",
	ClientToken:      "example-client-token-12345",
	Image:            "https://example.com/avatar.jpg",
}

// =========
// ACCOUNT/ACCESS
// =========

// AccountAccessRequest contains organization ID for checking access permissions
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

// ExampleResponse returns an example AccountAccessReply for OpenAPI documentation
func (r *AccountAccessReply) ExampleResponse() any {
	return AccountAccessReply{
		Reply:   rout.Reply{Success: true},
		Allowed: true,
	}
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

// AccountRolesRequest contains object IDs for retrieving roles associated with them
type AccountRolesRequest struct {
	// @deprecated use ObjectIDs instead, may be removed in a future release
	ObjectID    string   `json:"object_id,omitempty" description:" @deprecated use ObjectIDs instead. The ID of the object to check roles for" example:"01J4EXD5MM60CX4YNYN0DEE3Y1"`
	ObjectIDs   []string `json:"object_ids,omitempty" description:"The IDs of the object to check roles for, can be used to check multiple ids in one request"` // example:"["01J4EXD5MM60CX4YNYN0DEE3Y1", "01J4EXD5MM60CX4YNYN0DEE3Y2"]"
	ObjectType  string   `json:"object_type" description:"The type of object to check roles for, e.g. organization, program, procedure, etc" example:"organization"`
	SubjectType string   `json:"subject_type,omitempty" description:"The type of subject to check roles for, e.g. service, user" example:"user"`
	Relations   []string `json:"relations,omitempty" description:"The relations to check roles for, e.g. can_view, can_edit"`
}

// AccountRolesReply holds the fields that are sent on a response to the `/account/roles` endpoint
type AccountRolesReply struct {
	rout.Reply
	// Roles is a list of roles the user has for the specified object(s)
	// @deprecated use ObjectRoles instead, may be removed in a future release
	Roles []string `json:"roles" description:" @deprecated use ObjectRoles instead. A list of roles the subject has for the specified object"`
	// ObjectRoles is a map of object IDs to the roles the user has for each object ID
	ObjectRoles map[string][]string `json:"object_roles,omitempty" description:"A map of object IDs to the roles the subject has for each object ID"`
}

// ExampleResponse returns an example AccountRolesReply for OpenAPI documentation
func (r *AccountRolesReply) ExampleResponse() any {
	return AccountRolesReply{
		Reply: rout.Reply{Success: true},
		Roles: []string{"admin", "member"},
	}
}

// Validate ensures the required fields are set on the AccountAccessRequest
func (r *AccountRolesRequest) Validate() error {
	if r.ObjectID == "" && len(r.ObjectIDs) == 0 {
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

// AccountRolesOrganizationRequest contains organization ID for retrieving organization-specific roles
type AccountRolesOrganizationRequest struct {
	ID string `param:"id" description:"The ID of the organization to check roles for" example:"01J4HMNDSZCCQBTY93BF9CBF5D"`
}

// AccountRolesOrganizationReply holds the fields that are sent on a response to the `/account/roles/organization` endpoint
type AccountRolesOrganizationReply struct {
	rout.Reply
	Roles          []string `json:"roles" description:"The roles the user has in the organization, e.g. can_view, can_edit"`
	OrganizationID string   `json:"organization_id" description:"The ID of the organization the user has roles in" example:"01J4HMNDSZCCQBTY93BF9CBF5D"`
}

// ExampleResponse returns an example AccountRolesOrganizationReply for OpenAPI documentation
func (r *AccountRolesOrganizationReply) ExampleResponse() any {
	return AccountRolesOrganizationReply{
		Reply:          rout.Reply{Success: true},
		Roles:          []string{"can_view", "can_edit"},
		OrganizationID: exampleULID("organization"),
	}
}

// Validate ensures the required fields are set on the AccountRolesOrganizationRequest
func (r *AccountRolesOrganizationRequest) Validate() error {
	// ID is optional - if empty, handler will get it from auth context
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

// AccountFeaturesRequest contains organization ID for retrieving available features
type AccountFeaturesRequest struct {
	ID string `param:"id" description:"The ID of the organization to check roles for" example:"01J4HMNDSZCCQBTY93BF9CBF5D"`
}

// AccountFeaturesReply holds the fields that are sent on a response to the `/account/features` endpoint
type AccountFeaturesReply struct {
	rout.Reply
	Features       []string `json:"features" description:"The features the user has access to in the organization, e.g. policy-and-procedure-module, compliance-module"`
	OrganizationID string   `json:"organization_id" description:"The ID of the organization the user has features in" example:"01J4HMNDSZCCQBTY93BF9CBF5D"`
}

// ExampleResponse returns an example AccountFeaturesReply for OpenAPI documentation
func (r *AccountFeaturesReply) ExampleResponse() any {
	return AccountFeaturesReply{
		Reply:          rout.Reply{Success: true},
		Features:       []string{"policy-and-procedure-module", "compliance-module"},
		OrganizationID: exampleULID("organization"),
	}
}

// Validate ensures the required fields are set on the AccountFeaturesRequest
func (r *AccountFeaturesRequest) Validate() error {
	// ID is optional - if empty, handler will get it from auth context
	return nil
}

// ExampleAccountFeaturesRequest is an example of a successful `/account/features` request for OpenAPI documentation
var ExampleAccountFeaturesRequest = AccountFeaturesRequest{
	ID: "01J4HMNDSZCCQBTY93BF9CBF5D",
}

// ExampleAccountFeaturesReply is an example of a successful `/account/features` response for OpenAPI documentation
var ExampleAccountFeaturesReply = AccountFeaturesReply{
	Reply: rout.Reply{Success: true},
	Features: []string{
		"policy-and-procedure-module",
		"centralized-audit-documentation",
		"risk-management",
		"compliance-module",
	},
	OrganizationID: "01J4HMNDSZCCQBTY93BF9CBF5D",
}

// =========
// FILES
// =========

// UploadFilesRequest contains file upload data and metadata
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

// ExampleResponse returns an example UploadFilesReply for OpenAPI documentation
func (r *UploadFilesReply) ExampleResponse() any {
	return UploadFilesReply{
		Reply:     rout.Reply{Success: true},
		Message:   "Files uploaded successfully",
		FileCount: 2, // nolint:mnd
		Files: []File{
			{OriginalName: "example1.pdf"}, // nolint:mnd
			{OriginalName: "example2.jpg"}, // nolint:mnd
		},
	}
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
			OriginalName: "file.txt",
			MD5:          []byte("1234"),
			CreatedAt:    exampleTime(-time.Hour),
			UpdatedAt:    exampleTime(0),
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

// ExampleResponse returns an example TFAReply for OpenAPI documentation
func (r *TFAReply) ExampleResponse() any {
	return TFAReply{
		Reply:   rout.Reply{Success: true},
		Message: "Two-factor authentication validated successfully",
	}
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

// JobRunnerRegistrationReply is the response to begin a job runner registration
// this includes the credential creation options and the session token
type JobRunnerRegistrationReply struct {
	Reply   rout.Reply
	Message string `json:"message"`
}

// ExampleResponse returns an example JobRunnerRegistrationReply for OpenAPI documentation
func (r *JobRunnerRegistrationReply) ExampleResponse() any {
	return JobRunnerRegistrationReply{
		Reply:   rout.Reply{Success: true},
		Message: "Job runner registered successfully",
	}
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

// ExampleAcmeSolverRequest is an example ACME solver request for OpenAPI documentation
var ExampleAcmeSolverRequest = AcmeSolverRequest{
	Path: "01J4HMNDSZCCQBTY93BF9CBF5D",
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
	IsTest         bool   `json:"is_test" query:"is_test" description:"Used when testing the sso was successfully connected"`
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
	OrganizationID: exampleULID("organization"),
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
	OrganizationID: exampleULID("organization"),
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

// SSOLoginReply is the response for the SSO login
type SSOLoginReply struct {
	rout.Reply
	RedirectURI string `json:"redirect_uri,omitempty"`
}

// SSOStatusReply is the response for SSOStatusRequest
type SSOStatusReply struct {
	rout.Reply
	Enforced       bool              `json:"enforced"`
	Provider       enums.SSOProvider `json:"provider,omitempty"`
	DiscoveryURL   string            `json:"discovery_url,omitempty"`
	SAMLSignInURL  string            `json:"saml_signin_url,omitempty"`
	OrganizationID string            `json:"organization_id,omitempty"`
	OrgTFAEnforced bool              `json:"tfa_enforced"`
	UserTFAEnabled bool              `json:"user_tfa_enabled,omitempty"`
	IsOrgOwner     bool              `json:"is_org_owner,omitempty"`
}

// ExampleResponse returns an example SSOStatusReply for OpenAPI documentation
func (r *SSOStatusReply) ExampleResponse() any {
	return SSOStatusReply{
		Reply:          rout.Reply{Success: true},
		Enforced:       true,
		Provider:       enums.SSOProviderOkta,
		DiscoveryURL:   "https://accounts.example.com/.well-known/openid_configuration",
		SAMLSignInURL:  "https://accounts.example.com/saml/signin",
		OrganizationID: exampleULID("organization"),
		OrgTFAEnforced: true,
		UserTFAEnabled: false,
	}
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
	SAMLSignInURL:  "https://id.example.com/saml/signin",
	OrganizationID: exampleULID("organization_alt"),
	OrgTFAEnforced: true,
	UserTFAEnabled: false,
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

// ExampleResponse returns an example SSOTokenAuthorizeReply for OpenAPI documentation
func (r *SSOTokenAuthorizeReply) ExampleResponse() any {
	return SSOTokenAuthorizeReply{
		Reply:          rout.Reply{Success: true},
		OrganizationID: exampleULID("organization"),
		TokenID:        exampleULID("token"),
		Message:        "Token authorized successfully",
	}
}

// ExampleSSOTokenAuthorizeRequest is an example request for OpenAPI documentation
var ExampleSSOTokenAuthorizeRequest = SSOTokenAuthorizeRequest{
	OrganizationID: exampleULID("organization"),
	TokenID:        exampleULID("token"),
	TokenType:      "api",
}

// ExampleSSOTokenAuthorizeReply is an example response for OpenAPI documentation
var ExampleSSOTokenAuthorizeReply = SSOTokenAuthorizeReply{
	Reply:          rout.Reply{Success: true},
	OrganizationID: exampleULID("organization"),
	TokenID:        exampleULID("token"),
	Message:        "success",
}

// CreateTrustCenterAnonymousJWTResponse is the response to a request to create a trust center anonymous JWT
type CreateTrustCenterAnonymousJWTResponse struct {
	AuthData
}

// ExampleResponse returns an example CreateTrustCenterAnonymousJWTResponse for OpenAPI documentation
func (r *CreateTrustCenterAnonymousJWTResponse) ExampleResponse() any {
	return CreateTrustCenterAnonymousJWTResponse{
		AuthData: AuthData{
			AccessToken: "anonymous_jwt_token",
			TokenType:   "bearer",
		},
	}
}

// ExampleCreateTrustCenterAnonymousJWTResponse is an example trust center anonymous JWT response for OpenAPI documentation
var ExampleCreateTrustCenterAnonymousJWTResponse = CreateTrustCenterAnonymousJWTResponse{
	AuthData: AuthData{
		AccessToken:  "access_token_here",
		RefreshToken: "refresh_token_here",
	},
}

// GetQuestionnaireResponse is the response containing the questionnaire template's JSON configuration
type GetQuestionnaireResponse struct {
	Jsonconfig map[string]any `json:"jsonconfig,omitempty"`
	UISchema   map[string]any `json:"uischema,omitempty"`
}

// ExampleResponse returns an example GetQuestionnaireResponse for OpenAPI documentation
func (r *GetQuestionnaireResponse) ExampleResponse() any {
	return GetQuestionnaireResponse{
		Jsonconfig: map[string]any{
			"title":       "Sample Questionnaire",
			"description": "A sample questionnaire template",
			"questions": []map[string]any{
				{
					"id":       "q1",
					"question": "Sample question",
					"type":     "text",
				},
			},
		},
	}
}

// ExampleGetQuestionnaireResponse is an example questionnaire response for OpenAPI documentation
var ExampleGetQuestionnaireResponse = GetQuestionnaireResponse{
	Jsonconfig: map[string]any{
		"title":       "Sample Questionnaire",
		"description": "A sample questionnaire template",
	},
}

// SubmitQuestionnaireRequest is the request to submit questionnaire response data
type SubmitQuestionnaireRequest struct {
	Data map[string]any `json:"data" binding:"required"`
}

// ExampleSubmitQuestionnaireRequest is an example questionnaire submission request for OpenAPI documentation
var ExampleSubmitQuestionnaireRequest = SubmitQuestionnaireRequest{
	Data: map[string]any{
		"q1": "Answer to question 1",
		"q2": "Answer to question 2",
		"q3": map[string]any{
			"nested": "data",
		},
	},
}

// SubmitQuestionnaireResponse is the response after successfully submitting questionnaire data
type SubmitQuestionnaireResponse struct {
	DocumentDataID string `json:"document_data_id"`
	Status         string `json:"status"`
	CompletedAt    string `json:"completed_at"`
}

// ExampleResponse returns an example SubmitQuestionnaireResponse for OpenAPI documentation
func (r *SubmitQuestionnaireResponse) ExampleResponse() any {
	return SubmitQuestionnaireResponse{
		DocumentDataID: "document_data_id_here",
		Status:         "COMPLETED",
		CompletedAt:    "2024-01-01T00:00:00Z",
	}
}

// ExampleSubmitQuestionnaireResponse is an example questionnaire submission response for OpenAPI documentation
var ExampleSubmitQuestionnaireResponse = SubmitQuestionnaireResponse{
	DocumentDataID: "01JCQR8Z9X1A2B3C4D5E6F7G8H",
	Status:         "COMPLETED",
	CompletedAt:    "2024-01-01T12:00:00Z",
}

// =================
// IMPERSONATION
// =================
//

const (
	// MinImpersonationReasonLength is the minimum length for impersonation reason
	MinImpersonationReasonLength = 10
	// MaxImpersonationReasonLength is the maximum length for impersonation reason
	MaxImpersonationReasonLength = 500
)

// StartImpersonationRequest represents a request to start impersonating a user
type StartImpersonationRequest struct {
	TargetUserID   string   `json:"target_user_id" validate:"required" description:"The ID of the user to impersonate"`
	Type           string   `json:"type" validate:"required,oneof=support job admin" description:"The type of impersonation (support, job, admin)"`
	Reason         string   `json:"reason" validate:"required,min=10,max=500" description:"Reason for the impersonation"`
	Duration       *int     `json:"duration_hours,omitempty" description:"Duration in hours (optional, defaults to 1 hour)"`
	Scopes         []string `json:"scopes,omitempty" description:"Specific scopes for the impersonation session"`
	OrganizationID string   `json:"organization_id,omitempty" description:"Organization context for impersonation"`
}

// StartImpersonationReply represents the response when starting impersonation
type StartImpersonationReply struct {
	rout.Reply
	Token     string    `json:"token" description:"The impersonation token"`
	ExpiresAt time.Time `json:"expires_at" description:"When the impersonation token expires"`
	SessionID string    `json:"session_id" description:"The impersonation session ID"`
	Message   string    `json:"message" description:"Success message"`
}

// ExampleResponse returns an example StartImpersonationReply for OpenAPI documentation
func (r *StartImpersonationReply) ExampleResponse() any {
	return StartImpersonationReply{
		Reply:     rout.Reply{Success: true},
		Token:     "impersonation_token_example",
		ExpiresAt: exampleTime(time.Hour),
		SessionID: exampleULID("session"),
		Message:   "Impersonation session started successfully",
	}
}

// EndImpersonationRequest represents a request to end an impersonation session
type EndImpersonationRequest struct {
	SessionID string `json:"session_id" validate:"required" description:"The session ID to end"`
	Reason    string `json:"reason,omitempty" description:"Optional reason for ending the session"`
}

// EndImpersonationReply represents the response when ending impersonation
type EndImpersonationReply struct {
	rout.Reply
	Message string `json:"message" description:"Success message"`
}

// ExampleResponse returns an example EndImpersonationReply for OpenAPI documentation
func (r *EndImpersonationReply) ExampleResponse() any {
	return EndImpersonationReply{
		Reply:   rout.Reply{Success: true},
		Message: "Impersonation session ended successfully",
	}
}

// Validate ensures the required fields are set on the StartImpersonationRequest
func (r *StartImpersonationRequest) Validate() error {
	r.TargetUserID = strings.TrimSpace(r.TargetUserID)
	r.Type = strings.TrimSpace(r.Type)
	r.Reason = strings.TrimSpace(r.Reason)
	r.OrganizationID = strings.TrimSpace(r.OrganizationID)

	switch {
	case r.TargetUserID == "":
		return rout.NewMissingRequiredFieldError("target_user_id")
	case r.Type == "":
		return rout.NewMissingRequiredFieldError("type")
	case r.Reason == "":
		return rout.NewMissingRequiredFieldError("reason")
	case len(r.Reason) < MinImpersonationReasonLength:
		return rout.InvalidField("reason must be at least 10 characters")
	case len(r.Reason) > MaxImpersonationReasonLength:
		return rout.InvalidField("reason must be less than 500 characters")
	}

	return nil
}

// Validate ensures the required fields are set on the EndImpersonationRequest
func (r *EndImpersonationRequest) Validate() error {
	r.SessionID = strings.TrimSpace(r.SessionID)
	r.Reason = strings.TrimSpace(r.Reason)

	if r.SessionID == "" {
		return rout.NewMissingRequiredFieldError("session_id")
	}

	return nil
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

// ExampleResponse returns an example IntegrationTokenResponse for OpenAPI documentation
func (r *IntegrationTokenResponse) ExampleResponse() any {
	expiresAt := exampleTime(30 * 24 * time.Hour) // nolint:mnd

	return IntegrationTokenResponse{
		Reply:    rout.Reply{Success: true},
		Provider: "github",
		Token: &IntegrationToken{
			AccessToken:  "ghr_example_token",
			RefreshToken: "ghr_example_refresh_token",
		},
		ExpiresAt: &expiresAt,
	}
}

// ListIntegrationsResponse is the response for listing integrations
type ListIntegrationsResponse struct {
	rout.Reply
	Integrations any `json:"integrations"` // Will be []*ent.Integration
}

// DeleteIntegrationResponse is the response for deleting an integration
type DeleteIntegrationResponse struct {
	rout.Reply
	Message   string `json:"message"`
	DeletedID string `json:"deletedId,omitempty"`
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

// ExampleResponse returns an example IntegrationStatusResponse for OpenAPI documentation
func (r *IntegrationStatusResponse) ExampleResponse() any {
	return IntegrationStatusResponse{
		Reply:        rout.Reply{Success: true},
		Provider:     "github",
		Connected:    true,
		Status:       "connected",
		TokenValid:   true,
		TokenExpired: false,
		Message:      "Integration status retrieved successfully",
		Integration:  map[string]any{"id": exampleULID("integration"), "name": "GitHub Integration"},
	}
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

// ExampleRefreshIntegrationTokenRequest is an example refresh integration token request for OpenAPI documentation
var ExampleRefreshIntegrationTokenRequest = RefreshIntegrationTokenRequest{
	Provider: "github",
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
	Scopes      []string `json:"scopes,omitempty" description:"Additional OAuth scopes to request"` // example: ["repo", "gist"]
}

// Validate ensures the required fields are set on the OAuthFlowRequest
func (r *OAuthFlowRequest) Validate() error {
	r.Provider = strings.TrimSpace(strings.ToLower(r.Provider))

	if r.Provider == "" {
		return rout.NewMissingRequiredFieldError("provider")
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
	RequiresLogin bool   `json:"requiresLogin,omitempty" description:"Whether user needs to login before OAuth flow"`
}

// ExampleResponse returns an example OAuthFlowResponse for OpenAPI documentation
func (r *OAuthFlowResponse) ExampleResponse() any {
	return OAuthFlowResponse{
		Reply:         rout.Reply{Success: true},
		AuthURL:       "https://github.com/login/oauth/authorize?client_id=example&state=eyJvcmdJRCI6IjAxSE",
		State:         "eyJvcmdJRCI6IjAxSE",
		Message:       "",
		RequiresLogin: false,
	}
}

// OAuthCallbackRequest represents the OAuth callback data
type OAuthCallbackRequest struct {
	Provider string `json:"provider,omitempty" query:"provider" description:"OAuth provider (extracted from state)"`
	Code     string `json:"code" query:"code" description:"OAuth authorization code"`
	State    string `json:"state" query:"state" description:"OAuth state parameter"`
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

// ExampleStartImpersonationRequest is an example request for OpenAPI documentation
var ExampleStartImpersonationRequest = StartImpersonationRequest{
	TargetUserID: exampleULID("user_alt"),
	Type:         "support",
	Reason:       "Customer support assistance for account recovery",
	Duration:     nil, // Use default
}

// ExampleStartImpersonationReply is an example response for OpenAPI documentation
var ExampleStartImpersonationReply = StartImpersonationReply{
	Reply:     rout.Reply{Success: true},
	Token:     "imp_" + exampleULID("token"),
	ExpiresAt: exampleTime(time.Hour),
	SessionID: exampleULID("session"),
	Message:   "Impersonation session started successfully",
}

// ExampleEndImpersonationRequest is an example request for OpenAPI documentation
var ExampleEndImpersonationRequest = EndImpersonationRequest{
	SessionID: exampleULID("session"),
	Reason:    "Support task completed",
}

// ExampleEndImpersonationReply is an example response for OpenAPI documentation
var ExampleEndImpersonationReply = EndImpersonationReply{
	Reply:   rout.Reply{Success: true},
	Message: "Impersonation session ended successfully",
}

// OAuthCallbackResponse contains the result of OAuth callback processing
type OAuthCallbackResponse struct {
	rout.Reply
	Success     bool   `json:"success" description:"Whether the OAuth callback was processed successfully"`
	Integration any    `json:"integration,omitempty" description:"The created/updated integration object"`
	Message     string `json:"message" description:"Success or error message" example:"Successfully connected GitHub integration"`
}

// ExampleResponse returns an example OAuthCallbackResponse for OpenAPI documentation
func (r *OAuthCallbackResponse) ExampleResponse() any {
	return OAuthCallbackResponse{
		Reply:       rout.Reply{Success: true},
		Success:     true,
		Integration: map[string]any{"id": exampleULID("integration"), "provider": "github", "status": "connected"},
		Message:     "Successfully connected GitHub integration",
	}
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

// =========
// STRIPE WEBHOOK
// =========

// StripeWebhookRequest contains the query parameters for Stripe webhook requests
type StripeWebhookRequest struct {
	APIVersion string `query:"api_version" description:"Stripe API version for this webhook request" example:"2024-11-20.acacia"`
}

// Validate ensures the StripeWebhookRequest is valid
func (r *StripeWebhookRequest) Validate() error {
	// API version is optional, no validation required
	return nil
}

// =========
// PRODUCTS
// =========

// ProductCatalogRequest
type ProductCatalogRequest struct {
	IncludeBeta    bool `query:"include_beta" description:"Whether to include beta products in the catalog" example:"false"`
	IncludePrivate bool `query:"include_private" description:"Whether to include private products in the catalog" example:"false"`
}

// ProductCatalogReply holds the fields that are sent on a response to the `/products` endpoint
type ProductCatalogReply struct {
	rout.Reply
	catalog.Catalog
}

// ExampleResponse returns an example ProductCatalogReply for OpenAPI documentation
func (r *ProductCatalogReply) ExampleResponse() any {
	return ExampleProductCatalogReply
}

// Validate ensures the required fields are set on the ProductCatalogRequest
func (r *ProductCatalogRequest) Validate() error {
	// all fields are optional, if none are set only public proucts are returned
	return nil
}

// ExampleProductCatalogRequest is an example of a successful `/products` request for OpenAPI documentation
var ExampleProductCatalogRequest = ProductCatalogRequest{
	IncludeBeta:    false,
	IncludePrivate: false,
}

// ExampleProductCatalogReply is an example of a successful `/products` response for OpenAPI documentation
var ExampleProductCatalogReply = ProductCatalogReply{
	Reply:   rout.Reply{Success: true},
	Catalog: gencatalog.DefaultSandboxCatalog,
}

// IntegrationConfigRequest represents arbitrary credential configuration submitted for a provider.
type IntegrationConfigRequest struct {
	ServiceAccountEmail      string         `json:"serviceAccountEmail,omitempty"`
	Audience                 string         `json:"audience,omitempty"`
	ProjectID                string         `json:"projectId,omitempty"`
	OrganizationID           string         `json:"organizationId,omitempty"`
	WorkloadIdentityProvider string         `json:"workloadIdentityProvider,omitempty"`
	FindingFilter            string         `json:"findingFilter,omitempty"`
	Additional               map[string]any `json:"-"`
}

// ExampleIntegrationConfigRequest is an example configuration payload for OpenAPI documentation.
var ExampleIntegrationConfigRequest = IntegrationConfigRequest{
	ServiceAccountEmail:      "scc-runner@example.iam.gserviceaccount.com",
	Audience:                 "//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/pool/providers/provider",
	ProjectID:                "sample-project",
	OrganizationID:           "1234567890",
	WorkloadIdentityProvider: "projects/123456789/locations/global/workloadIdentityPools/pool/providers/provider",
}

// UnmarshalJSON captures known fields and preserves additional properties.
func (r *IntegrationConfigRequest) UnmarshalJSON(data []byte) error {
	type Alias IntegrationConfigRequest
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*r = IntegrationConfigRequest(alias)

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	known := map[string]struct{}{
		"serviceAccountEmail":      {},
		"audience":                 {},
		"projectId":                {},
		"organizationId":           {},
		"workloadIdentityProvider": {},
		"findingFilter":            {},
	}
	r.Additional = make(map[string]any)
	for key, value := range raw {
		if _, ok := known[key]; ok {
			continue
		}
		r.Additional[key] = value
	}

	return nil
}

// ToMap flattens the request into a map for schema validation and keystore.
func (r IntegrationConfigRequest) ToMap() map[string]any {
	base := map[string]any{
		"serviceAccountEmail":      r.ServiceAccountEmail,
		"audience":                 r.Audience,
		"projectId":                r.ProjectID,
		"organizationId":           r.OrganizationID,
		"workloadIdentityProvider": r.WorkloadIdentityProvider,
		"findingFilter":            r.FindingFilter,
	}

	nonEmpty := lo.PickBy(base, func(_ string, value any) bool {
		if str, ok := value.(string); ok {
			return strings.TrimSpace(str) != ""
		}
		return value != nil
	})

	additional := lo.PickBy(r.Additional, func(_ string, value any) bool {
		if str, ok := value.(string); ok {
			return strings.TrimSpace(str) != ""
		}
		return value != nil
	})

	if len(additional) == 0 {
		return nonEmpty
	}

	return lo.Assign(nonEmpty, additional)
}

// IntegrationConfigParams captures path parameters for the integration config endpoint.
type IntegrationConfigParams struct {
	Provider string `param:"provider" description:"Integration provider identifier" example:"gcp_scc"`
}

// ExampleIntegrationConfigParams is an example of the path parameters for integration configuration.
var ExampleIntegrationConfigParams = IntegrationConfigParams{
	Provider: "gcp_scc",
}

// IntegrationConfigPayload wraps path parameters with the request payload.
type IntegrationConfigPayload struct {
	IntegrationConfigParams
	Body IntegrationConfigRequest `json:"payload"`
}

// IntegrationOperationParams captures path parameters for operation requests.
type IntegrationOperationParams struct {
	Provider string `param:"provider" description:"Integration provider identifier" example:"gcp_scc"`
}

// IntegrationOperationRequest describes a provider operation to run.
type IntegrationOperationRequest struct {
	Operation string         `json:"operation" validate:"required"`
	Config    map[string]any `json:"config,omitempty"`
	Force     bool           `json:"force,omitempty"`
}

// IntegrationOperationPayload wraps the params with the operation body.
type IntegrationOperationPayload struct {
	IntegrationOperationParams
	Body IntegrationOperationRequest `json:"payload"`
}

// ExampleIntegrationOperationPayload demonstrates a sample operation request.
var ExampleIntegrationOperationPayload = IntegrationOperationPayload{
	IntegrationOperationParams: IntegrationOperationParams{Provider: "gcp_scc"},
	Body: IntegrationOperationRequest{
		Operation: "findings.collect",
		Config: map[string]any{
			"sourceId": "organizations/123/sources/456",
			"filter":   `severity="HIGH"`,
		},
		Force: true,
	},
}

// IntegrationOperationMetadata describes an operation published by a provider.
type IntegrationOperationMetadata struct {
	Name         string         `json:"name"`
	Kind         string         `json:"kind"`
	Description  string         `json:"description,omitempty"`
	Client       string         `json:"client,omitempty"`
	ConfigSchema map[string]any `json:"configSchema,omitempty"`
}

// IntegrationProviderMetadata describes the data required for rendering integration forms.
type IntegrationProviderMetadata struct {
	Name                   string                               `json:"name"`
	DisplayName            string                               `json:"displayName"`
	Category               string                               `json:"category"`
	AuthType               keystore.AuthType                    `json:"authType"`
	Active                 bool                                 `json:"active"`
	LogoURL                string                               `json:"logoUrl,omitempty"`
	DocsURL                string                               `json:"docsUrl,omitempty"`
	OAuth                  *IntegrationOAuthMetadata            `json:"oauth,omitempty"`
	GoogleWorkloadIdentity *keystore.GoogleWorkloadIdentitySpec `json:"workloadIdentity,omitempty"`
	GitHubApp              *keystore.GitHubAppSpec              `json:"githubApp,omitempty"`
	Persistence            *keystore.PersistenceSpec            `json:"persistence,omitempty"`
	CredentialsSchema      map[string]any                       `json:"credentialsSchema,omitempty"`
	Labels                 map[string]string                    `json:"labels,omitempty"`
	Operations             []IntegrationOperationMetadata       `json:"operations,omitempty"`
}

// ExampleIntegrationConfigPayload demonstrates a full integration configuration request.
var ExampleIntegrationConfigPayload = IntegrationConfigPayload{
	IntegrationConfigParams: IntegrationConfigParams{Provider: "gcp_scc"},
	Body: IntegrationConfigRequest{
		ServiceAccountEmail:      "scc-runner@example.iam.gserviceaccount.com",
		Audience:                 "//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/pool/providers/provider",
		ProjectID:                "sample-project",
		OrganizationID:           "1234567890",
		WorkloadIdentityProvider: "projects/123456789/locations/global/workloadIdentityPools/pool/providers/provider",
		Additional: map[string]any{
			"serviceAccountKey": "{ \"type\": \"service_account\", ... }",
		},
	},
}

// IntegrationOAuthMetadata captures OAuth-specific metadata for integrations.
type IntegrationOAuthMetadata struct {
	AuthURL     string            `json:"authUrl,omitempty"`
	TokenURL    string            `json:"tokenUrl,omitempty"`
	RedirectURI string            `json:"redirectUri,omitempty"`
	Scopes      []string          `json:"scopes,omitempty"`
	UsePKCE     bool              `json:"usePkce,omitempty"`
	AuthParams  map[string]string `json:"authParams,omitempty"`
	TokenParams map[string]string `json:"tokenParams,omitempty"`
}

// IntegrationProvidersResponse is returned by the provider metadata endpoint.
type IntegrationProvidersResponse struct {
	rout.Reply
	Schema    *jsonschema.Schema            `json:"schema"`
	Providers []IntegrationProviderMetadata `json:"providers"`
}

// IntegrationConfigResponse is returned after persisting provider configuration.
type IntegrationConfigResponse struct {
	rout.Reply
	Provider string `json:"provider"`
}

// IntegrationOperationResponse reports the result of a provider operation.
type IntegrationOperationResponse struct {
	rout.Reply
	Provider  string         `json:"provider"`
	Operation string         `json:"operation"`
	Status    string         `json:"status"`
	Summary   string         `json:"summary"`
	Details   map[string]any `json:"details,omitempty"`
}

// ExampleResponse returns an example IntegrationConfigResponse for OpenAPI documentation.
func (r *IntegrationConfigResponse) ExampleResponse() any {
	return IntegrationConfigResponse{
		Reply:    rout.Reply{Success: true},
		Provider: "gcp_scc",
	}
}

// ExampleResponse returns a sample IntegrationOperationResponse.
func (r *IntegrationOperationResponse) ExampleResponse() any {
	return IntegrationOperationResponse{
		Reply:     rout.Reply{Success: true},
		Provider:  "gcp_scc",
		Operation: "findings.collect",
		Status:    "ok",
		Summary:   "Collected 5 findings from organizations/123/sources/456",
		Details: map[string]any{
			"totalFindings": exampleFindingsCount,
		},
	}
}

// DisconnectIntegrationRequest is the request payload for disconnecting an integration
type DisconnectIntegrationRequest struct {
	Provider      string `param:"provider" description:"Integration provider key" example:"github"`
	IntegrationID string `query:"integration_id,omitempty" description:"Specific integration ID to delete"`
}

// ExampleDisconnectIntegrationRequest provides an example disconnect request for OpenAPI documentation
var ExampleDisconnectIntegrationRequest = DisconnectIntegrationRequest{
	Provider: "github",
}
