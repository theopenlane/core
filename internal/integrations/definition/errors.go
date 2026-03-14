package definition

import "errors"

var (
	// ErrBuilderNil indicates a builder dependency was nil
	ErrBuilderNil = errors.New("integrationsv2/definition: builder is nil")
	// ErrRegistryRequired indicates the definition registry dependency is missing
	ErrRegistryRequired = errors.New("integrationsv2/definition: registry required")
)
