package azuresecuritycenter

import "errors"

var (
	// ErrTenantIDMissing indicates the Azure tenant ID is missing
	ErrTenantIDMissing = errors.New("azuresecuritycenter: tenant ID required")
	// ErrClientIDMissing indicates the Azure client ID is missing
	ErrClientIDMissing = errors.New("azuresecuritycenter: client ID required")
	// ErrClientSecretMissing indicates the Azure client secret is missing
	ErrClientSecretMissing = errors.New("azuresecuritycenter: client secret required")
	// ErrSubscriptionIDMissing indicates the Azure subscription ID is missing
	ErrSubscriptionIDMissing = errors.New("azuresecuritycenter: subscription ID required")
	// ErrClientType indicates the provided client is not the expected type
	ErrClientType = errors.New("azuresecuritycenter: unexpected client type")
	// ErrTokenExchangeFailed indicates the Azure token exchange failed
	ErrTokenExchangeFailed = errors.New("azuresecuritycenter: token exchange failed")
)
