package celx

import "errors"

var (
	// ErrJSONMapExpected indicates a CEL value could not be converted into a JSON object map
	ErrJSONMapExpected = errors.New("expected JSON object")
	// ErrNilOutput indicates a CEL evaluation returned a nil value
	ErrNilOutput = errors.New("CEL evaluation returned nil output")
	// ErrTypeMismatch indicates the CEL result type did not match the expected type
	ErrTypeMismatch = errors.New("CEL expression type mismatch")
	// ErrCompileFailed indicates a CEL expression failed to compile or type-check
	ErrCompileFailed = errors.New("CEL compilation failed")
	// ErrEntityDataInvalid indicates entity JSON could not be unmarshaled for expression evaluation
	ErrEntityDataInvalid = errors.New("failed to unmarshal entity data for evaluation")
)
