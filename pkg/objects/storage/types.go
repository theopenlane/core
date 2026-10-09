package storage

import (
	"context"
	"maps"
	"net/http"

	"github.com/theopenlane/core/common/storagetypes"

	"github.com/theopenlane/iam/tokens"
)

// Alias types from common/storagetypes to maintain clean imports
// having a bunch of smaller subpackages seemed to just complicate things
type (
	Provider           = storagetypes.Provider
	ProviderType       = storagetypes.ProviderType
	PresignMode        = storagetypes.PresignMode
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
	DiskProvider     = storagetypes.DiskProvider
	DatabaseProvider = storagetypes.DatabaseProvider
	GCSProvider      = storagetypes.GCSProvider
	// Presign mode constants
	PresignModeProvider = storagetypes.PresignModeProvider
	PresignModeProxy    = storagetypes.PresignModeProxy
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
	Keys []string `json:"keys" koanf:"keys" default:"[uploadFile]"`
	// MaxSizeMB is the maximum file size allowed in MB
	MaxSizeMB int64 `json:"maxsizemb" koanf:"maxsizemb"`
	// MaxMemoryMB is the maximum memory to use for file uploads in MB
	MaxMemoryMB int64 `json:"maxmemorymb" koanf:"maxmemorymb"`
	// DevMode automatically configures a local disk storage provider (and ensures directories exist) and ignores other provider configs
	DevMode bool `json:"devmode" koanf:"devmode" default:"false"`
	// Providers contains configuration for each storage provider
	Providers Providers `json:"providers" koanf:"providers"`
}

// Providers contains the configuration for each storage provider
type Providers struct {
	// S3 provider configuration
	S3 S3Config `json:"s3" koanf:"s3"`
	// R2 provider configuration
	R2 R2Config `json:"r2" koanf:"r2"`
	// GCS provider configuration
	GCS GCSConfig `json:"gcs" koanf:"gcs"`
	// Disk provider configuration
	Disk DiskConfig `json:"disk" koanf:"disk"`
	// Database provider configuration
	Database DatabaseConfig `json:"database" koanf:"database"`
}

// ProviderCommon holds the settings every storage provider shares
type ProviderCommon struct {
	// Enabled indicates if this provider is enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// EnsureAvailable enforces provider availability before completing server startup
	EnsureAvailable bool `json:"ensureavailable" koanf:"ensureavailable" default:"false"`
	// Bucket is the bucket name, or the directory path for the disk provider
	Bucket string `json:"bucket" koanf:"bucket"`
	// Backup optionally replicates this provider's objects to another provider asynchronously
	Backup *BackupConfig `json:"backup" koanf:"backup"`
}

// ProxyPresign holds the proxy-signed download URL settings
type ProxyPresign struct {
	// ProxyPresignEnabled toggles proxy-signed download URL generation
	ProxyPresignEnabled bool `json:"proxypresignenabled" koanf:"proxypresignenabled" default:"false"`
	// BaseURL is the prefix for proxy download URLs (e.g., http://localhost:17608/v1/files)
	BaseURL string `json:"baseurl" koanf:"baseurl" default:"http://localhost:17608/v1/files"`
}

// S3Config configures the Amazon S3 provider
type S3Config struct {
	ProviderCommon `koanf:",squash"`
	ProxyPresign   `koanf:",squash"`
	// Region the bucket lives in
	Region string `json:"region" koanf:"region"`
	// Endpoint overrides the AWS endpoint for S3 compatible services
	Endpoint string `json:"endpoint" koanf:"endpoint"`
	// Credentials contains the access key pair for the bucket
	Credentials AccessKeyCredentials `json:"credentials" koanf:"credentials"`
}

// R2Config configures the Cloudflare R2 provider
type R2Config struct {
	ProviderCommon `koanf:",squash"`
	ProxyPresign   `koanf:",squash"`
	// Endpoint overrides the account endpoint derived from the account ID
	Endpoint string `json:"endpoint" koanf:"endpoint"`
	// Credentials contains the access key pair and account for the bucket
	Credentials R2Credentials `json:"credentials" koanf:"credentials"`
}

// GCSConfig configures the Google Cloud Storage provider
type GCSConfig struct {
	ProviderCommon `koanf:",squash"`
	ProxyPresign   `koanf:",squash"`
	// Endpoint overrides the Google API endpoint, for an emulator
	Endpoint string `json:"endpoint" koanf:"endpoint"`
	// ProjectID is the Google Cloud project that owns the bucket
	ProjectID string `json:"projectid" koanf:"projectid"`
}

