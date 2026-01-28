package reconciler

import "errors"

var (
	// ErrNilClient indicates the reconciler requires a client
	ErrNilClient = errors.New("workflow reconciler requires client")
	// ErrEmitFailureMissingPayload indicates emit failure details are incomplete
	ErrEmitFailureMissingPayload = errors.New("emit failure missing topic or payload")
)
