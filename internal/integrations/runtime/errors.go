package runtime

import "errors"

var (
	// ErrInstallationRequired indicates the installation record dependency is missing
	ErrInstallationRequired = errors.New("integrations/runtime: installation required")
	// ErrIntegrationIDRequired indicates resolution requires an explicit integration ID
	ErrIntegrationIDRequired = errors.New("integrations/runtime: integration id required")
	// ErrInstallationDefinitionMismatch indicates the resolved installation does not match the requested definition
	ErrInstallationDefinitionMismatch = errors.New("integrations/runtime: installation definition mismatch")
	// ErrConnectionRequired indicates the operation requires a credential-selected connection
	ErrConnectionRequired = errors.New("integrations/runtime: connection required")
	// ErrConnectionNotFound indicates the requested connection could not be resolved
	ErrConnectionNotFound = errors.New("integrations/runtime: connection not found")
	// ErrDefinitionNotFound indicates the requested integration definition is not registered
	ErrDefinitionNotFound = errors.New("integrations/runtime: definition not found")
	// ErrOperationNotFound indicates the requested operation is not registered for the definition
	ErrOperationNotFound = errors.New("integrations/runtime: operation not found")
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
	// ErrLegacyConfigAmbiguous indicates a stored client config key belongs to more than one operation
	ErrLegacyConfigAmbiguous = errors.New("integrations/runtime: legacy client config key claimed by multiple operations")
)
