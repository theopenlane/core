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
	// ErrBrokerRequired indicates a client pool was constructed without a credential broker/source
	ErrBrokerRequired = errors.New("keystore: credential broker required")
	// ErrClientBuilderRequired indicates a client pool was constructed without a builder
	ErrClientBuilderRequired = errors.New("keystore: client builder required")
	// ErrClientUnavailable indicates the requested client could not be created
	ErrClientUnavailable = errors.New("keystore: client unavailable")
	// ErrClientNotRegistered indicates no client descriptor/pool exists for the requested provider+client pair
	ErrClientNotRegistered = errors.New("keystore: client not registered")
	// ErrClientDescriptorInvalid indicates a provider published an invalid client descriptor
	ErrClientDescriptorInvalid = errors.New("keystore: client descriptor invalid")
	// ErrOperationNameRequired indicates the caller omitted an operation identifier
	ErrOperationNameRequired = errors.New("keystore: operation name required")
	// ErrOperationNotRegistered indicates no operation exists for the requested provider/name pair
	ErrOperationNotRegistered = errors.New("keystore: operation not registered")
	// ErrOperationDescriptorInvalid indicates a provider published an invalid operation descriptor
	ErrOperationDescriptorInvalid = errors.New("keystore: operation descriptor invalid")
	// ErrOperationClientManagerRequired indicates an operation requires a client pool but none was provided
	ErrOperationClientManagerRequired = errors.New("keystore: client manager required for operation")
	// ErrStoreNotInitialized indicates the store instance is nil
	ErrStoreNotInitialized = errors.New("keystore: store not initialized")
)