// DiskConfig configures the local filesystem provider
type DiskConfig struct {
	ProviderCommon `koanf:",squash"`
	ProxyPresign   `koanf:",squash"`
	// Endpoint is the URL files are served from when proxy presigning is disabled
	Endpoint string `json:"endpoint" koanf:"endpoint"`
}

// DatabaseConfig configures the provider that stores file bytes in the database
type DatabaseConfig struct {
	ProviderCommon `koanf:",squash"`
	// BaseURL is the prefix for proxy download URLs (e.g., http://localhost:17608/v1/files)
	BaseURL string `json:"baseurl" koanf:"baseurl" default:"http://localhost:17608/v1/files"`
}

// BackupBucketSuffix is appended to the destination provider's bucket when writing backups, so
// replicated objects never share a bucket with the destination's live objects
const BackupBucketSuffix = "-backup"

// BackupConfig defines an asynchronous replication target for a provider's objects. It only
// names whether backups run and where they go; the destination provider's own configuration
// supplies the region, endpoint, and credentials
type BackupConfig struct {
	// Enabled indicates if this backup target is enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// Provider names the destination backend type, e.g. s3; empty replicates to the source
	// provider itself, which writes the backup to the suffixed bucket alongside the live one
	Provider ProviderType `json:"provider" koanf:"provider"`
	// ReadFromBackup serves reads from this backup target instead of the source provider, intended
	// to be enabled during a disaster recovery event when the source provider storage is lost
	ReadFromBackup bool `json:"readfrombackup" koanf:"readfrombackup" default:"false"`
	// Region optionally overrides the destination provider's region, so a backup can replicate
	// into a region other than the one holding the live objects
	Region string `json:"region" koanf:"region"`
}

// ByType returns the shared settings of each provider keyed by its provider type
func (p Providers) ByType() map[ProviderType]ProviderCommon {
	return map[ProviderType]ProviderCommon{
		S3Provider:       p.S3.ProviderCommon,
		R2Provider:       p.R2.ProviderCommon,
		GCSProvider:      p.GCS.ProviderCommon,
		DiskProvider:     p.Disk.ProviderCommon,
		DatabaseProvider: p.Database.ProviderCommon,
	}
}

// BackupDestination returns the provider type this provider's backups are written to and whether
// an enabled backup is configured for it. A backup with no provider named replicates to the
// source itself, which writes to the suffixed bucket alongside the live one
func (c ProviderCommon) BackupDestination(source ProviderType) (ProviderType, bool) {
	if !c.Enabled || c.Backup == nil || !c.Backup.Enabled {
		return "", false
	}

	if c.Backup.Provider == "" {
		return source, true
	}

	return c.Backup.Provider, true
}

// BackupBucket returns the bucket backups are written to for the given destination bucket
func BackupBucket(bucket string) string {
	if bucket == "" {
		return ""
	}

	return bucket + BackupBucketSuffix
}

// AccessKeyCredentials is the access key pair used by S3 compatible providers
type AccessKeyCredentials struct {
	// AccessKeyID for the bucket
	AccessKeyID string `json:"accesskeyid" koanf:"accesskeyid" sensitive:"true"`
	// SecretAccessKey for the bucket
	SecretAccessKey string `json:"secretaccesskey" koanf:"secretaccesskey" sensitive:"true"`
}

// R2Credentials is the access key pair plus the Cloudflare account that owns the bucket
type R2Credentials struct {
	AccessKeyCredentials `koanf:",squash"`
	// AccountID for Cloudflare R2
	AccountID string `json:"accountid" koanf:"accountid" sensitive:"true"`
}

// ProviderCredentials is the runtime credential set handed to provider builders
type ProviderCredentials struct {
	// AccessKeyID for cloud providers
	AccessKeyID string `json:"accesskeyid" koanf:"accesskeyid" sensitive:"true"`
	// SecretAccessKey for cloud providers
	SecretAccessKey string `json:"secretaccesskey" koanf:"secretaccesskey" sensitive:"true"`
	// AccountID for Cloudflare R2
	AccountID string `json:"accountid" koanf:"accountid" sensitive:"true"`
}

// ProviderCredentials returns the runtime credentials for the access key pair
func (c AccessKeyCredentials) ProviderCredentials() ProviderCredentials {
	return ProviderCredentials{AccessKeyID: c.AccessKeyID, SecretAccessKey: c.SecretAccessKey}
}

