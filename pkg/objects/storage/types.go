package storage

import (
	"context"
	"io"
	"net/http"
	"time"

	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

// Alias types from storage/types to maintain clean imports
// having a bunch of smaller subpackages seemed to just complicate things
type (
	Provider             = storagetypes.Provider
	ProviderType         = storagetypes.ProviderType
	UploadFileOptions    = storagetypes.UploadFileOptions
	UploadedFileMetadata = storagetypes.UploadedFileMetadata
	DownloadFileOptions  = storagetypes.DownloadFileOptions
	DownloadFileMetadata = storagetypes.DownloadFileMetadata
	ProviderHints        = storagetypes.ProviderHints
	FileStorageMetadata  = storagetypes.FileStorageMetadata
)

// Provider type constants so we can range, switch, etc
const (
	S3Provider   = storagetypes.S3Provider
	R2Provider   = storagetypes.R2Provider
	GCSProvider  = storagetypes.GCSProvider
	DiskProvider = storagetypes.DiskProvider
)

// Configuration constants
const (
	DefaultMaxFileSize   = 32 << 20 // 32MB
	DefaultMaxMemory     = 32 << 20 // 32MB
	DefaultUploadFileKey = "uploadFile"
)

// Default function implementations
var (
	DefaultValidationFunc = func(_ context.Context, _ *UploadOptions) error {
		return nil
	}

	DefaultNameGeneratorFunc = func(originalName string) string {
		return originalName
	}

	DefaultSkipper = func(_ *http.Request) bool {
		return false
	}

	DefaultErrorResponseHandler = func(err error, statusCode int) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, err.Error(), statusCode)
		}
	}
)

// Integration in the objects package representation of information we store about the 3d party system
// This is being made so we can establish relationships between a "provider" (e.g. AWS) and the credentials used to access it
// and the organization that owns it. This is separate from the actual credentials which are stored encrypted in the "hush" system
// the combination of integration + hush secret + organization (or "system" for global) is what will drive multi-tenant support and "BYO-storage"
type Integration struct {
	// ID is the unique identifier for the integration
	ID string `json:"id"`
	// OrganizationID is the organization this integration belongs to
	OrganizationID string `json:"organization_id"`
	// ProviderType is the type of storage provider (s3, r2, gcs, disk)
	ProviderType ProviderType `json:"provider_type"`
	// HushID references the encrypted credentials
	HushID string `json:"hush_id"`
	// Config contains provider-specific configuration
	Config map[string]any `json:"config"`
	// Enabled indicates if this integration is active
	Enabled bool `json:"enabled"`
	// Name is a human-readable name for the integration
	Name string `json:"name"`
	// Description provides additional context about the integration
	Description string `json:"description,omitempty"`
}

