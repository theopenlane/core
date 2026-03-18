package azuresecuritycenter

import "errors"

var (
	// ErrAuthTypeMismatch indicates the provider spec specifies an incompatible auth type.
	ErrAuthTypeMismatch = errors.New("azuresecuritycenter: auth type mismatch")
	// ErrBeginAuthNotSupported indicates BeginAuth is not supported for this provider.
	ErrBeginAuthNotSupported = errors.New("azuresecuritycenter: BeginAuth is not supported; configure credentials via metadata")
	// ErrProviderMetadataRequired indicates required provider metadata is missing.
	ErrProviderMetadataRequired = errors.New("azuresecuritycenter: provider metadata required")
	// ErrTenantIDMissing indicates the tenant ID is missing.
	ErrTenantIDMissing = errors.New("azuresecuritycenter: tenant ID missing")
	// ErrClientIDMissing indicates the client ID is missing.
	ErrClientIDMissing = errors.New("azuresecuritycenter: client ID missing")
	// ErrClientSecretMissing indicates the client secret is missing.
	ErrClientSecretMissing = errors.New("azuresecuritycenter: client secret missing")
	// ErrSubscriptionIDMissing indicates the subscription ID is missing.
	ErrSubscriptionIDMissing = errors.New("azuresecuritycenter: subscription ID missing")
	// ErrTokenExchangeFailed indicates the client credentials token exchange failed.
	ErrTokenExchangeFailed = errors.New("azuresecuritycenter: token exchange failed")
)
