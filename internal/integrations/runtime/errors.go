package runtime

import "errors"

var (
	// ErrOwnerIDRequired indicates installation resolution requires an owner id when resolving by definition
	ErrOwnerIDRequired = errors.New("integrations/runtime: owner id required")
	// ErrDefinitionIDRequired indicates installation resolution requires a definition id when no explicit installation id is given
	ErrDefinitionIDRequired = errors.New("integrations/runtime: definition id required")
	// ErrInstallationRequired indicates the installation record dependency is missing
	ErrInstallationRequired = errors.New("integrations/runtime: installation required")
	// ErrInstallationIDRequired indicates installation resolution requires an installation id when owner plus definition is ambiguous
	ErrInstallationIDRequired = errors.New("integrations/runtime: installation id required")
	// ErrInstallationNotFound indicates no matching installation could be resolved
	ErrInstallationNotFound = errors.New("integrations/runtime: installation not found")
	// ErrInstallationDefinitionMismatch indicates the resolved installation does not match the requested definition
	ErrInstallationDefinitionMismatch = errors.New("integrations/runtime: installation definition mismatch")
	// ErrCredentialNotFound indicates no credential could be resolved for the installation
	ErrCredentialNotFound = errors.New("integrations/runtime: credential not found")
	// ErrDefinitionNotFound indicates the requested integration definition is not registered
	ErrDefinitionNotFound = errors.New("integrations/runtime: definition not found")
	// ErrOperationNotFound indicates the requested operation is not registered for the definition
	ErrOperationNotFound = errors.New("integrations/runtime: operation not found")
)
