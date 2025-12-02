package objstore

import "errors"

var (
	// ErrProviderResolutionFailed is returned when storage provider resolution failed
	ErrProviderResolutionFailed = errors.New("storage provider resolution failed")
	// ErrProviderCreationFailed is returned when storage provider creation failed
	ErrProviderCreationFailed = errors.New("storage provider creation failed")
	// ErrInvalidUploadOptions is returned when invalid upload options provided
	ErrInvalidUploadOptions = errors.New("invalid upload options")
	// ErrInvalidIntegration is returned when an invalid integration is provided
	ErrInvalidIntegration = errors.New("invalid integration provided")
	// ErrNoOrganizationID is returned when no organization ID is available
	ErrNoOrganizationID = errors.New("no organization ID available")
	// ErrMissingIntegrationID is returned when integration ID is missing
	ErrMissingIntegrationID = errors.New("integration ID is missing")
	// ErrMissingHushID is returned when hush ID is missing
	ErrMissingHushID = errors.New("hush ID is missing")
	// ErrSystemIntegrationMissingOrgID is returned when system integration is missing organization ID
	ErrSystemIntegrationMissingOrgID = errors.New("system integration missing organization ID")
	// ErrNoIntegrationOrCredentials is returned when no integration or credentials are available
	ErrNoIntegrationOrCredentials = errors.New("no integration or credentials available")
	// ErrNoSystemIntegration is returned when no system integration is found for a provider
	ErrNoSystemIntegration = errors.New("no system integration found")
	// ErrNoIntegrationWithSecrets is returned when an integration lacks associated secrets
	ErrNoIntegrationWithSecrets = errors.New("no active integration with secrets found")
	// ErrMissingFileID is returned when file ID is required but missing
	ErrMissingFileID = errors.New("file id required for presigned URL")
)
