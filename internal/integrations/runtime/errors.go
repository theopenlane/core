package runtime

import "errors"

var (
	// ErrInstallationRequired indicates the installation record dependency is missing
	ErrInstallationRequired = errors.New("integrations/runtime: installation required")
	// ErrInstallationNotFound indicates no matching installation could be resolved
	ErrInstallationNotFound = errors.New("integrations/runtime: installation not found")
	// ErrConnectionRequired indicates the operation requires a credential-selected connection
	ErrConnectionRequired = errors.New("integrations/runtime: connection required")
	// ErrConnectionNotFound indicates the requested connection could not be resolved
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
	// ErrRuntimeClientNotFound indicates no pre-built runtime client exists for the definition
	ErrRuntimeClientNotFound = errors.New("integrations/runtime: runtime client not found")
	// ErrOperationRateLimited indicates the operation's RateLimit policy rejected this run
	ErrOperationRateLimited = errors.New("integrations/runtime: operation rate limited")
	// ErrInstallationInstanceIDRequired indicates the connection resolved no instance id
	ErrInstallationInstanceIDRequired = errors.New("integrations/runtime: installation instance id required")
	// ErrInstallationInstanceMismatch indicates the credential resolves to a different instance
	ErrInstallationInstanceMismatch = errors.New("integrations/runtime: installation instance mismatch")
	// ErrInstallationUpgradeFailed indicates the installation couldn't reach the current version
	ErrInstallationUpgradeFailed = errors.New("integrations/runtime: installation upgrade failed")
)
