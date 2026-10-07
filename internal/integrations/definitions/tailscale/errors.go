package tailscale

import "errors"

var (
	// ErrClientIDMissing indicates the OAuth client ID is missing from the credential
	ErrClientIDMissing = errors.New("tailscale: oauth client id missing")
	// ErrClientSecretMissing indicates the OAuth client secret is missing from the credential
	ErrClientSecretMissing = errors.New("tailscale: oauth client secret missing")
	// ErrUsersFetchFailed indicates the Tailscale users list request failed
	ErrUsersFetchFailed = errors.New("tailscale: users fetch failed")
	// ErrTailnetUnresolved indicates no member user carried a tailnet name to identify the tailnet by
	ErrTailnetUnresolved = errors.New("tailscale: tailnet unresolved")
	// ErrDevicesFetchFailed indicates the Tailscale devices list request failed
	ErrDevicesFetchFailed = errors.New("tailscale: devices fetch failed")
	// ErrPayloadEncode indicates a collected Tailscale payload could not be serialized for ingest
	ErrPayloadEncode = errors.New("tailscale: ingest payload encode failed")
)
