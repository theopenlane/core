package storage

import "errors"

var (
	// ErrJSONParseFailed is returned when JSON parsing fails
	ErrJSONParseFailed = errors.New("failed to parse JSON")
	// ErrYAMLParseFailed is returned when YAML parsing fails
	ErrYAMLParseFailed = errors.New("failed to parse YAML")
	// ErrProviderResolutionRequired is returned when provider resolution is required but not available
	ErrProviderResolutionRequired = errors.New("provider resolution required - use external orchestration layer")
)

type errorMsg string

func (e errorMsg) Error() string { return string(e) }

const (
	// ErrNoFilesUploaded is returned when no files were uploaded in the request
	ErrNoFilesUploaded = errorMsg("objects: no uploadable files found in request")
)
