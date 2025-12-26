package objects

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/pkg/logx"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/eddy"
	"github.com/theopenlane/iam/auth"
)

// ProviderCacheKey implements eddy.CacheKey for provider caching
type ProviderCacheKey struct {
	TenantID        string
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
}

// Config holds configuration for creating a new Service
type Config struct {
	Resolver       *eddy.Resolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	ClientService  *eddy.ClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	ValidationFunc storage.ValidationFunc
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
	}
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

// Download downloads a file using provider resolution
func (s *Service) Download(ctx context.Context, provider storage.Provider, file *storagetypes.File, opts *storage.DownloadOptions) (*storage.DownloadedMetadata, error) {
	if provider == nil {
		resolvedprovider, err := s.resolveDownloadProvider(ctx, file)
		if err != nil {
			return nil, err
		}
		provider = resolvedprovider
	}

	return s.objectService.Download(ctx, provider, file, opts)
}

// GetPresignedURL gets a presigned URL for a file using provider resolution
func (s *Service) GetPresignedURL(ctx context.Context, file *storagetypes.File, duration time.Duration) (string, error) {
	if file == nil {
		return "", ErrMissingFileID
	}

	provider, err := s.resolveDownloadProvider(ctx, file)
	if err != nil {
		return "", err
	}

	opts := &storagetypes.PresignedURLOptions{Duration: duration}
	return s.objectService.GetPresignedURL(ctx, provider, file, opts)
}

// Delete deletes a file using provider resolution
func (s *Service) Delete(ctx context.Context, file *storagetypes.File, opts *storagetypes.DeleteFileOptions) error {
	provider, err := s.resolveDownloadProvider(ctx, file)
	if err != nil {
		return err
	}

	return s.objectService.Delete(ctx, provider, file, opts)
}

// Exists checks if a file exists using provider resolution
func (s *Service) Exists(ctx context.Context, file *storagetypes.File) (bool, error) {
	provider, err := s.resolveDownloadProvider(ctx, file)
	if err != nil {
		return false, err
	}

	return provider.Exists(ctx, file)
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
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil || orgID == "" {
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
	orgID, _ := auth.GetOrganizationIDFromContext(ctx)

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
