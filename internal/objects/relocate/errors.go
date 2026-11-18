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
	// ErrPullFailed is returned when reading the object through the target did not succeed
	ErrPullFailed = errors.New("failed to read object through target bucket")
	// ErrCopyFailed is returned when the target could not copy the object from its current bucket
	ErrCopyFailed = errors.New("failed to copy object into target bucket")
	// ErrResolveSourceFailed is returned when the provider a record is stored on cannot be resolved
	ErrResolveSourceFailed = errors.New("failed to resolve source provider")
	// ErrOpenSourceFailed is returned when the object cannot be opened on its source provider
	ErrOpenSourceFailed = errors.New("failed to open object on source provider")
	// ErrUploadFailed is returned when the object cannot be uploaded to the target
	ErrUploadFailed = errors.New("failed to upload object to target bucket")
	// ErrConfirmFailed is returned when the target bucket cannot be listed to confirm the object landed
	ErrConfirmFailed = errors.New("failed to list target bucket for object")
	// ErrStatFailed is returned when the landed object's metadata cannot be read
	ErrStatFailed = errors.New("failed to read landed object metadata")
	// ErrRepointFailed is returned when the file record cannot be updated to the target
	ErrRepointFailed = errors.New("failed to update file record to target")
)
