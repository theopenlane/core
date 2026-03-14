package keystore

import "errors"

var (
	// ErrCredentialNotFound indicates no credential exists for the supplied installation
	ErrCredentialNotFound = errors.New("keystore: credential not found")
	// ErrStoreNotInitialized indicates the store instance is nil
	ErrStoreNotInitialized = errors.New("keystore: store not initialized")
	// ErrBrokerNotInitialized indicates the broker was unable to be initialized due to missing dependencies
	ErrBrokerNotInitialized = errors.New("keystore: broker not initialized")
	// ErrInstallationIDRequired indicates the caller did not provide an installation identifier
	ErrInstallationIDRequired = errors.New("keystore: installation id required")
	// ErrDefinitionNotFound indicates the installation's definition could not be resolved from the registry
	ErrDefinitionNotFound = errors.New("keystore: definition not found")
	// ErrRefreshNotSupported indicates the definition has no auth refresh function registered
	ErrRefreshNotSupported = errors.New("keystore: credential refresh not supported for this definition")
)
