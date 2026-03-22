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
	// ErrUsersFetchFailed indicates the Microsoft Graph users listing request failed
	ErrUsersFetchFailed = errors.New("azureentraid: users fetch failed")
	// ErrGroupsFetchFailed indicates the Microsoft Graph groups listing request failed
	ErrGroupsFetchFailed = errors.New("azureentraid: groups fetch failed")
	// ErrMembersFetchFailed indicates the Microsoft Graph group members request failed
	ErrMembersFetchFailed = errors.New("azureentraid: group members fetch failed")
	// ErrPayloadEncode indicates an ingest envelope payload could not be serialized
	ErrPayloadEncode = errors.New("azureentraid: payload encode failed")
	// ErrTenantIDNotFound indicates the tenant ID claim was not found in OAuth material
	ErrTenantIDNotFound = errors.New("azureentraid: tenant id not found in claims")
	// ErrCredentialEncode indicates the credential could not be serialized
	ErrCredentialEncode = errors.New("azureentraid: credential encode failed")
	// ErrCredentialDecode indicates the credential could not be deserialized
	ErrCredentialDecode = errors.New("azureentraid: credential decode failed")
)