// ProviderCredentials returns the runtime credentials for the access key pair and account
func (c R2Credentials) ProviderCredentials() ProviderCredentials {
	return ProviderCredentials{AccessKeyID: c.AccessKeyID, SecretAccessKey: c.SecretAccessKey, AccountID: c.AccountID}
}

// ProviderOption configures runtime provider options
type ProviderOption func(*ProviderOptions)

// ProviderOptions captures runtime configuration shared across storage providers
type ProviderOptions struct {
	Credentials         ProviderCredentials
	Bucket              string
	Region              string
	Endpoint            string
	LocalURL            string
	ProxyPresignEnabled bool
	ProxyPresignConfig  *ProxyPresignConfig
	extras              map[string]any
}

// ProxyPresignConfig carries runtime dependencies for proxy download URL generation.
type ProxyPresignConfig struct {
	TokenManager  *tokens.TokenManager
	TokenIssuer   string
	TokenAudience string
	BaseURL       string
}

// ProxyPresignOption configures a ProxyPresignConfig.
type ProxyPresignOption func(*ProxyPresignConfig)

// NewProxyPresignConfig builds a ProxyPresignConfig applying the supplied options.
func NewProxyPresignConfig(opts ...ProxyPresignOption) *ProxyPresignConfig {
	return ApplyProxyPresignOptions(nil, opts...)
}

// Apply applies the supplied options to the existing ProxyPresignConfig.
func (p *ProxyPresignConfig) Apply(opts ...ProxyPresignOption) *ProxyPresignConfig {
	if p == nil {
		return ApplyProxyPresignOptions(nil, opts...)
	}

	return ApplyProxyPresignOptions(p, opts...)
}

// ApplyProxyPresignOptions applies options to the provided config, allocating one if needed.
func ApplyProxyPresignOptions(cfg *ProxyPresignConfig, opts ...ProxyPresignOption) *ProxyPresignConfig {
	if cfg == nil {
		cfg = &ProxyPresignConfig{}
	}

	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	return cfg
}

// WithProxyPresignTokenManager sets the token manager when provided.
func WithProxyPresignTokenManager(tm *tokens.TokenManager) ProxyPresignOption {
	return func(cfg *ProxyPresignConfig) {
		if tm != nil {
			cfg.TokenManager = tm
		}
	}
}

// WithProxyPresignTokenIssuer sets the token issuer when provided.
func WithProxyPresignTokenIssuer(issuer string) ProxyPresignOption {
	return func(cfg *ProxyPresignConfig) {
		if issuer != "" {
			cfg.TokenIssuer = issuer
		}
	}
}

// WithProxyPresignTokenAudience sets the token audience when provided.
func WithProxyPresignTokenAudience(audience string) ProxyPresignOption {
	return func(cfg *ProxyPresignConfig) {
		if audience != "" {
			cfg.TokenAudience = audience
		}
	}
}

// WithProxyPresignBaseURL sets the base URL for generated download links.
func WithProxyPresignBaseURL(baseURL string) ProxyPresignOption {
	return func(cfg *ProxyPresignConfig) {
		if baseURL != "" {
			cfg.BaseURL = baseURL
		}
	}
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

	if p.ProxyPresignConfig != nil {
		cfg := *p.ProxyPresignConfig
		clone.ProxyPresignConfig = &cfg
	}

	if len(p.extras) > 0 {
		clone.extras = make(map[string]any, len(p.extras))
		maps.Copy(clone.extras, p.extras)
	}

	return &clone
}

// WithProxyPresignEnabled toggles proxy URL generation.
func WithProxyPresignEnabled(enabled bool) ProviderOption {
	return func(p *ProviderOptions) {
		p.ProxyPresignEnabled = enabled
	}
}

// WithProxyPresignConfig sets proxy presign runtime dependencies.
func WithProxyPresignConfig(cfg *ProxyPresignConfig) ProviderOption {
	return func(p *ProviderOptions) {
		p.ProxyPresignConfig = cfg
	}
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

// WithLocalURL sets the local URL used for presigned links
func WithLocalURL(url string) ProviderOption {
	return func(p *ProviderOptions) {
		p.LocalURL = url
	}
}

// BackupTargetExtraKey marks provider options that were built from a backup destination
// configuration, so a caller can tell a resolved backup apart from a live provider
const BackupTargetExtraKey = "backup_target"

// GCSProjectIDExtraKey carries the Google Cloud project that owns the GCS bucket, read by the gcs provider when listing buckets
const GCSProjectIDExtraKey = "gcs_project_id"

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
