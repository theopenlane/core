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
	Download(ctx context.Context, file *File, opts *DownloadFileOptions) (*DownloadedFileMetadata, error)
	// Delete deletes a file from the storage provider
	Delete(ctx context.Context, file *File, opts *DeleteFileOptions) error
	// GetPresignedURL generates a presigned URL for file access
	GetPresignedURL(ctx context.Context, file *File, opts *PresignedURLOptions) (string, error)
	// Exists checks if a file exists in the storage provider
	Exists(ctx context.Context, file *File) (bool, error)
	// GetScheme returns the URL scheme for this provider
	GetScheme() *string
	// ListBuckets is used to list the buckets in the storage backend
	ListBuckets() ([]string, error)
	// ProviderType returns the type of the storage provider (e.g., S3, R2)
	ProviderType() ProviderType
	io.Closer
}

// ProviderType represents the type of storage provider
type ProviderType string

const (
	// S3Provider is the type for AWS S3 storage
	S3Provider ProviderType = "s3"
	// R2Provider is the type for Cloudflare R2 storage
	R2Provider ProviderType = "r2"
	// DiskProvider is the type for local disk storage
	DiskProvider ProviderType = "disk"
	// DatabaseProvider is the type for database storage
	DatabaseProvider ProviderType = "database"
)

// PresignMode determines how presigned URLs are generated for a provider
type PresignMode string

const (
	// PresignModeProvider uses the upstream storage provider to generate presigned URLs
	PresignModeProvider PresignMode = "provider"
	// PresignModeProxy uses the internal token-based proxy for presigned URLs
	PresignModeProxy PresignMode = "proxy"
)

// Valid reports whether the presign mode is recognized
func (m PresignMode) Valid() bool {
	return m == PresignModeProvider || m == PresignModeProxy
}

// ResolvePresignMode returns a valid presign mode, falling back to the provided default when necessary.
func ResolvePresignMode(value PresignMode, fallback PresignMode) PresignMode {
	if value.Valid() {
		return value
	}

	return fallback
}

// File represents a consolidated file object in the system that combines upload, storage, and metadata information
// This is the objects package representation of the file schema we use in ent
type File struct {
	// File is the file to be uploaded
	RawFile io.ReadSeeker
	// ID is the unique identifier for the file
	ID string `json:"id"`
	// OriginalName is the original filename that was provided by the client on submission
	OriginalName string `json:"original_name,omitempty"`
	// FieldName denotes the field from the multipart form
	FieldName string `json:"field_name,omitempty"`
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
	// Metadata contains additional file metadata
	Metadata map[string]string `json:"metadata,omitempty"`
	// CorrelatedObjectID is the ID of the object this file belongs to
	CorrelatedObjectID string
	// CorrelatedObjectType is the type of object this file belongs to
	CorrelatedObjectType string
	// ProviderType indicates which storage provider is used for this file
	ProviderType ProviderType `json:"provider_type,omitempty"`
	// FileMetadata contains common metadata about the file
	FileMetadata
}

// ParentObject represents the parent object of a file
type ParentObject struct {
	// ID is the unique identifier of the parent object
	ID string `json:"id"`
	// Type is the type of the parent object
	Type string `json:"type"`
}

// UploadFileOptions contains options for uploading files
type UploadFileOptions struct {
	// FileName is the name/key for the uploaded file
	FileName string `json:"file_name"`
	// ContentType is the MIME type of the file
	ContentType string `json:"content_type"`
	// Bucket is the storage bucket name
	Bucket string `json:"bucket,omitempty"`
	// FolderDestination is the folder/path within the bucket
	FolderDestination string `json:"folder_destination,omitempty"`
	// FileMetadata contains common metadata about the file
	FileMetadata
}

// UploadedFileMetadata contains metadata about an uploaded file
type UploadedFileMetadata struct {
	// TimeUploaded is the time the file was uploaded
	TimeUploaded time.Time `json:"time_uploaded"`
	// ElapsedTime is the duration taken to upload the file
	ElapsedTime time.Duration `json:"elapsed_time"`
	// FileMetadata contains common metadata about the file
	FileMetadata
}

// FileMetadata contains common metadata about file storage operations
type FileMetadata struct {
	// Key is the key that was used when originally parsing out the file (its source, so to speak)
	Key string `json:"key"`
	// Bucket is the bucket where the file is stored
	Bucket string `json:"bucket,omitempty"`
	// Region is the region where the file is stored
	Region string `json:"region,omitempty"`
	// Folder is the folder/path within the bucket the file is stored
	Folder string `json:"folder,omitempty"`
	// FullURI is the full URI to access the file which would include the storage scheme, e.g. s3://bucket/folder/file
	FullURI string `json:"full_uri,omitempty"`
	// Size is the size of the file in bytes
	Size int64 `json:"size"`
	// ContentType is the MIME type of the file
	ContentType string `json:"content_type,omitempty"`
	// ProviderType indicates which storage provider was used
	ProviderType ProviderType `json:"provider_type,omitempty"`
	// PresignedURL is the URL that can be used to download the file
	PresignedURL string `json:"presigned_url,omitempty"`
	// Name is the display name of the file
	Name string `json:"name,omitempty"`
	// ProviderHints contains hints for provider selection
	ProviderHints *ProviderHints `json:"provider_hints,omitempty"`
}

// DownloadFileOptions contains options for downloading files
type DownloadFileOptions struct {
	// FileName is the name you'd like the file to be saved as when downloading
	FileName string `json:"file_name"`
	// ContentType is the MIME type of the file to be set for the download
	ContentType string `json:"content_type"`
	// DownloadFileLocation is the local path to save the downloaded file which does not include the name
	DownloadFileLocation string `json:"download_file_location,omitempty"`
	// Writer can be passed in if you want to provide a specific io.Writer or similar
	Writer any `json:"-"`
	// FileMetadata contains common metadata about the file
	FileMetadata
}

// DownloadedFileMetadata contains metadata about a downloaded file
type DownloadedFileMetadata struct {
	// File contains the file data
	File []byte `json:"file"`
	// Size is the size of the downloaded file
	Size int64 `json:"size"`
	// TimeDownloaded is the time the file was downloaded
	TimeDownloaded time.Time `json:"time_downloaded"`
	// ElapsedTime is the duration taken to download the file
	ElapsedTime time.Duration `json:"elapsed_time"`
	FileMetadata
}

// ProviderHints contains hints for provider selection and configuration
type ProviderHints struct {
	// KnownProvider indicates a known provider type
	KnownProvider ProviderType `json:"known_provider,omitempty"`
	// PreferredProvider indicates the preferred provider type
	PreferredProvider ProviderType `json:"preferred_provider,omitempty"`
	// OrganizationID is the organization ID for provider selection
	OrganizationID string `json:"organization_id,omitempty"`
	// IntegrationID is the specific integration to use
	IntegrationID string `json:"integration_id,omitempty"`
	// HushID is the specific credential set to use
	HushID string `json:"hush_id,omitempty"`
	// Module indicates the specific module within a feature
	Module any `json:"module,omitempty"`
	// Metadata contains additional hints for provider selection
	Metadata map[string]string `json:"metadata,omitempty"`
}

// PresignedURLOptions contains options for generating presigned URLs
type PresignedURLOptions struct {
	// Duration is the duration the presigned URL should be valid for
	Duration time.Duration `json:"duration"`
}

// DeleteFileOptions contains options for deleting files
type DeleteFileOptions struct {
	// Reason is the reason for deleting the file
	Reason string `json:"reason,omitempty"`
}
