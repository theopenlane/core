package azureentraid

import "errors"

var (
	// ErrCredentialMetadataRequired indicates the credential provider data is missing
	ErrCredentialMetadataRequired = errors.New("azureentraid: credential metadata required")
	// ErrTokenAcquireFailed indicates the client credentials token request failed
	ErrTokenAcquireFailed = errors.New("azureentraid: failed to acquire access token")
	// ErrUsersFetchFailed indicates the Microsoft Graph users listing request failed
	ErrUsersFetchFailed = errors.New("azureentraid: users fetch failed")
	// ErrGroupsFetchFailed indicates the Microsoft Graph groups listing request failed
	ErrGroupsFetchFailed = errors.New("azureentraid: groups fetch failed")
	// ErrMembersFetchFailed indicates the Microsoft Graph group members request failed
	ErrMembersFetchFailed = errors.New("azureentraid: group members fetch failed")
	// ErrPayloadEncode indicates an ingest envelope payload could not be serialized
	ErrPayloadEncode = errors.New("azureentraid: payload encode failed")
)
