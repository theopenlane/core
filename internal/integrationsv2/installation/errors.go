package installation

import "errors"

var (
	// ErrDBClientRequired indicates the Ent client dependency is missing
	ErrDBClientRequired = errors.New("integrationsv2/installation: db client required")
	// ErrInstallationIDRequired indicates the installation identifier is missing
	ErrInstallationIDRequired = errors.New("integrationsv2/installation: installation id required")
	// ErrOwnerIDRequired indicates the owner identifier is missing for an org-owned installation
	ErrOwnerIDRequired = errors.New("integrationsv2/installation: owner id required")
	// ErrDefinitionIDRequired indicates the definition identifier is missing
	ErrDefinitionIDRequired = errors.New("integrationsv2/installation: definition id required")
	// ErrInstallationNameRequired indicates the installation name is missing
	ErrInstallationNameRequired = errors.New("integrationsv2/installation: installation name required")
)
