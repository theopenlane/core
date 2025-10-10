package storage

import (
	"context"
	"maps"
	"net/http"

	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

// Alias types from storage/types to maintain clean imports
// having a bunch of smaller subpackages seemed to just complicate things
type (
	Provider           = storagetypes.Provider
	ProviderType       = storagetypes.ProviderType
	File               = storagetypes.File
	UploadOptions      = storagetypes.UploadFileOptions
	UploadedMetadata   = storagetypes.UploadedFileMetadata
	DownloadOptions    = storagetypes.DownloadFileOptions
	DownloadedMetadata = storagetypes.DownloadedFileMetadata
	ProviderHints      = storagetypes.ProviderHints
	FileMetadata       = storagetypes.FileMetadata
	ParentObject       = storagetypes.ParentObject
)

// Provider type constants so we can range, switch, etc
const (
	S3Provider       = storagetypes.S3Provider
	R2Provider       = storagetypes.R2Provider
	GCSProvider      = storagetypes.GCSProvider
	DiskProvider     = storagetypes.DiskProvider
	DatabaseProvider = storagetypes.DatabaseProvider
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
	// EnsureAvailable enforces provider availability before completing server startup
	EnsureAvailable bool `json:"ensureAvailable" koanf:"ensureAvailable" default:"false"`
	// CredentialSync contains options for synchronizing provider credentials into the database
	CredentialSync CredentialSyncConfig `json:"credentialSync" koanf:"credentialSync"`
	// Providers contains configuration for each storage provider
	Providers Providers `json:"providers" koanf:"providers"`
}

// CredentialSyncConfig controls whether provider credentials are synchronized into the database
type CredentialSyncConfig struct {
	// Enabled indicates whether credential synchronization runs on startup
	Enabled bool `json:"enabled" koanf:"enabled" default:"true"`
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
	// Database provider configuration
	Database ProviderConfigs `json:"database" koanf:"database"`
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

// ProviderOption configures runtime provider options
type ProviderOption func(*ProviderOptions)

// ProviderOptions captures runtime configuration shared across storage providers
type ProviderOptions struct {
	Credentials ProviderCredentials
	Bucket      string
	Region      string
	Endpoint    string
	BasePath    string
	LocalURL    string
	extras      map[string]any
}

// NewProviderOptions constructs ProviderOptions applying the supplied options
func NewProviderOptions(opts ...ProviderOption) *ProviderOptions {
	po := &ProviderOptions{}
	po.Apply(opts...)
	return po
}

// Apply applies option functions to ProviderOptions
func (p *ProviderOptions) Apply(opts ...ProviderOption) {
	for _, opt := range opts {
		if opt != nil {
			opt(p)
		}
	}
}

// Clone returns a deep copy of ProviderOptions
func (p *ProviderOptions) Clone() *ProviderOptions {
	if p == nil {
		return nil
	}

	clone := *p

	if len(p.extras) > 0 {
		clone.extras = make(map[string]any, len(p.extras))
		maps.Copy(clone.extras, p.extras)
	}

	return &clone
}

// WithCredentials sets provider credentials
func WithCredentials(creds ProviderCredentials) ProviderOption {
	return func(p *ProviderOptions) {
		p.Credentials = creds
	}
}

// WithBucket sets the bucket/path value
func WithBucket(bucket string) ProviderOption {
	return func(p *ProviderOptions) {
		p.Bucket = bucket
	}
}

// WithRegion sets the region value
func WithRegion(region string) ProviderOption {
	return func(p *ProviderOptions) {
		p.Region = region
	}
}

// WithEndpoint sets the custom endpoint
func WithEndpoint(endpoint string) ProviderOption {
	return func(p *ProviderOptions) {
		p.Endpoint = endpoint
	}
}

// WithBasePath sets the local base path for disk providers
func WithBasePath(path string) ProviderOption {
	return func(p *ProviderOptions) {
		p.BasePath = path
	}
}

// WithLocalURL sets the local URL used for presigned links
func WithLocalURL(url string) ProviderOption {
	return func(p *ProviderOptions) {
		p.LocalURL = url
	}
}

// WithExtra attaches provider specific metadata
func WithExtra(key string, value any) ProviderOption {
	return func(p *ProviderOptions) {
		if p.extras == nil {
			p.extras = make(map[string]any)
		}
		p.extras[key] = value
	}
}

// Extra returns provider specific metadata
func (p *ProviderOptions) Extra(key string) (any, bool) {
	if p == nil || len(p.extras) == 0 {
		return nil, false
	}

	val, ok := p.extras[key]

	return val, ok
}
