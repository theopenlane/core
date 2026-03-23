package okta

import "errors"

var (
	// ErrAPITokenMissing indicates the Okta API token is missing from the credential
	ErrAPITokenMissing = errors.New("okta: api token missing")
	// ErrOrgURLMissing indicates the Okta org URL is missing from the credential
	ErrOrgURLMissing = errors.New("okta: org url missing")
	// ErrClientType indicates the provided client is not an Okta API client
	ErrClientType = errors.New("okta: unexpected client type")
	// ErrCredentialInvalid indicates credential metadata could not be decoded
	ErrCredentialInvalid = errors.New("okta: credential invalid")
	// ErrClientConfigInvalid indicates the Okta client configuration is invalid
	ErrClientConfigInvalid = errors.New("okta: client config invalid")
	// ErrUserLookupFailed indicates the current user lookup failed
	ErrUserLookupFailed = errors.New("okta: user lookup failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("okta: result encode failed")
	// ErrDirectoryUsersFetchFailed indicates the Okta users listing failed
	ErrDirectoryUsersFetchFailed = errors.New("okta: directory users fetch failed")
	// ErrDirectoryGroupsFetchFailed indicates the Okta groups listing failed
	ErrDirectoryGroupsFetchFailed = errors.New("okta: directory groups fetch failed")
	// ErrDirectoryGroupMembersFetchFailed indicates group member listing failed
	ErrDirectoryGroupMembersFetchFailed = errors.New("okta: directory group members fetch failed")
	// ErrPayloadEncode indicates a directory payload could not be serialized
	ErrPayloadEncode = errors.New("okta: payload encode failed")
)
