package runtime

import "errors"

var (
	// ErrOwnerIDRequired indicates installation resolution requires an owner id when resolving by definition
	ErrOwnerIDRequired = errors.New("integrationsv2/runtime: owner id required")
	// ErrInstallationRequired indicates the installation record dependency is missing
	ErrInstallationRequired = errors.New("integrationsv2/runtime: installation required")
	// ErrInstallationIDRequired indicates installation resolution requires an installation id when owner plus definition is ambiguous
	ErrInstallationIDRequired = errors.New("integrationsv2/runtime: installation id required")
	// ErrInstallationNotFound indicates no matching installation could be resolved
	ErrInstallationNotFound = errors.New("integrationsv2/runtime: installation not found")
	// ErrInstallationDefinitionMismatch indicates the resolved installation does not match the requested definition
	ErrInstallationDefinitionMismatch = errors.New("integrationsv2/runtime: installation definition mismatch")
	// ErrCredentialNotFound indicates no credential could be resolved for the installation
	ErrCredentialNotFound = errors.New("integrationsv2/runtime: credential not found")
)
