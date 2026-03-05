package runtime

import "errors"

var (
	// ErrRegistryRequired indicates the integration registry dependency is missing.
	ErrRegistryRequired = errors.New("integrations runtime: registry required")
	// ErrDBClientRequired indicates the database client dependency is missing.
	ErrDBClientRequired = errors.New("integrations runtime: db client required")
)
