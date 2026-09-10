package objects

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/internal/consts"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/metrics"
	pkgobjects "github.com/theopenlane/core/v2/pkg/objects"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/eddy"
	"github.com/theopenlane/iam/auth"
)

// ProviderCacheKey implements eddy.CacheKey for provider caching
type ProviderCacheKey struct {
	// TenantID is the organization the cached provider client belongs to
	TenantID string
	// IntegrationType is the provider type the client was built for, e.g. s3
	IntegrationType string
}

// String returns the cache key as a string
func (k ProviderCacheKey) String() string {
	return fmt.Sprintf("%s:%s", k.TenantID, k.IntegrationType)
}

// Service orchestrates storage operations using eddy provider resolution
type Service struct {
	resolver      *eddy.Resolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	clientService *eddy.ClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	objectService *storage.ObjectService
	// backups holds backup destination targets keyed by source provider type
	backups map[storage.ProviderType]BackupTarget
}

// Config holds configuration for creating a new Service
type Config struct {
	// Resolver selects the storage provider to use for a given request from context hints
	Resolver *eddy.Resolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	// ClientService caches and returns provider clients built by the resolver
	ClientService *eddy.ClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	// ValidationFunc validates files before upload
	ValidationFunc storage.ValidationFunc
	// Backups maps a source provider type to its backup destination target
	Backups map[storage.ProviderType]BackupTarget
}

// NewService creates a new storage orchestration service
func NewService(cfg Config) *Service {
	objectService := storage.NewObjectService()

	// Configure validation if provided
	if cfg.ValidationFunc != nil {
		objectService = objectService.WithValidation(cfg.ValidationFunc)
	}

	return &Service{
		resolver:      cfg.Resolver,
		clientService: cfg.ClientService,
		objectService: objectService,
		backups:       cfg.Backups,
	}
}

// BackupProviderFor returns the configured backup destination provider for a source provider
// type, if one is configured. When absent, the source provider has no backup and behaves as today.
func (s *Service) BackupProviderFor(source storage.ProviderType) (storage.Provider, bool) {
	target, ok := s.backups[source]
	if !ok {
		return nil, false
	}

	return target.Provider, true
}

// ReadFromBackupFor returns the backup destination provider for a source provider when reads are
// configured to be served from the backup instead of the source
func (s *Service) ReadFromBackupFor(source storage.ProviderType) (storage.Provider, bool) {
	target, ok := s.backups[source]
	if !ok || !target.ReadFromBackup {
		return nil, false
	}

	return target.Provider, true
}

// operations a read can be served from a backup replica for, used as a metric label
const (
	backupOpDownload     = "download"
	backupOpPresignedURL = "presigned_url"
	backupOpExists       = "exists"
)

// replicaTarget returns the backup provider and a file pointing at the replica when the file has
// one recorded at a configured backup destination. It says nothing about whether reads should use
// it, so callers that only touch the replica, such as deleting it, can share this lookup
func (s *Service) replicaTarget(file *storagetypes.File) (storage.Provider, *storagetypes.File, bool) {
	if file == nil || file.BackupLocation == nil || file.BackupLocation.Key == "" {
		return nil, nil, false
	}

	target, ok := s.backups[file.ProviderType]
	if !ok {
		return nil, nil, false
	}

	return target.Provider, replicaFile(file), true
}

// replicaFile rewrites a file to point at its replica at the backup provider
func replicaFile(file *storagetypes.File) *storagetypes.File {
	location := file.BackupLocation

	replica := *file
	replica.ProviderType = location.ProviderType
	replica.BackupLocation = nil
	replica.Key = location.Key
	replica.Bucket = location.Bucket
	replica.Region = location.Region
	replica.FullURI = location.FullURI
	replica.ProviderHints = &storagetypes.ProviderHints{
		KnownProvider: location.ProviderType,
	}

	return &replica
}

