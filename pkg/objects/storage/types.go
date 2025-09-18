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
	Provider           = storagetypes.Provider
	ProviderType       = storagetypes.ProviderType
	UploadOptions      = storagetypes.UploadFileOptions
	UploadedMetadata   = storagetypes.UploadedFileMetadata
	DownloadOptions    = storagetypes.DownloadFileOptions
	DownloadedMetadata = storagetypes.DownloadedFileMetadata
	ProviderHints      = storagetypes.ProviderHints
	FileMetadata       = storagetypes.FileMetadata
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
	DefaultValidationFunc ValidationFunc = func(_ File) error {
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

// File represents a consolidated file object in the system that combines upload, storage, and metadata information
// This is the objects package representation of the file schema we use in ent
type File struct {
	// File is the file to be uploaded
	RawFile io.ReadSeeker
	// ID is the unique identifier for the file
	ID string `json:"id"`
	// OriginalName is the original filename that was provided by the client on submission
	OriginalName string `json:"original_name,omitempty"`
	// MD5 hash of the file for integrity checking
	MD5 []byte `json:"md5,omitempty"`
	// ProvidedExtension is the extension provided by the client
	ProvidedExtension string `json:"provided_extension,omitempty"`
	// CreatedAt is the time the file was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the time the file was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// Parent is the parent object of the file, if any
	Parent ParentObject `json:"parent"`
	// FieldName denotes the field from the multipart form
	FieldName string `json:"field_name,omitempty"`
	// Metadata contains additional file metadata
	Metadata map[string]string `json:"metadata,omitempty"`
	// CorrelatedObjectID is the ID of the object this file belongs to
	CorrelatedObjectID string
	// CorrelatedObjectType is the type of object this file belongs to
	CorrelatedObjectType string
	ProviderType         ProviderType `json:"provider_type,omitempty"`
	FileMetadata
}

// ParentObject represents the parent object of a file
type ParentObject struct {
	// ID is the unique identifier of the parent object
	ID string `json:"id"`
	// Type is the type of the parent object
	Type string `json:"type"`
}

// ValidationFunc is a type that can be used to dynamically validate a file
type ValidationFunc func(f File) error

// UploaderFunc handles the file upload process and returns uploaded files
type UploaderFunc func(ctx context.Context, service *ObjectService, files []File) ([]File, error)

// NameGeneratorFunc generates names for uploaded files
type NameGeneratorFunc func(originalName string) string

// SkipperFunc defines a function to skip middleware processing
type SkipperFunc func(r *http.Request) bool

// ErrResponseHandler is a custom error handler for upload failures
type ErrResponseHandler func(err error, statusCode int) http.HandlerFunc

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
	Providers Providers `json:"providers" koanf:"providers"`
}

type Providers struct {
	// S3 provider configuration
	S3 ProviderConfigs `json:"s3" koanf:"s3"`
	// CloudflareR2 provider configuration
	CloudflareR2 ProviderConfigs `json:"cloudflareR2" koanf:"cloudflareR2"`
	// GCS provider configuration
	GCS ProviderConfigs `json:"gcs" koanf:"gcs"`
	// Disk provider configuration
	Disk ProviderConfigs `json:"disk" koanf:"disk"`
}

// ProviderConfigs contains configuration for all storage providers
// This is structured to allow easy extension for additional providers in the future
type ProviderConfigs struct {
	// Enabled indicates if this provider is enabled
	Enabled bool `json:"enabled" koanf:"enabled"`
	// Region for cloud providers
	Region string `json:"region" koanf:"region"`
	// Bucket name for cloud providers
	Bucket string `json:"bucket" koanf:"bucket"`
	// Endpoint for custom endpoints
	Endpoint string `json:"endpoint" koanf:"endpoint"`
	// Credentials contains the credentials for accessing the provider
	Credentials ProviderCredentials `json:"credentials" koanf:"credentials"`
}

// ProviderCredentials contains credentials for a storage provider
type ProviderCredentials struct {
	// AccessKeyID for cloud providers
	AccessKeyID string `json:"accessKeyID" koanf:"accessKeyID"`
	// SecretAccessKey for cloud providers
	SecretAccessKey string `json:"secretAccessKey" koanf:"accessKeySecret"`
	// Endpoint for custom endpoints
	Endpoint string `json:"endpoint" koanf:"endpoint"`
	// ProjectID for GCS
	ProjectID string `json:"projectID" koanf:"projectID"`
	// AccountID for Cloudflare R2
	AccountID string `json:"accountID" koanf:"accountID"`
	// APIToken for Cloudflare R2
	APIToken string `json:"apiToken" koanf:"apiToken"`
}
