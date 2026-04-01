package providerkit

import "errors"

var (
	// ErrFilterExprEval is returned when a filter CEL expression fails during evaluation
	ErrFilterExprEval = errors.New("filter expression evaluation failed")
	// ErrMapExprEval is returned when a map CEL expression fails during evaluation
	ErrMapExprEval = errors.New("map expression evaluation failed")
)
