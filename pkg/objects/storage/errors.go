package storage

import "errors"

var (
	// ErrInvalidS3Bucket is returned when an invalid s3 bucket is provided
	ErrInvalidS3Bucket = errors.New("invalid s3 bucket provided")
	// ErrInvalidFolderPath is returned when an invalid folder path is provided
	ErrInvalidFolderPath = errors.New("invalid folder path provided")
	// ErrMissingRequiredAWSParams is returned when required AWS parameters are missing
	ErrMissingRequiredAWSParams = errors.New("missing required AWS parameters")
	// ErrMissingLocalURL = errors.New("missing local URL in disk storage options"
	ErrMissingLocalURL = errors.New("missing local URL in disk storage options")
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
