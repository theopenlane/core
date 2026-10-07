package googleworkspace

import "errors"

var (
	// ErrOAuthTokenMissing indicates the OAuth access token is missing
	ErrOAuthTokenMissing = errors.New("googleworkspace: oauth token missing")
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
	// ErrCredentialEncode indicates the credential could not be serialized
	ErrCredentialEncode = errors.New("googleworkspace: credential encode failed")
	// ErrCustomerFetchFailed indicates the customer lookup request failed
	ErrCustomerFetchFailed = errors.New("googleworkspace: customer fetch failed")
	// ErrCustomerUnresolved indicates the customer lookup returned neither an id nor a domain
	ErrCustomerUnresolved = errors.New("googleworkspace: customer unresolved")
	// ErrInstallationMetadataInvalid indicates stored installation metadata could not be decoded
	ErrInstallationMetadataInvalid = errors.New("googleworkspace: installation metadata invalid")
	// ErrCustomerIDMissing indicates installation metadata is missing the required customer identifier
	ErrCustomerIDMissing = errors.New("googleworkspace: customer id missing from installation metadata")
)
