package reconciler

import "errors"

var (
	// ErrEmitFailureMissingPayload indicates emit failure details are incomplete
	ErrEmitFailureMissingPayload = errors.New("emit failure missing topic or payload")
)
