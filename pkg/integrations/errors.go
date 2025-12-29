package integrations

import (
	"errors"
)

var (

	// ErrKeystoreRequired indicates the keystore dependency is missing.
	ErrKeystoreRequired = errors.New("integrations: keystore required")
	// ErrSessionStoreRequired indicates the session store dependency is missing.
	ErrSessionStoreRequired = errors.New("integrations: session store required")
	// ErrProviderNotFound signals the requested provider does not exist in the registry.
	ErrProviderNotFound = errors.New("integrations: provider not found")
	// ErrProviderConfigNotFound signals the requested provider config metadata is unavailable.
	ErrProviderConfigNotFound = errors.New("integrations: provider config not found")
	// ErrProviderRegistryUninitialized indicates provider operations were attempted before the registry was built.
	ErrProviderRegistryUninitialized = errors.New("integrations: provider registry uninitialized")
	// ErrOrgIDRequired indicates the org identifier was omitted.
	ErrOrgIDRequired = errors.New("integrations: org id required")
	// ErrIntegrationIDRequired indicates the integration identifier was omitted.
	ErrIntegrationIDRequired = errors.New("integrations: integration id required")
	// ErrCredentialNotFound is returned when no credential is stored for the requested integration.
	ErrCredentialNotFound = errors.New("integrations: credential not found")
	// ErrCredentialExpired indicates a stored credential is no longer valid.
	ErrCredentialExpired = errors.New("integrations: credential expired")
	// ErrStateRequired indicates the OAuth state parameter is missing.
	ErrStateRequired = errors.New("integrations: state required")
	// ErrAuthorizationCodeRequired indicates the OAuth authorization code was omitted.
	ErrAuthorizationCodeRequired = errors.New("integrations: authorization code required")
	// ErrAuthorizationStateNotFound indicates the provided state does not map to an active session.
	ErrAuthorizationStateNotFound = errors.New("integrations: authorization state not found")
	// ErrAuthorizationStateExpired indicates the stored session has expired.
	ErrAuthorizationStateExpired = errors.New("integrations: authorization state expired")
	// ErrAuthSessionInvalid indicates the stored auth session reference is invalid.
	ErrAuthSessionInvalid = errors.New("integrations: auth session invalid")
)
