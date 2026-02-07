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
)

// LoaderPathError captures loader failures tied to a specific path.
type LoaderPathError struct {
	// Err is the base error for the loader failure
	Err   error
	// Path is the file or directory that failed to load
	Path  string
	// Cause is the underlying error returned by the filesystem or parser
	Cause error
}

// Error returns the base error message for the loader failure
func (e *LoaderPathError) Error() string {
	return e.Err.Error()
}

// Unwrap exposes the base error for errors.Is and errors.As
func (e *LoaderPathError) Unwrap() error {
	return e.Err
}

// SchemaVersionError captures schema version mismatch details.
type SchemaVersionError struct {
	// Path is the provider spec path that failed validation
	Path    string
	// Version is the unsupported schema version
	Version string
}

// Error returns the base schema version error message
func (e *SchemaVersionError) Error() string {
	return ErrSchemaVersionUnsupported.Error()
}

// Unwrap exposes the base schema version error for errors.Is and errors.As
func (e *SchemaVersionError) Unwrap() error {
	return ErrSchemaVersionUnsupported
}
