package relocate

import "errors"

var (
	// ErrMissingStorageKey is returned when a file record has no storage key to relocate
	ErrMissingStorageKey = errors.New("file has no storage key")
	// ErrUnknownProvider is returned when a record names no provider and its URI has no scheme
	ErrUnknownProvider = errors.New("file records no storage provider")
	// ErrTouchUnsupported is returned when the target provider cannot touch objects
	ErrTouchUnsupported = errors.New("storage provider does not support touching objects")
	// ErrCopyUnsupported is returned when the target provider cannot copy objects between buckets
	ErrCopyUnsupported = errors.New("storage provider does not support copying objects")
	// ErrOpenUnsupported is returned when the source provider cannot stream objects
	ErrOpenUnsupported = errors.New("storage provider does not support streaming objects")
	// ErrKeyChanged is returned when the target stored a transferred object under another key
	ErrKeyChanged = errors.New("storage provider changed the object key")
)
