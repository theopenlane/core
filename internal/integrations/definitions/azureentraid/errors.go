package azureentraid

import "errors"

var (
	// ErrCredentialMetadataRequired indicates the credential provider data is missing
	ErrCredentialMetadataRequired = errors.New("azureentraid: credential metadata required")
	// ErrTokenAcquireFailed indicates the client credentials token request failed
	ErrTokenAcquireFailed = errors.New("azureentraid: failed to acquire access token")
	// ErrUsersFetchFailed indicates the Microsoft Graph users listing request failed
	ErrUsersFetchFailed = errors.New("azureentraid: users fetch failed")
	// ErrGroupsFetchFailed indicates the Microsoft Graph groups listing request failed
	ErrGroupsFetchFailed = errors.New("azureentraid: groups fetch failed")
	// ErrMembersFetchFailed indicates the Microsoft Graph group members request failed
	ErrMembersFetchFailed = errors.New("azureentraid: group members fetch failed")
	// ErrPayloadEncode indicates an ingest envelope payload could not be serialized
	ErrPayloadEncode = errors.New("azureentraid: payload encode failed")
	// ErrTenantIDNotFound indicates the tenant ID claim was not found in OAuth material
	ErrTenantIDNotFound = errors.New("azureentraid: tenant id not found in claims")
	// ErrCredentialEncode indicates the credential could not be serialized
	ErrCredentialEncode = errors.New("azureentraid: credential encode failed")
	// ErrConsentStateGeneration indicates the CSRF state could not be generated
	ErrConsentStateGeneration = errors.New("azureentraid: admin consent state generation failed")
	// ErrConsentStateInvalid indicates the stored admin consent start state could not be decoded
	ErrConsentStateInvalid = errors.New("azureentraid: admin consent state invalid")
	// ErrConsentStateMismatch indicates the callback state does not match the stored CSRF state
	ErrConsentStateMismatch = errors.New("azureentraid: admin consent state mismatch")
	// ErrConsentDenied indicates the admin denied or cancelled the consent request
	ErrConsentDenied = errors.New("azureentraid: admin consent denied")
)
