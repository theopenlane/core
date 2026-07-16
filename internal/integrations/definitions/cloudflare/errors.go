package cloudflare

import "errors"

var (
	// ErrAPITokenMissing indicates the Cloudflare API token is missing from the credential
	ErrAPITokenMissing = errors.New("cloudflare: api token missing")
	// ErrCredentialInvalid indicates credential metadata could not be decoded
	ErrCredentialInvalid = errors.New("cloudflare: credential invalid")
	// ErrCredentialMetadataRequired indicates no credential metadata was provided
	ErrCredentialMetadataRequired = errors.New("cloudflare: credential metadata required")
	// ErrTokenVerificationFailed indicates the Cloudflare token verification failed
	ErrTokenVerificationFailed = errors.New("cloudflare: token verification failed")
	// ErrTokenNotActive indicates the Cloudflare token is not in an active state
	ErrTokenNotActive = errors.New("cloudflare: token is not active")
	// ErrClientType indicates the provided client is not a Cloudflare client
	ErrClientType = errors.New("cloudflare: unexpected client type")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("cloudflare: result encode failed")
	// ErrMembersFetchFailed indicates the account members list request failed
	ErrMembersFetchFailed = errors.New("cloudflare: members fetch failed")
	// ErrMembersFetchFailed indicates the account members list request failed
	ErrGroupsFetchFailed = errors.New("cloudflare: groups fetch failed")
	// ErrPayloadEncode indicates a collected Cloudflare payload could not be serialized for ingest
	ErrPayloadEncode = errors.New("cloudflare: ingest payload encode failed")
	// ErrAccountIDMissing indicates the account ID is missing from user input
	ErrAccountIDMissing = errors.New("cloudflare: account id missing")
	// ErrOperationConfigInvalid indicates operation configuration could not be decoded
	ErrOperationConfigInvalid = errors.New("cloudflare: operation config invalid")
	// ErrFindingsFetchFailed indicates the Security Center insights list request failed
	ErrFindingsFetchFailed = errors.New("cloudflare: findings fetch failed")
	// ErrAssetsFetchFailed indicates the Registrar registrations list request failed
	ErrAssetsFetchFailed = errors.New("cloudflare: assets fetch failed")
	// ErrRuntimeConfigDecode indicates the runtime Cloudflare config could not be decoded
	ErrRuntimeConfigDecode = errors.New("cloudflare: runtime config decode failed")
	// ErrRuntimeConfigInvalid indicates the runtime Cloudflare config is missing required fields
	ErrRuntimeConfigInvalid = errors.New("cloudflare: runtime config invalid")
	// ErrDomainScanSubmitFailed indicates the URL Scanner bulk submission request failed
	ErrDomainScanSubmitFailed = errors.New("cloudflare: domain scan submit failed")
	// ErrDomainScanResultFailed indicates the URL Scanner result retrieval request failed
	ErrDomainScanResultFailed = errors.New("cloudflare: domain scan result fetch failed")
	// ErrDomainScanTaskFailed indicates Cloudflare reported the scan task itself failed
	ErrDomainScanTaskFailed = errors.New("cloudflare: domain scan task failed")
)
