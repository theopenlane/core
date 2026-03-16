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
	// ErrCredentialInvalid indicates credential metadata could not be decoded
	ErrCredentialInvalid = errors.New("azuresecuritycenter: credential invalid")
	// ErrTokenExchangeFailed indicates the Azure token exchange failed
	ErrTokenExchangeFailed = errors.New("azuresecuritycenter: token exchange failed")
	// ErrPricingsClientBuildFailed indicates the pricings client could not be constructed
	ErrPricingsClientBuildFailed = errors.New("azuresecuritycenter: pricings client build failed")
	// ErrPricingFetchFailed indicates the pricing list request failed
	ErrPricingFetchFailed = errors.New("azuresecuritycenter: pricing fetch failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("azuresecuritycenter: result encode failed")
)
