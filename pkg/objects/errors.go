package objects

import (
	"errors"
	"net/http"
)

// ErrResponseHandler is a custom error that should be used to handle errors when
// an upload fails
type ErrResponseHandler func(err error, statusCode int) http.HandlerFunc

var (
	// ErrFilesNotFound is returned when files could not be found in key from http request
	ErrFilesNotFound = errors.New("files could not be found in key from http request")
	// ErrFileOpenFailed is returned when a file could not be opened
	ErrFileOpenFailed = errors.New("could not open file")
	// ErrInvalidMimeType is returned when a file has an invalid mime type
	ErrInvalidMimeType = errors.New("invalid mimetype")
	// ErrValidationFailed is returned when a validation fails
	ErrValidationFailed = errors.New("validation failed")
	// ErrUnsupportedMimeType is returned when a file has an unsupported mime type
	ErrUnsupportedMimeType = errors.New("unsupported mime type uploaded")
	// ErrMustProvideStorageBackend is returned when a storage backend is not provided
	ErrMustProvideStorageBackend = errors.New("you must provide a storage backend")
	// ErrUnexpectedType is returned when an invalid type is provided
	ErrUnexpectedType = errors.New("unexpected type provided")
	// ErrSeekError is returned when an error occurs while seeking
	ErrSeekError = errors.New("error seeking")
)

type errorMsg string

func (e errorMsg) Error() string { return string(e) }

const (
	// ErrNoFilesUploaded is returned when no files were uploaded in the request
	ErrNoFilesUploaded = errorMsg("objects: no uploadable files found in request")
)
