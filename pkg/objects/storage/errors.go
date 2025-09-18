package storage

import (
	"errors"
)

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
	// ErrProviderNotFound is returned when a requested storage provider is not found
	ErrProviderNotFound = errors.New("storage provider not found")
	// ErrInvalidProviderConfig is returned when a provider configuration is invalid
	ErrInvalidProviderConfig = errors.New("invalid provider configuration")
	// ErrNoStorageService is returned when no storage service is available
	ErrNoStorageService = errors.New("no storage service available")
	// ErrInvalidFileUpload is returned when file upload data is invalid
	ErrInvalidFileUpload = errors.New("invalid file upload data")
	// ErrContextMissingIntegration is returned when context is missing integration information
	ErrContextMissingIntegration = errors.New("context missing integration information")
	// ErrContextMissingCredentials is returned when context is missing credential information
	ErrContextMissingCredentials = errors.New("context missing credential information")
	// ErrNoStorageProviderAvailable is returned when no storage provider is available
	ErrNoStorageProviderAvailable = errors.New("no storage provider available")
	// ErrS3BucketRequired is returned when S3 bucket is not specified
	ErrS3BucketRequired = errors.New("S3 bucket is required")
	// ErrS3CredentialsRequired is returned when required S3 credentials are missing
	ErrS3CredentialsRequired = errors.New("missing required S3 credentials: bucket, region")
	// ErrR2CredentialsRequired is returned when required R2 credentials are missing
	ErrR2CredentialsRequired = errors.New("missing required R2 credentials: bucket, account_id, access_key_id, secret_access_key")
	// ErrDiskPathRequired is returned when disk base path is not provided
	ErrDiskPathRequired = errors.New("missing required disk credential: base_path")
	// ErrR2BucketRequired is returned when R2 bucket is not specified
	ErrR2BucketRequired = errors.New("R2 bucket is required")
	// ErrR2AccountIDRequired is returned when R2 account ID is not specified
	ErrR2AccountIDRequired = errors.New("R2 account ID is required")
	// ErrR2CredentialsMissing is returned when R2 access keys are missing
	ErrR2CredentialsMissing = errors.New("R2 access key ID and secret access key are required")
	// ErrDiskBasePathRequired is returned when disk base path is required
	ErrDiskBasePathRequired = errors.New("disk base path is required")
	// ErrJSONParseFailed is returned when JSON parsing fails
	ErrJSONParseFailed = errors.New("failed to parse JSON")
	// ErrYAMLParseFailed is returned when YAML parsing fails
	ErrYAMLParseFailed = errors.New("failed to parse YAML")
	// ErrFileKeyNotFound is returned when file key not found in multipart form
	ErrFileKeyNotFound = errors.New("file key not found")
	// ErrProviderResolutionRequired is returned when provider resolution is required but not available
	ErrProviderResolutionRequired = errors.New("provider resolution required - use external orchestration layer")
)

type errorMsg string

func (e errorMsg) Error() string { return string(e) }

const (
	// ErrNoFilesUploaded is returned when no files were uploaded in the request
	ErrNoFilesUploaded = errorMsg("objects: no uploadable files found in request")
)
