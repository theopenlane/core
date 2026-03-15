package runtime

import "errors"

var (
	// ErrDBClientRequired indicates the Ent client dependency is missing
	ErrDBClientRequired = errors.New("integrationsv2/runtime: db client required")
	// ErrGalaRequired indicates the Gala runtime dependency is missing
	ErrGalaRequired = errors.New("integrationsv2/runtime: gala required")
	// ErrCredentialStoreRequired indicates the credential store dependency is missing
	ErrCredentialStoreRequired = errors.New("integrationsv2/runtime: credential store required")
	// ErrAuthStateStoreRequired indicates the auth state store dependency is missing
	ErrAuthStateStoreRequired = errors.New("integrationsv2/runtime: auth state store required")
)
