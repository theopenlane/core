package azureentraid

import "errors"

var (
	// ErrOAuthTokenMissing indicates the OAuth access token is missing
	ErrOAuthTokenMissing = errors.New("azureentraid: oauth token missing")
	// ErrClientType indicates the provided client is not the expected type
	ErrClientType = errors.New("azureentraid: unexpected client type")
	// ErrNoOrganizations indicates the Microsoft Graph organization endpoint returned no results
	ErrNoOrganizations = errors.New("azureentraid: no organizations returned")
	// ErrOrganizationLookupFailed indicates the Microsoft Graph organization lookup failed
	ErrOrganizationLookupFailed = errors.New("azureentraid: organization lookup failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("azureentraid: result encode failed")
)
