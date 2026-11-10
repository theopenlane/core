package relocate

import "errors"

var (
	// ErrMissingStorageKey is returned when a file record has no storage key to relocate
	ErrMissingStorageKey = errors.New("file has no storage key")
	// ErrTouchUnsupported is returned when the resolved provider cannot read a single byte of an object
	ErrTouchUnsupported = errors.New("storage provider does not support touching objects")
	// ErrStatUnsupported is returned when the resolved provider cannot report object metadata
	ErrStatUnsupported = errors.New("storage provider does not support reading object metadata")
)
