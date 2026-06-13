package tailscale

import "errors"

var (
	// ErrClientIDMissing indicates the OAuth client ID is missing from the credential
	ErrClientIDMissing = errors.New("tailscale: oauth client id missing")
	// ErrClientSecretMissing indicates the OAuth client secret is missing from the credential
	ErrClientSecretMissing = errors.New("tailscale: oauth client secret missing")
	// ErrCredentialInvalid indicates credential metadata could not be decoded
	ErrCredentialInvalid = errors.New("tailscale: credential invalid")
	// ErrCredentialMetadataRequired indicates no credential metadata was provided
	ErrCredentialMetadataRequired = errors.New("tailscale: credential metadata required")
	// ErrHealthCheckFailed indicates the Tailscale API health check failed
	ErrHealthCheckFailed = errors.New("tailscale: health check failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("tailscale: result encode failed")
	// ErrUsersFetchFailed indicates the Tailscale users list request failed
	ErrUsersFetchFailed = errors.New("tailscale: users fetch failed")
	// ErrDevicesFetchFailed indicates the Tailscale devices list request failed
	ErrDevicesFetchFailed = errors.New("tailscale: devices fetch failed")
	// ErrPayloadEncode indicates a collected Tailscale payload could not be serialized for ingest
	ErrPayloadEncode = errors.New("tailscale: ingest payload encode failed")
	// ErrOperationConfigInvalid indicates operation config could not be decoded
	ErrOperationConfigInvalid = errors.New("tailscale: operation config invalid")
)
