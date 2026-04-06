package celx

import "errors"

var (
	// ErrJSONMapExpected indicates a CEL value could not be converted into a JSON object map
	ErrJSONMapExpected = errors.New("expected JSON object")
	// ErrNilOutput indicates a CEL evaluation returned a nil value
	ErrNilOutput = errors.New("CEL evaluation returned nil output")
	// ErrTypeMismatch indicates the CEL result type did not match the expected type
	ErrTypeMismatch = errors.New("CEL expression type mismatch")
)
