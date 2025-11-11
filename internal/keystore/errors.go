package keystore

import "errors"

var (
	// ErrProviderRequired indicates a credential operation lacked a provider identifier
	ErrProviderRequired = errors.New("keystore: provider required")
	// ErrOrgIDRequired indicates the caller did not provide an organization identifier
	ErrOrgIDRequired = errors.New("keystore: org id required")
	// ErrCredentialNotFound indicates no credential exists for the supplied org/provider
	ErrCredentialNotFound = errors.New("keystore: credential not found")
	// ErrProviderNotRegistered indicates the registry does not have a provider implementation for the requested type
	ErrProviderNotRegistered = errors.New("keystore: provider not registered")
)
