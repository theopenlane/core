package microsoft

import "errors"

var (
	// ErrTenantIDRequired indicates a token credential was requested without a tenant ID
	ErrTenantIDRequired = errors.New("microsoft: tenant id required")
	// ErrTokenCredentialBuild indicates the client secret token credential could not be constructed
	ErrTokenCredentialBuild = errors.New("microsoft: token credential build failed")
	// ErrSubscriptionListFailed indicates the ARM subscriptions listing request failed
	ErrSubscriptionListFailed = errors.New("microsoft: subscription list failed")
	// ErrResourceGraphQueryFailed indicates an Azure Resource Graph query failed
	ErrResourceGraphQueryFailed = errors.New("microsoft: resource graph query failed")
	// ErrTenantIDNotFound indicates the tenant ID was not returned on the admin consent callback
	ErrTenantIDNotFound = errors.New("microsoft: tenant id not found in admin consent callback")
	// ErrCredentialEncode indicates the credential could not be serialized
	ErrCredentialEncode = errors.New("microsoft: credential encode failed")
	// ErrConsentStateGeneration indicates the CSRF state could not be generated for the admin consent flow
	ErrConsentStateGeneration = errors.New("microsoft: admin consent state generation failed")
	// ErrConsentStateInvalid indicates the stored admin consent start state could not be decoded
	ErrConsentStateInvalid = errors.New("microsoft: admin consent state invalid")
	// ErrConsentStateMismatch indicates the callback state does not match the stored CSRF state
	ErrConsentStateMismatch = errors.New("microsoft: admin consent state mismatch")
	// ErrConsentDenied indicates the admin denied or cancelled the consent request
	ErrConsentDenied = errors.New("microsoft: admin consent denied")
)
