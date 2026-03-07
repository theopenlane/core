package runtime

import "errors"

var (
	// ErrDBClientRequired indicates the database client dependency is missing.
	ErrDBClientRequired = errors.New("integrations runtime: db client required")
)
