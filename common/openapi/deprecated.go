package openapi

import "github.com/theopenlane/utils/rout"

// Deprecated aliases for the *Reply to *Response rename, kept so released consumers
// (github.com/theopenlane/go-client) continue to compile until they migrate to the
// *Response names; delete this file once downstream modules have been updated

// JobRunnerRegistrationRequest is the request to register a new node; the endpoint was
// removed on this branch and the type remains only for released consumers
//
// Deprecated: the job runner registration endpoint no longer exists
type JobRunnerRegistrationRequest struct {
	// IPAddress is the ip_address value.
	IPAddress string `json:"ip_address" description:"The IP address of the node being registered"`
	// Token is the token value.
	Token string `json:"token" description:"Your agent registration token"`
	// Name is the name value.
	Name string `json:"name" description:"the name of your job runner node"`
	// Tags is the tags value.
	Tags []string `json:"tags" description:"The tags for your runner node"`
}

// Validate ensures the required fields are set on the JobRunnerRegistrationRequest
//
// Deprecated: the job runner registration endpoint no longer exists
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

// JobRunnerRegistrationReply is the response to a job runner registration; the endpoint
// was removed on this branch and the type remains only for released consumers
//
// Deprecated: the job runner registration endpoint no longer exists
type JobRunnerRegistrationReply struct {
	// Reply is the reply value.
	Reply rout.Reply
	// Message is the message value.
	Message string `json:"message"`
}

// AccountAccessReply is an alias for AccountAccessResponse
//
// Deprecated: use AccountAccessResponse
type AccountAccessReply = AccountAccessResponse

// AccountFeaturesReply is an alias for AccountFeaturesResponse
//
// Deprecated: use AccountFeaturesResponse
type AccountFeaturesReply = AccountFeaturesResponse

// AccountRolesMeReply is an alias for AccountRolesMeResponse
//
// Deprecated: use AccountRolesMeResponse
type AccountRolesMeReply = AccountRolesMeResponse

// AccountRolesOrganizationReply is an alias for AccountRolesOrganizationResponse
//
// Deprecated: use AccountRolesOrganizationResponse
type AccountRolesOrganizationReply = AccountRolesOrganizationResponse

// AccountRolesReply is an alias for AccountRolesResponse
//
// Deprecated: use AccountRolesResponse
type AccountRolesReply = AccountRolesResponse

// AvailableAuthTypeReply is an alias for AvailableAuthTypeResponse
//
// Deprecated: use AvailableAuthTypeResponse
type AvailableAuthTypeReply = AvailableAuthTypeResponse

// EndImpersonationReply is an alias for EndImpersonationResponse
//
// Deprecated: use EndImpersonationResponse
type EndImpersonationReply = EndImpersonationResponse

// FileDownloadReply is an alias for FileDownloadResponse
//
// Deprecated: use FileDownloadResponse
type FileDownloadReply = FileDownloadResponse

// ForgotPasswordReply is an alias for ForgotPasswordResponse
//
// Deprecated: use ForgotPasswordResponse
type ForgotPasswordReply = ForgotPasswordResponse

// InviteReply is an alias for InviteResponse
//
// Deprecated: use InviteResponse
type InviteReply = InviteResponse

// LoginReply is an alias for LoginResponse
//
// Deprecated: use LoginResponse
type LoginReply = LoginResponse

// LogoutReply is an alias for LogoutResponse
//
// Deprecated: use LogoutResponse
type LogoutReply = LogoutResponse

// OrganizationRolesReply is an alias for OrganizationRolesResponse
//
// Deprecated: use OrganizationRolesResponse
type OrganizationRolesReply = OrganizationRolesResponse

// ProductCatalogReply is an alias for ProductCatalogResponse
//
// Deprecated: use ProductCatalogResponse
type ProductCatalogReply = ProductCatalogResponse

// RefreshReply is an alias for RefreshResponse
//
// Deprecated: use RefreshResponse
type RefreshReply = RefreshResponse

// RegisterReply is an alias for RegisterResponse
//
// Deprecated: use RegisterResponse
type RegisterReply = RegisterResponse

// ResendReply is an alias for ResendResponse
//
// Deprecated: use ResendResponse
type ResendReply = ResendResponse

// ResetPasswordReply is an alias for ResetPasswordResponse
//
// Deprecated: use ResetPasswordResponse
type ResetPasswordReply = ResetPasswordResponse

// RolesReply is an alias for RolesResponse
//
// Deprecated: use RolesResponse
type RolesReply = RolesResponse

// ScopesReply is an alias for ScopesResponse
//
// Deprecated: use ScopesResponse
type ScopesReply = ScopesResponse

// SnapshotReply is an alias for SnapshotResponse
//
// Deprecated: use SnapshotResponse
type SnapshotReply = SnapshotResponse

// SSOLoginReply is an alias for SSOLoginResponse
//
// Deprecated: use SSOLoginResponse
type SSOLoginReply = SSOLoginResponse

// SSOStatusReply is an alias for SSOStatusResponse
//
// Deprecated: use SSOStatusResponse
type SSOStatusReply = SSOStatusResponse

// SSOTokenAuthorizeReply is an alias for SSOTokenAuthorizeResponse
//
// Deprecated: use SSOTokenAuthorizeResponse
type SSOTokenAuthorizeReply = SSOTokenAuthorizeResponse

// StartImpersonationReply is an alias for StartImpersonationResponse
//
// Deprecated: use StartImpersonationResponse
type StartImpersonationReply = StartImpersonationResponse

// SupportAccessReply is an alias for SupportAccessResponse
//
// Deprecated: use SupportAccessResponse
type SupportAccessReply = SupportAccessResponse

// SwitchOrganizationReply is an alias for SwitchOrganizationResponse
//
// Deprecated: use SwitchOrganizationResponse
type SwitchOrganizationReply = SwitchOrganizationResponse

// TFAReply is an alias for TFAResponse
//
// Deprecated: use TFAResponse
type TFAReply = TFAResponse

// UnsubscribeReply is an alias for UnsubscribeResponse
//
// Deprecated: use UnsubscribeResponse
type UnsubscribeReply = UnsubscribeResponse

// UploadFilesReply is an alias for UploadFilesResponse
//
// Deprecated: use UploadFilesResponse
type UploadFilesReply = UploadFilesResponse

// UserInfoReply is an alias for UserInfoResponse
//
// Deprecated: use UserInfoResponse
type UserInfoReply = UserInfoResponse

// VerifyReply is an alias for VerifyResponse
//
// Deprecated: use VerifyResponse
type VerifyReply = VerifyResponse

// VerifySubscribeReply is an alias for VerifySubscribeResponse
//
// Deprecated: use VerifySubscribeResponse
type VerifySubscribeReply = VerifySubscribeResponse
