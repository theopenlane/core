package keymaker

import "errors"

var (
	// ErrDefinitionIDRequired indicates the caller did not provide a definition identifier
	ErrDefinitionIDRequired = errors.New("keymaker: definition id required")
	// ErrDefinitionNotFound signals the requested definition does not exist in the registry
	ErrDefinitionNotFound = errors.New("keymaker: definition not found")
	// ErrDefinitionAuthRequired indicates the definition does not have an auth registration
	ErrDefinitionAuthRequired = errors.New("keymaker: definition has no auth registration")
	// ErrInstallationIDRequired indicates the caller did not provide an installation identifier
	ErrInstallationIDRequired = errors.New("keymaker: installation id required")
	// ErrDefinitionResolverRequired indicates the definition resolver dependency is missing
	ErrDefinitionResolverRequired = errors.New("keymaker: definition resolver required")
	// ErrCredentialWriterRequired indicates the credential writer dependency is missing
	ErrCredentialWriterRequired = errors.New("keymaker: credential writer required")
	// ErrAuthStateStoreRequired indicates the auth state store dependency is missing
	ErrAuthStateStoreRequired = errors.New("keymaker: definition auth state store required")
	// ErrAuthStateNotFound indicates the provided state token does not map to an active session
	ErrAuthStateNotFound = errors.New("keymaker: definition auth state not found")
	// ErrAuthStateExpired indicates the stored session has expired
	ErrAuthStateExpired = errors.New("keymaker: definition auth state expired")
	// ErrAuthStateStoreFull indicates the auth state store has reached capacity
	ErrAuthStateStoreFull = errors.New("keymaker: definition auth state store full")
	// ErrAuthStateTokenRequired indicates the state token is required for session lookup
	ErrAuthStateTokenRequired = errors.New("keymaker: state token required")
)
