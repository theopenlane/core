package keystore

import "errors"

var (
	// ErrCredentialNotFound indicates no credential exists for the supplied installation
	ErrCredentialNotFound = errors.New("keystore: credential not found")
	// ErrStoreNotInitialized indicates the store instance is nil
	ErrStoreNotInitialized = errors.New("keystore: store not initialized")
	// ErrInstallationIDRequired indicates the caller did not provide an installation identifier
	ErrInstallationIDRequired = errors.New("keystore: installation id required")
)
