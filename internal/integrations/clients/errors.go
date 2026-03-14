package clients

import "errors"

var (
	// ErrRegistryRequired indicates the definition registry dependency is missing
	ErrRegistryRequired = errors.New("integrationsv2/clients: registry required")
	// ErrCredentialResolverRequired indicates the credential resolver dependency is missing
	ErrCredentialResolverRequired = errors.New("integrationsv2/clients: credential resolver required")
	// ErrInstallationRequired indicates the installation record dependency is missing
	ErrInstallationRequired = errors.New("integrationsv2/clients: installation required")
	// ErrDefinitionIDRequired indicates the installation is missing its definition id
	ErrDefinitionIDRequired = errors.New("integrationsv2/clients: definition id required")
	// ErrCredentialNotFound indicates no credential could be resolved for the installation
	ErrCredentialNotFound = errors.New("integrationsv2/clients: credential not found")
)
