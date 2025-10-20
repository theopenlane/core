package proxy

import "errors"

var (
	// ErrMissingFileID indicates that the file ID is missing
	ErrMissingFileID = errors.New("missing file id")
	// ErrMissingObjectURI indicates that the file metadata is missing the object URI
	ErrMissingObjectURI = errors.New("file metadata missing object URI")
	// ErrInvalidSecretLength indicates that the secret must be 128 bytes when provided
	ErrInvalidSecretLength = errors.New("secret must be 128 bytes when provided")
	// ErrAuthenticatedUserRequired indicates that an authenticated user is required in the context
	ErrAuthenticatedUserRequired = errors.New("authenticated user required in context")
)
