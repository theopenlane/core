package objects

import (
	"context"
	"io"

	"github.com/rs/zerolog"
	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
	"github.com/theopenlane/iam/auth"
)

// Service orchestrates storage operations using cp provider resolution
type Service struct {
	resolver      *cp.Resolver[storage.Provider]
	clientService *cp.ClientService[storage.Provider]
	objectService *storage.ObjectService
}

// NewService creates a new storage orchestration service
func NewService(resolver *cp.Resolver[storage.Provider], clientService *cp.ClientService[storage.Provider]) *Service {
	return &Service{
		resolver:      resolver,
		clientService: clientService,
		objectService: storage.NewObjectService(),
	}
}

// SetObjectService allows overriding the internal object service (useful for dev mode)
func (s *Service) SetObjectService(objectService *storage.ObjectService) {
	s.objectService = objectService
}

// Upload uploads a file using provider resolution
func (s *Service) Upload(ctx context.Context, reader io.Reader, opts *storage.UploadOptions) (*storage.File, error) {
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
//func (s *Service) GetPresignedURL(ctx context.Context, file *storage.File, duration time.Duration) (string, error) {
//	provider, err := s.resolveUploadProvider(ctx, file)
//	if err != nil {
//		return "", err
//	}
//
//	return s.objectService.GetPresignedURL(ctx, provider, file, duration)
//}

// Delete deletes a file using provider resolution
func (s *Service) Delete(ctx context.Context, file *storagetypes.File, opts *storagetypes.DeleteFileOptions) error {
	provider, err := s.resolveDownloadProvider(ctx, file)
	if err != nil {
		return err
	}

	return s.objectService.Delete(ctx, provider, file, opts)
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
	resolution := s.resolver.Resolve(enrichedCtx)

	res := resolution.OrEmpty()
	if res.ClientType == "" {
		zerolog.Ctx(ctx).Error().Msg("storage provider resolution failed")
		return nil, ErrProviderResolutionFailed
	}

	// Get organization ID from auth context
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return nil, ErrNoOrganizationID
	}

	cacheKey := cp.ClientCacheKey{
		TenantID:        orgID,
		IntegrationType: string(res.ClientType),
		//		IntegrationID:   integrationID,
		//		HushID:          hushID,
	}

	client := s.clientService.GetClient(ctx, cacheKey, res.ClientType, res.Credentials, res.Config)
	if !client.IsPresent() {
		zerolog.Ctx(ctx).Error().Msg("failed to create provider from resolution")
		return nil, ErrProviderResolutionFailed
	}

	return client.MustGet(), nil
}

// resolveProviderForFile resolves a storage provider for file operations (download, delete, presigned URL)
func (s *Service) resolveDownloadProvider(ctx context.Context, file *storagetypes.File) (storage.Provider, error) {
	enrichedCtx := s.buildResolutionContextForFile(ctx, file)
	resolution := s.resolver.Resolve(enrichedCtx)

	providerType, hasResolution := resolution.Get()
	if !hasResolution {
		zerolog.Ctx(ctx).Error().Msgf("storage provider resolution failed for file %s", file.ID)
		return nil, ErrProviderResolutionFailed
	}

	// Build ClientCacheKey using file metadata with auth context as backup
	orgID, _ := auth.GetOrganizationIDFromContext(ctx)

	cacheKey := cp.ClientCacheKey{
		TenantID:        orgID,
		IntegrationType: string(providerType.ClientType),
		//		IntegrationID:   file.IntegrationID,
		//		HushID:          file.HushID,
	}

	return s.clientService.GetClient(ctx, cacheKey, providerType.ClientType, providerType.Credentials, providerType.Config).
		OrElse(nil), nil
}

// buildResolutionContext builds context for provider resolution from upload options
func (s *Service) buildResolutionContext(ctx context.Context, opts *storage.UploadOptions) context.Context {
	// Add organization and user information from auth context
	orgID, _ := auth.GetOrganizationIDFromContext(ctx)
	ctx = cp.WithValue(ctx, orgID)

	subjectID, _ := auth.GetSubjectIDFromContext(ctx)
	ctx = cp.WithValue(ctx, subjectID)

	// Add upload options
	ctx = cp.WithValue(ctx, opts)

	// Add provider hints if present
	if opts.ProviderHints != nil {
		ctx = cp.WithValue(ctx, opts.ProviderHints)
	}

	return ctx
}

// buildResolutionContextForFile builds context for provider resolution from file metadata
func (s *Service) buildResolutionContextForFile(ctx context.Context, file *storagetypes.File) context.Context {
	// Add organization and user information from auth context
	orgID, _ := auth.GetOrganizationIDFromContext(ctx)
	ctx = cp.WithValue(ctx, orgID)

	subjectID, _ := auth.GetSubjectIDFromContext(ctx)
	ctx = cp.WithValue(ctx, subjectID)

	// Add the entire file
	ctx = cp.WithValue(ctx, file)

	return ctx
}
