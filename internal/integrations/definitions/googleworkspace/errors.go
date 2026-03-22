package googleworkspace

import "errors"

var (
	// ErrOAuthTokenMissing indicates the OAuth access token is missing
	ErrOAuthTokenMissing = errors.New("googleworkspace: oauth token missing")
	// ErrClientType indicates the provided client is not the expected type
	ErrClientType = errors.New("googleworkspace: unexpected client type")
	// ErrAdminServiceBuildFailed indicates the Admin SDK client could not be constructed
	ErrAdminServiceBuildFailed = errors.New("googleworkspace: admin service build failed")
	// ErrHealthCheckFailed indicates the health check request failed
	ErrHealthCheckFailed = errors.New("googleworkspace: health check failed")
	// ErrDirectoryUsersFetchFailed indicates the users listing failed
	ErrDirectoryUsersFetchFailed = errors.New("googleworkspace: directory users fetch failed")
	// ErrDirectoryGroupsFetchFailed indicates the groups listing failed
	ErrDirectoryGroupsFetchFailed = errors.New("googleworkspace: directory groups fetch failed")
	// ErrDirectoryGroupMembersFetchFailed indicates the group members listing failed
	ErrDirectoryGroupMembersFetchFailed = errors.New("googleworkspace: directory group members fetch failed")
	// ErrPayloadEncode indicates a provider payload could not be serialized
	ErrPayloadEncode = errors.New("googleworkspace: payload encode failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("googleworkspace: result encode failed")
	// ErrCredentialEncode indicates the credential could not be serialized
	ErrCredentialEncode = errors.New("googleworkspace: credential encode failed")
	// ErrCredentialDecode indicates the credential could not be deserialized
	ErrCredentialDecode = errors.New("googleworkspace: credential decode failed")
)
