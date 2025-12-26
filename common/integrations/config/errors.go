package config

import "errors"

var (
	// ErrFSLoaderNotConfigured indicates the filesystem loader was not properly initialized
	ErrFSLoaderNotConfigured = errors.New("config: fs loader not configured")
	// ErrReadDirectory indicates a failure reading the provider specs directory
	ErrReadDirectory = errors.New("config: failed to read directory")
	// ErrReadFile indicates a failure reading a provider spec file
	ErrReadFile = errors.New("config: failed to read file")
	// ErrDecodeSpec indicates a failure decoding a provider spec
	ErrDecodeSpec = errors.New("config: failed to decode provider spec")
	// ErrRawBytesProviderRead indicates rawBytesProvider does not support the Read operation
	ErrRawBytesProviderRead = errors.New("config: rawBytesProvider does not support Read")
	// ErrSchemaVersionUnsupported indicates a provider spec declares an unknown schema version.
	ErrSchemaVersionUnsupported = errors.New("integrations: schema version unsupported")
	// ErrLoaderRequired indicates a loader dependency was omitted.
	ErrLoaderRequired = errors.New("integrations: loader required")
)
