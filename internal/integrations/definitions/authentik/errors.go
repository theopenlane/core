package authentik

import "errors"

var (
	// ErrAPITokenMissing indicates the API token is missing from the credential
	ErrAPITokenMissing = errors.New("authentik: api token missing")
	// ErrBaseURLMissing indicates the Authentik base URL is missing from the credential
	ErrBaseURLMissing = errors.New("authentik: base url missing")
	// ErrClientType indicates the provided client is not the expected authentik client type
	ErrClientType = errors.New("authentik: unexpected client type")
	// ErrClientConfigInvalid indicates the Authentik client configuration is invalid
	ErrClientConfigInvalid = errors.New("authentik: client config invalid")
	// ErrCredentialEncode indicates the credential could not be serialized
	ErrCredentialEncode = errors.New("authentik: credential encode failed")
	// ErrCredentialDecode indicates the credential could not be deserialized
	ErrCredentialDecode = errors.New("authentik: credential decode failed")
	// ErrHealthCheckFailed indicates the health check request failed
	ErrHealthCheckFailed = errors.New("authentik: health check failed")
	// ErrDirectoryUsersFetchFailed indicates the users listing failed
	ErrDirectoryUsersFetchFailed = errors.New("authentik: directory users fetch failed")
	// ErrDirectoryGroupsFetchFailed indicates the groups listing failed
	ErrDirectoryGroupsFetchFailed = errors.New("authentik: directory groups fetch failed")
	// ErrDirectoryGroupMembersFetchFailed indicates the group members listing failed
	ErrDirectoryGroupMembersFetchFailed = errors.New("authentik: directory group members fetch failed")
	// ErrPayloadEncode indicates a provider payload could not be serialized
	ErrPayloadEncode = errors.New("authentik: payload encode failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("authentik: result encode failed")
)
