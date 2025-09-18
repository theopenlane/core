package storagetypes

import (
	"context"
	"io"
	"time"
)

//go:generate_input *.go
//go:generate mockery --config .mockery.yml

// Provider defines the interface for storage providers
type Provider interface {
	// Upload uploads a file to the storage provider
	Upload(ctx context.Context, reader io.Reader, opts *UploadFileOptions) (*UploadedFileMetadata, error)
	// Download downloads a file from the storage provider
	Download(ctx context.Context, opts *DownloadFileOptions) (*DownloadFileMetadata, error)
	// Delete deletes a file from the storage provider
	Delete(ctx context.Context, key string) error
	// GetPresignedURL generates a presigned URL for file access
	GetPresignedURL(key string, expires time.Duration) (string, error)
	// Exists checks if a file exists in the storage provider
	Exists(ctx context.Context, key string) (bool, error)
	// GetScheme returns the URL scheme for this provider
	GetScheme() *string
	// Close closes any resources used by the provider
	Close() error
}

// ProviderType represents the type of storage provider
type ProviderType string

const (
	S3Provider   ProviderType = "s3"
	R2Provider   ProviderType = "r2"
	GCSProvider  ProviderType = "gcs"
	DiskProvider ProviderType = "disk"
)

// UploadFileOptions contains options for uploading files
type UploadFileOptions struct {
	// FileName is the name/key for the uploaded file
	FileName string `json:"file_name"`
	// ContentType is the MIME type of the file
	ContentType string `json:"content_type"`
	// Metadata contains additional metadata for the file
	Metadata map[string]string `json:"metadata,omitempty"`
	// ProviderHints contains hints for provider selection
	ProviderHints *ProviderHints `json:"provider_hints,omitempty"`
}

// FileStorageMetadata contains common metadata about file storage operations
type FileStorageMetadata struct {
	// Key is the storage key/path for the file
	Key string `json:"key"`
	// Size is the size of the file in bytes
	Size int64 `json:"size"`
	// ContentType is the MIME type of the file
	ContentType string `json:"content_type,omitempty"`
	// IntegrationID references the integration used for storage
	IntegrationID string `json:"integration_id,omitempty"`
	// HushID references the credentials used for storage
	HushID string `json:"hush_id,omitempty"`
	// OrganizationID identifies the owning organization
	OrganizationID string `json:"organization_id,omitempty"`
	// ProviderType indicates which storage provider was used
	ProviderType ProviderType `json:"provider_type,omitempty"`
	// Bucket is the storage bucket name
	Bucket string `json:"bucket,omitempty"`
}

// UploadedFileMetadata contains metadata about an uploaded file
type UploadedFileMetadata struct {
	FileStorageMetadata
	// FolderDestination is the folder that holds the uploaded file
	FolderDestination string `json:"folder_destination,omitempty"`
}

// DownloadFileOptions contains options for downloading files
type DownloadFileOptions struct {
	// FileName is the storage key/path for the file to download
	FileName string `json:"file_name"`
	// Metadata contains additional metadata for the download
	Metadata map[string]string `json:"metadata,omitempty"`
	// IntegrationID references the integration used for storage
	IntegrationID string `json:"integration_id,omitempty"`
	// HushID references the credentials used for storage
	HushID string `json:"hush_id,omitempty"`
	// OrganizationID identifies the owning organization
	OrganizationID string `json:"organization_id,omitempty"`
}

// DownloadFileMetadata contains metadata about a downloaded file
type DownloadFileMetadata struct {
	// File contains the file data
	File []byte `json:"file"`
	// Size is the size of the downloaded file
	Size int64 `json:"size"`
	// Writer is the writer used for streaming downloads
	Writer any `json:"-"`
}

// ProviderHints contains hints for provider selection and configuration
type ProviderHints struct {
	// PreferredProvider indicates the preferred provider type
	PreferredProvider ProviderType `json:"preferred_provider,omitempty"`
	// OrganizationID is the organization ID for provider selection
	OrganizationID string `json:"organization_id,omitempty"`
	// IntegrationID is the specific integration to use
	IntegrationID string `json:"integration_id,omitempty"`
	// HushID is the specific credential set to use
	HushID string `json:"hush_id,omitempty"`
	// ContentType helps with provider selection based on file type
	ContentType string `json:"content_type,omitempty"`
	// Size helps with provider selection based on file size
	Size int64 `json:"size,omitempty"`
	// Feature indicates a more granual feature than we could specify with a type like OrgModule
	Feature string `json:"feature,omitempty"`
	// Module indicates the specific module within a feature
	Module string `json:"module,omitempty"`
	// Metadata contains additional hints for provider selection
	Metadata map[string]string `json:"metadata,omitempty"`
}
