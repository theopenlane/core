package jsonx

import "errors"

var (
	// ErrObjectExpected is returned when a JSON round-trip does not produce an object
	ErrObjectExpected = errors.New("json value is not an object")
)
