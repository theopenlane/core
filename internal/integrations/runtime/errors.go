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
	// ErrConnectionMismatch indicates the credential belongs to a different connection than the installation's
	ErrConnectionMismatch = errors.New("integrations/runtime: credential belongs to a different connection than the installation's")
	// ErrInstallationMetadataInvalid indicates the installation metadata document failed schema validation
	ErrInstallationMetadataInvalid = errors.New("integrations/runtime: installation metadata invalid")
	// ErrRuntimeClientNotFound indicates no pre-built runtime client exists for the definition
	ErrRuntimeClientNotFound = errors.New("integrations/runtime: runtime client not found")
	// ErrOperationRateLimited indicates the operation's RateLimit policy rejected this run
	ErrOperationRateLimited = errors.New("integrations/runtime: operation rate limited")
	// ErrInstallationInstanceIDRequired indicates the connection resolved no instance id
	ErrInstallationInstanceIDRequired = errors.New("integrations/runtime: installation instance id required")
	// ErrInstallationInstanceMismatch indicates the credential resolves to a different instance
	ErrInstallationInstanceMismatch = errors.New("integrations/runtime: installation instance mismatch")
	// ErrInstallationVersionAhead indicates a newer binary stamped the installation with a definition version this binary does not have
	ErrInstallationVersionAhead = errors.New("integrations/runtime: installation definition version is ahead of this binary")
	// ErrInstallationUpgradeFailed indicates the installation couldn't reach the current version
	ErrInstallationUpgradeFailed = errors.New("integrations/runtime: installation upgrade failed")
	// ErrClientUnresolved indicates the installation could not establish a connection and needs to be reconnected
	ErrClientUnresolved = errors.New("the integration could not establish a connection and needs to be reconnected")
	// ErrReconcileExhausted indicates a reconcile loop exhausted its error budget
	ErrReconcileExhausted = errors.New("repeated sync failures")
)
