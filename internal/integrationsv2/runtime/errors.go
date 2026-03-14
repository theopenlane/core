package runtime

import "errors"

var (
	// ErrDBClientRequired indicates the Ent client dependency is missing
	ErrDBClientRequired = errors.New("integrationsv2/runtime: db client required")
	// ErrGalaRequired indicates the Gala runtime dependency is missing
	ErrGalaRequired = errors.New("integrationsv2/runtime: gala required")
	// ErrCredentialResolverRequired indicates the credential resolver dependency is missing
	ErrCredentialResolverRequired = errors.New("integrationsv2/runtime: credential resolver required")
)
