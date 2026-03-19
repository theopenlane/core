package azureentraid

import "errors"

var (
	// ErrCredentialMetadataRequired indicates the credential provider data is missing
	ErrCredentialMetadataRequired = errors.New("azureentraid: credential metadata required")
	// ErrMetadataDecode indicates the credential metadata could not be decoded
	ErrMetadataDecode = errors.New("azureentraid: credential metadata decode failed")
	// ErrTokenAcquireFailed indicates the client credentials token request failed
	ErrTokenAcquireFailed = errors.New("azureentraid: failed to acquire access token")
	// ErrClientType indicates the provided client is not the expected type
	ErrClientType = errors.New("azureentraid: unexpected client type")
	// ErrNoOrganizations indicates the Microsoft Graph organization endpoint returned no results
	ErrNoOrganizations = errors.New("azureentraid: no organizations returned")
	// ErrOrganizationLookupFailed indicates the Microsoft Graph organization lookup failed
	ErrOrganizationLookupFailed = errors.New("azureentraid: organization lookup failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("azureentraid: result encode failed")
)
