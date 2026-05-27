package keycloak

import "errors"

var (
	// ErrClientIDMissing indicates the client ID is missing from the credential
	ErrClientIDMissing = errors.New("keycloak: client id missing")
	// ErrClientSecretMissing indicates the client secret is missing from the credential
	ErrClientSecretMissing = errors.New("keycloak: client secret missing")
	// ErrBaseURLMissing indicates the Keycloak base URL is missing from the credential
	ErrBaseURLMissing = errors.New("keycloak: base url missing")
	// ErrRealmMissing indicates the Keycloak realm is missing from the credential
	ErrRealmMissing = errors.New("keycloak: realm missing")
	// ErrClientType indicates the provided client is not the expected keycloak client type
	ErrClientType = errors.New("keycloak: unexpected client type")
	// ErrCredentialEncode indicates the credential could not be serialized
	ErrCredentialEncode = errors.New("keycloak: credential encode failed")
	// ErrCredentialDecode indicates the credential could not be deserialized
	ErrCredentialDecode = errors.New("keycloak: credential decode failed")
	// ErrTokenAcquireFailed indicates the token could not be acquired from Keycloak
	ErrTokenAcquireFailed = errors.New("keycloak: token acquire failed")
	// ErrHealthCheckFailed indicates the health check request failed
	ErrHealthCheckFailed = errors.New("keycloak: health check failed")
	// ErrInstallationResolveFailed indicates the realm info could not be resolved
	ErrInstallationResolveFailed = errors.New("keycloak: installation resolve failed")
	// ErrDirectoryUsersFetchFailed indicates the users listing failed
	ErrDirectoryUsersFetchFailed = errors.New("keycloak: directory users fetch failed")
	// ErrDirectoryGroupsFetchFailed indicates the groups listing failed
	ErrDirectoryGroupsFetchFailed = errors.New("keycloak: directory groups fetch failed")
	// ErrDirectoryGroupMembersFetchFailed indicates the group members listing failed
	ErrDirectoryGroupMembersFetchFailed = errors.New("keycloak: directory group members fetch failed")
	// ErrPayloadEncode indicates a provider payload could not be serialized
	ErrPayloadEncode = errors.New("keycloak: payload encode failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("keycloak: result encode failed")
)