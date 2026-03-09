package jsonx

import "errors"

var (
	// ErrObjectExpected is returned when a JSON round-trip does not produce an object
	ErrObjectExpected = errors.New("json value is not an object")
	// ErrKeyRequired is returned when a JSON object key is empty
	ErrKeyRequired = errors.New("json key is required")
)