// File represents a consolidated file object in the system that combines upload, storage, and metadata information
// This is the objects package representation of the file schema we use in ent
type File struct {
	// ID is the unique identifier for the file
	ID string `json:"id"`
	// Name is the display name of the file
	Name string `json:"name"`
	// OriginalName is the original filename from upload
	OriginalName string `json:"original_name,omitempty"`
	// FileStorageMetadata contains provider and storage information
	FileStorageMetadata
	// MD5 hash of the file for integrity checking
	MD5 []byte `json:"md5,omitempty"`
	// URI is the complete URI for accessing the file
	URI string `json:"uri"`
	// PresignedURL is a time-limited URL for direct access
	PresignedURL string `json:"presigned_url,omitempty"`
	// FieldName denotes the field from the multipart form
	FieldName string `json:"field_name,omitempty"`
	// FolderDestination is the folder that holds the uploaded file
	FolderDestination string `json:"folder_destination,omitempty"`
	// ProvidedExtension is the extension provided by the client
	ProvidedExtension string `json:"provided_extension,omitempty"`
	// CreatedAt is the time the file was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the time the file was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// Parent is the parent object of the file, if any
	Parent ParentObject `json:"parent,omitempty"`
	// Metadata contains additional file metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// UploadOptions contains options for uploading files
type UploadOptions struct {
	// FileName is the name of the file
	FileName string `json:"file_name"`
	// ContentType is the MIME type of the file
	ContentType string `json:"content_type"`
	// Metadata contains additional metadata for the file
	Metadata map[string]string `json:"metadata,omitempty"`
	// ProviderHints contains hints for provider selection
	ProviderHints any `json:"provider_hints,omitempty"`
}

// DownloadResult contains the result of a file download
type DownloadResult struct {
	// File contains the file data
	File []byte `json:"file"`
	// Size is the size of the downloaded file
	Size int64 `json:"size"`
}

// FileUpload represents a file upload from multipart form or GraphQL
type FileUpload struct {
	// File is the file data reader
	File io.ReadSeeker
	// Filename is the name of the uploaded file
	Filename string
	// Size is the size of the file in bytes
	Size int64
	// ContentType is the MIME type of the file
	ContentType string
	// Key is the field name from the form or GraphQL input
	Key string
	// CorrelatedObjectID is the ID of the object this file belongs to
	CorrelatedObjectID string
	// CorrelatedObjectType is the type of object this file belongs to
	CorrelatedObjectType string
}

// ValidationFunc validates upload options before uploading
type ValidationFunc func(ctx context.Context, opts *UploadOptions) error

// UploaderFunc handles the file upload process and returns uploaded files
type UploaderFunc func(ctx context.Context, service *ObjectService, files []FileUpload) ([]File, error)

// NameGeneratorFunc generates names for uploaded files
type NameGeneratorFunc func(originalName string) string

// SkipperFunc defines a function to skip middleware processing
type SkipperFunc func(r *http.Request) bool

// ErrResponseHandler is a custom error handler for upload failures
type ErrResponseHandler func(err error, statusCode int) http.HandlerFunc

// ParentObject represents the parent object of a file
type ParentObject struct {
	// ID is the unique identifier of the parent object
	ID string `json:"id"`
	// Type is the type of the parent object
	Type string `json:"type"`
}

// Files is a map of file uploads organized by key
type Files map[string][]File

// ProviderConfig contains configuration for object storage providers
type ProviderConfig struct {
	// Enabled indicates if object storage is enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"true"`
	// Keys are the form field keys that will be processed for uploads
	Keys []string `json:"keys" koanf:"keys" default:"[\"uploadFile\"]"`
	// MaxSizeMB is the maximum file size allowed in MB
	MaxSizeMB int64 `json:"maxSizeMB" koanf:"maxSizeMB"`
	// MaxMemoryMB is the maximum memory to use for file uploads in MB
	MaxMemoryMB int64 `json:"maxMemoryMB" koanf:"maxMemoryMB"`
	// DevMode enables simple file upload handling for local development and testing
	DevMode bool `json:"devMode" koanf:"devMode" default:"false"`
	// Providers contains configuration for each storage provider
	Providers ProviderConfigs `json:"providers" koanf:"providers"`
}

// ProviderConfigs contains configuration for all storage providers
// This is structured to allow easy extension for additional providers in the future
type ProviderConfigs struct {
	// S3 provider configuration
	S3 ProviderCredentials `json:"s3" koanf:"s3"`
	// CloudflareR2 provider configuration
	CloudflareR2 ProviderCredentials `json:"cloudflareR2" koanf:"cloudflareR2"`
	// GCS provider configuration
	GCS ProviderCredentials `json:"gcs" koanf:"gcs"`
	// Disk provider configuration
	Disk ProviderCredentials `json:"disk" koanf:"disk"`
}

// ProviderCredentials contains credentials and configuration for a storage provider
// Given most provides had a smiliar set of credentials and only some minor exceptions
// we consolidated into a single struct to make it easier to manage and extend
type ProviderCredentials struct {
	// Enabled indicates if this provider is enabled
	Enabled bool `json:"enabled" koanf:"enabled"`
	// AccessKeyID for cloud providers
	AccessKeyID string `json:"accessKeyID" koanf:"accessKeyID"`
	// SecretAccessKey for cloud providers
	SecretAccessKey string `json:"secretAccessKey" koanf:"secretAccessKey"`
	// Region for cloud providers
	Region string `json:"region" koanf:"region"`
	// Bucket name for cloud providers
	Bucket string `json:"bucket" koanf:"bucket"`
	// Endpoint for custom endpoints
	Endpoint string `json:"endpoint" koanf:"endpoint"`
	// ProjectID for GCS
	ProjectID string `json:"projectID" koanf:"projectID"`
	// CredentialsJSON for GCS
	CredentialsJSON string `json:"credentialsJSON" koanf:"credentialsJSON"`
	// AccountID for Cloudflare R2
	AccountID string `json:"accountID" koanf:"accountID"`
	// APIToken for Cloudflare R2
	APIToken string `json:"apiToken" koanf:"apiToken"`
}
