package operations

import "errors"

var (
	// ErrRegistryRequired indicates the definition registry dependency is missing
	ErrRegistryRequired = errors.New("integrationsv2/operations: registry required")
	// ErrInstallationStoreRequired indicates the installation store dependency is missing
	ErrInstallationStoreRequired = errors.New("integrationsv2/operations: installation store required")
	// ErrCredentialResolverRequired indicates the credential resolver dependency is missing
	ErrCredentialResolverRequired = errors.New("integrationsv2/operations: credential resolver required")
	// ErrClientServiceRequired indicates the client service dependency is missing
	ErrClientServiceRequired = errors.New("integrationsv2/operations: client service required")
	// ErrRunStoreRequired indicates the run store dependency is missing
	ErrRunStoreRequired = errors.New("integrationsv2/operations: run store required")
	// ErrGalaRequired indicates the gala dependency is missing
	ErrGalaRequired = errors.New("integrationsv2/operations: gala required")
	// ErrInstallationIDRequired indicates the installation identifier is missing
	ErrInstallationIDRequired = errors.New("integrationsv2/operations: installation id required")
	// ErrOperationNameRequired indicates the operation identifier is missing
	ErrOperationNameRequired = errors.New("integrationsv2/operations: operation name required")
	// ErrRunIDRequired indicates the run identifier is missing
	ErrRunIDRequired = errors.New("integrationsv2/operations: run id required")
)