// Upload uploads a file using provider resolution
func (s *Service) Upload(ctx context.Context, reader io.Reader, opts *storage.UploadOptions) (*pkgobjects.File, error) {
	provider, err := s.resolveUploadProvider(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Upload the file
	file, err := s.objectService.Upload(ctx, provider, reader, opts)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// readTarget returns the provider and file a read should be served from. A read goes to the
// backup replica when the file's source provider is configured to serve reads from it and the
// file has a replica recorded; that is a routing decision made before any I/O, not a retry, so
// the source provider is never contacted. Everything else resolves the source provider as normal
func (s *Service) readTarget(ctx context.Context, file *storagetypes.File, operation string) (storage.Provider, *storagetypes.File, error) {
	if file != nil {
		source := file.ProviderType

		_, readFromBackup := s.ReadFromBackupFor(source)

		// add logging fields to be reused in the context
		ctx = logx.WithFields(ctx, map[string]any{
			"file_id":          file.ID,
			"source":           string(source),
			"operation":        operation,
			"read_from_backup": readFromBackup,
			"has_replica":      file.BackupLocation != nil && file.BackupLocation.Key != ""})

		if len(s.backups) > 0 {
			logx.FromContext(ctx).Debug().Msg("backup read routing decision")
		}

		if readFromBackup {
			provider, replica, ok := s.replicaTarget(file)
			if !ok {
				// reads come from the backup but this file was never replicated, so the source
				// provider is the only place it can be served from
				metrics.RecordStorageBackupReadMissingReplica(string(source), operation)

				logx.FromContext(ctx).Warn().Msg("no backup replica for file; falling through to source provider")
			} else {
				metrics.RecordStorageBackupRead(string(source), string(replica.ProviderType), operation)

				logx.FromContext(ctx).Debug().Str("destination", string(replica.ProviderType)).Str("bucket", replica.Bucket).Str("key", replica.Key).Msg("serving file from backup replica")

				return provider, replica, nil
			}
		}
	}

	provider, err := s.resolveDownloadProvider(ctx, file)
	if err != nil {
		return nil, nil, err
	}

	return provider, file, nil
}

// Download downloads a file using provider resolution
func (s *Service) Download(ctx context.Context, provider storage.Provider, file *storagetypes.File, opts *storage.DownloadOptions) (*storage.DownloadedMetadata, error) {
	if provider == nil {
		target, targetFile, err := s.readTarget(ctx, file, backupOpDownload)
		if err != nil {
			return nil, err
		}

		return s.objectService.Download(ctx, target, targetFile, opts)
	}

	return s.objectService.Download(ctx, provider, file, opts)
}

// GetPresignedURL gets a presigned URL for a file using provider resolution
func (s *Service) GetPresignedURL(ctx context.Context, file *storagetypes.File, duration time.Duration) (string, error) {
	if file == nil {
		return "", ErrMissingFileID
	}

	opts := &storagetypes.PresignedURLOptions{Duration: duration}

	provider, target, err := s.readTarget(ctx, file, backupOpPresignedURL)
	if err != nil {
		return "", err
	}

	return s.objectService.GetPresignedURL(ctx, provider, target, opts)
}

// Delete deletes a file using provider resolution
func (s *Service) Delete(ctx context.Context, file *storagetypes.File, opts *storagetypes.DeleteFileOptions) error {
	provider, err := s.resolveDownloadProvider(ctx, file)
	if err != nil {
		return err
	}

	return s.objectService.Delete(ctx, provider, file, opts)
}

// DeleteBackup deletes the replica of a file from its backup provider so the replica shares the
// lifetime of the object it backs up. Files with no usable replica, or whose source provider has
// no backup configured, are a no-op
func (s *Service) DeleteBackup(ctx context.Context, file *storagetypes.File, opts *storagetypes.DeleteFileOptions) error {
	provider, replica, ok := s.replicaTarget(file)
	if !ok {
		return nil
	}

	err := s.objectService.Delete(ctx, provider, replica, opts)
	metrics.RecordStorageBackupDelete(string(file.ProviderType), string(replica.ProviderType), err)

	return err
}

// Exists checks if a file exists using provider resolution
func (s *Service) Exists(ctx context.Context, file *storagetypes.File) (bool, error) {
	provider, target, err := s.readTarget(ctx, file, backupOpExists)
	if err != nil {
		return false, err
	}

	return provider.Exists(ctx, target)
}

// Skipper returns the configured skipper function
func (s *Service) Skipper() storage.SkipperFunc {
	return s.objectService.Skipper()
}

// ErrorResponseHandler returns the configured error response handler
func (s *Service) ErrorResponseHandler() storage.ErrResponseHandler {
	return s.objectService.ErrorResponseHandler()
}

// MaxSize returns the configured maximum file size
func (s *Service) MaxSize() int64 {
	return s.objectService.MaxSize()
}

// Keys returns the configured form keys
func (s *Service) Keys() []string {
	return s.objectService.Keys()
}

// IgnoreNonExistentKeys returns whether to ignore non-existent form keys
func (s *Service) IgnoreNonExistentKeys() bool {
	return s.objectService.IgnoreNonExistentKeys()
}

// resolveProvider resolves a storage provider for upload operations
func (s *Service) resolveUploadProvider(ctx context.Context, opts *storage.UploadOptions) (storage.Provider, error) {
	enrichedCtx := s.buildResolutionContext(ctx, opts)
	result := s.resolver.Resolve(enrichedCtx)

	if !result.IsPresent() {
		logx.FromContext(ctx).Error().Msg("storage provider resolution failed: no provider resolved")
		return nil, ErrProviderResolutionFailed
	}

	res := result.MustGet()
	if res.Builder == nil {
		logx.FromContext(ctx).Error().Msg("storage provider resolution failed: resolved provider missing builder")
		return nil, ErrProviderResolutionFailed
	}

	// Get organization ID from auth context
	var orgID string
	if svcCaller, svcOk := auth.CallerFromContext(ctx); svcOk && svcCaller != nil {
		orgID = svcCaller.OrganizationID
		if orgID == "" && svcCaller.Has(auth.CapSystemAdmin) {
			orgID = consts.SystemAdminOrgID
		}
	}

	if orgID == "" {
		return nil, ErrNoOrganizationID
	}

	cacheKey := ProviderCacheKey{
		TenantID:        orgID,
		IntegrationType: res.Builder.ProviderType(),
	}

	client := s.clientService.GetClient(ctx, cacheKey, res.Builder, res.Output, res.Config)
	if !client.IsPresent() {
		logx.FromContext(ctx).Error().Str("integration_type", res.Builder.ProviderType()).Msg("storage provider resolution failed: provider client unavailable")
		return nil, ErrProviderResolutionFailed
	}

	return client.MustGet(), nil
}

// resolveProviderForFile resolves a storage provider for file operations (download, delete, presigned URL)
func (s *Service) resolveDownloadProvider(ctx context.Context, file *storagetypes.File) (storage.Provider, error) {
	enrichedCtx := s.buildResolutionContextForFile(ctx, file)
	result := s.resolver.Resolve(enrichedCtx)

	res, hasResult := result.Get()
	if !hasResult {
		logx.FromContext(ctx).Error().Msgf("storage provider resolution failed for file %s", file.ID)
		return nil, ErrProviderResolutionFailed
	}

	// Build ProviderCacheKey using file metadata with auth context as backup
	var orgID string
	if dlCaller, dlOk := auth.CallerFromContext(ctx); dlOk && dlCaller != nil {
		orgID = dlCaller.OrganizationID
	}

	cacheKey := ProviderCacheKey{
		TenantID:        orgID,
		IntegrationType: res.Builder.ProviderType(),
	}

	return s.clientService.GetClient(ctx, cacheKey, res.Builder, res.Output, res.Config).
		OrElse(nil), nil
}

// buildResolutionContext builds context for provider resolution from upload options
func (s *Service) buildResolutionContext(ctx context.Context, opts *storage.UploadOptions) context.Context {
	// Add provider hints if present
	if opts.ProviderHints != nil {
		ctx = ApplyProviderHints(ctx, opts.ProviderHints)
	}

	return ctx
}

// buildResolutionContextForFile builds context for provider resolution from file metadata
func (s *Service) buildResolutionContextForFile(ctx context.Context, file *storagetypes.File) context.Context {
	// Add provider hints from file
	ctx = ApplyProviderHints(ctx, file.ProviderHints)

	return ctx
}
