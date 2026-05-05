package runtime

import "errors"

var (
	// ErrIntegrationIDRequired indicates resolution requires an explicit integration ID
	ErrIntegrationIDRequired = errors.New("integrations/runtime: integration id required")
	// ErrInstallationRequired indicates the installation record dependency is missing
	ErrInstallationRequired = errors.New("integrations/runtime: installation required")
	// ErrInstallationNotFound indicates no matching installation could be resolved
	ErrInstallationNotFound = errors.New("integrations/runtime: installation not found")
	// ErrInstallationDefinitionMismatch indicates the resolved installation does not match the requested definition
	ErrInstallationDefinitionMismatch = errors.New("integrations/runtime: installation definition mismatch")
	// ErrConnectionRequired indicates the installation operation requires a credential-selected connection
	ErrConnectionRequired = errors.New("integrations/runtime: connection required")
	// ErrConnectionNotFound indicates the requested connection could not be resolved for the definition
	ErrConnectionNotFound = errors.New("integrations/runtime: connection not found")
	// ErrDefinitionNotFound indicates the requested integration definition is not registered
	ErrDefinitionNotFound = errors.New("integrations/runtime: definition not found")
	// ErrOperationNotFound indicates the requested operation is not registered for the definition
	ErrOperationNotFound = errors.New("integrations/runtime: operation not found")
	// ErrOperationConfigInvalid indicates the operation config payload failed schema validation
	ErrOperationConfigInvalid = errors.New("integrations/runtime: operation config invalid")
	// ErrUserInputInvalid indicates the user input payload failed schema validation
	ErrUserInputInvalid = errors.New("integrations/runtime: user input invalid")
	// ErrCredentialInvalid indicates the credential payload failed schema validation
	ErrCredentialInvalid = errors.New("integrations/runtime: credential invalid")
	// ErrCredentialNotDeclared indicates the credential is not declared on the resolved connection
	ErrCredentialNotDeclared = errors.New("integrations/runtime: credential not declared on connection")
	// ErrRuntimeClientNotFound indicates no pre-built runtime client exists for the requested definition
	ErrRuntimeClientNotFound = errors.New("integrations/runtime: runtime client not found")
)
