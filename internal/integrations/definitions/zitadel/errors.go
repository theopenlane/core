package zitadel

import "errors"

var (
	// ErrTokenMissing indicates the personal access token is missing from the credential
	ErrTokenMissing = errors.New("zitadel: token missing")
	// ErrDomainMissing indicates the Zitadel domain is missing from the credential
	ErrDomainMissing = errors.New("zitadel: domain missing")
	// ErrCredentialEncode indicates the credential could not be serialized
	ErrCredentialEncode = errors.New("zitadel: credential encode failed")
	// ErrCredentialDecode indicates the credential could not be deserialized
	ErrCredentialDecode = errors.New("zitadel: credential decode failed")
	// ErrClientBuildFailed indicates the Zitadel API client could not be constructed
	ErrClientBuildFailed = errors.New("zitadel: client build failed")
	// ErrHealthCheckFailed indicates the health check request failed
	ErrHealthCheckFailed = errors.New("zitadel: health check failed")
	// ErrDirectoryUsersFetchFailed indicates the users listing failed
	ErrDirectoryUsersFetchFailed = errors.New("zitadel: directory users fetch failed")
	// ErrPayloadEncode indicates a provider payload could not be serialized
	ErrPayloadEncode = errors.New("zitadel: payload encode failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("zitadel: result encode failed")
)