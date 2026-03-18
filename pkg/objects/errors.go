package objects

import "errors"

var (
	// ErrNoStorageProvider is returned when no storage provider is available
	ErrNoStorageProvider = errors.New("no storage provider available")
	// ErrInsufficientProviderInfo is returned when insufficient information to resolve storage client
	ErrInsufficientProviderInfo = errors.New("insufficient information to resolve storage client: need integration_id+hush_id or organization_id")
	// ErrProviderHintsRequired is returned when provider hints are required for file upload
	ErrProviderHintsRequired = errors.New("provider hints required for file upload")
	// ErrReaderCannotBeNil is returned when a nil reader is provided to BufferedReader
	ErrReaderCannotBeNil = errors.New("reader cannot be nil")
	// ErrFailedToReadData is returned when reading data from a reader fails
	ErrFailedToReadData = errors.New("failed to read data from reader")
	// ErrFileSizeExceedsLimit is returned when file size exceeds the specified limit
	ErrFileSizeExceedsLimit = errors.New("file size exceeds limit")
	// ErrUnsupportedMimeType is returned when an unsupported mime type is uploaded
	ErrUnsupportedMimeType = errors.New("unsupported mime type uploaded")
)
