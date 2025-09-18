package objects

import (
	"context"
	"io"
	"maps"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
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
		objectService: storage.NewObjectService().WithValidation(objects.ApplicationMimeTypeValidator),
	}
}

// SetObjectService allows overriding the internal object service (useful for dev mode)
func (s *Service) SetObjectService(objectService *storage.ObjectService) {
	s.objectService = objectService
}

// Upload uploads a file using provider resolution
func (s *Service) Upload(ctx context.Context, reader io.Reader, opts *storage.UploadOptions) (*storage.File, error) {
	// Ensure we have ProviderHints to capture resolution metadata
	if opts.ProviderHints == nil {
		opts.ProviderHints = &storage.ProviderHints{}
	}

	provider, err := s.resolveProvider(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Upload the file
	file, err := s.objectService.Upload(ctx, provider, reader, opts)
	if err != nil {
		return nil, err
	}

	// The resolveProvider method should have populated the ProviderHints with resolution metadata
	// Extract that information and add it to the returned file
	if hints, ok := opts.ProviderHints.(*storage.ProviderHints); ok {
		file.IntegrationID = hints.IntegrationID
		file.HushID = hints.HushID
		file.OrganizationID = hints.OrganizationID
		if hints.Metadata != nil {
			if file.Metadata == nil {
				file.Metadata = make(map[string]string)
			}

			maps.Copy(file.Metadata, hints.Metadata)
		}
	}

	return file, nil
}

// Download downloads a file using provider resolution
func (s *Service) Download(ctx context.Context, file *storage.File) (*storage.DownloadResult, error) {
	provider, err := s.resolveProviderForFile(ctx, file)
	if err != nil {
		return nil, err
	}

	return s.objectService.Download(ctx, provider, file)
}

// GetPresignedURL gets a presigned URL for a file using provider resolution
func (s *Service) GetPresignedURL(ctx context.Context, file *storage.File, duration time.Duration) (string, error) {
	provider, err := s.resolveProviderForFile(ctx, file)
	if err != nil {
		return "", err
	}

	return s.objectService.GetPresignedURL(ctx, provider, file, duration)
}

// Delete deletes a file using provider resolution
func (s *Service) Delete(ctx context.Context, file *storage.File) error {
	provider, err := s.resolveProviderForFile(ctx, file)
	if err != nil {
		return err
	}

	return s.objectService.Delete(ctx, provider, file)
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
func (s *Service) resolveProvider(ctx context.Context, opts *storage.UploadOptions) (storage.Provider, error) {
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

	// Extract integration IDs from provider hints or credentials
	integrationID := lo.ValueOr(res.Credentials, "integration_id", "")
	hushID := lo.ValueOr(res.Credentials, "hush_id", "")

	// Provider hints take precedence over credentials
	if hints, ok := opts.ProviderHints.(*storage.ProviderHints); ok {
		integrationID = lo.Ternary(hints.IntegrationID != "", hints.IntegrationID, integrationID)
		hushID = lo.Ternary(hints.HushID != "", hints.HushID, hushID)
	}

	// For system integrations, use system organization ID from credentials
	orgID = lo.ValueOr(res.Credentials, "system_org_id", orgID)

	if integrationID == "" || hushID == "" {
		return nil, ErrNoIntegrationOrCredentials
	}

	cacheKey := cp.ClientCacheKey{
		TenantID:        orgID,
		IntegrationType: string(res.ClientType),
		IntegrationID:   integrationID,
		HushID:          hushID,
	}

	client := s.clientService.GetClient(ctx, cacheKey, res.ClientType, res.Credentials, res.Config)
	if !client.IsPresent() {
		zerolog.Ctx(ctx).Error().Msg("failed to create provider from resolution")
		return nil, ErrProviderResolutionFailed
	}

	return client.MustGet(), nil
}

// resolveProviderForFile resolves a storage provider for file operations (download, delete, presigned URL)
func (s *Service) resolveProviderForFile(ctx context.Context, file *storage.File) (storage.Provider, error) {
	enrichedCtx := s.buildResolutionContextForFile(ctx, file)
	resolution := s.resolver.Resolve(enrichedCtx)

	providerType, hasResolution := resolution.Get()
	if !hasResolution {
		zerolog.Ctx(ctx).Error().Msgf("storage provider resolution failed for file %s", file.ID)
		return nil, ErrProviderResolutionFailed
	}

	// Build ClientCacheKey using file metadata with auth context as backup
	orgID, _ := auth.GetOrganizationIDFromContext(ctx)
	tenantID := lo.Ternary(file.OrganizationID != "", file.OrganizationID, orgID)

	if tenantID == "" {
		return nil, ErrNoOrganizationID
	}
	if file.IntegrationID == "" {
		return nil, ErrMissingIntegrationID
	}
	if file.HushID == "" {
		return nil, ErrMissingHushID
	}

	cacheKey := cp.ClientCacheKey{
		TenantID:        tenantID,
		IntegrationType: string(providerType.ClientType),
		IntegrationID:   file.IntegrationID,
		HushID:          file.HushID,
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
	if hints, ok := opts.ProviderHints.(*storage.ProviderHints); ok {
		ctx = cp.WithValue(ctx, hints)
	}

	return ctx
}

// buildResolutionContextForFile builds context for provider resolution from file metadata
func (s *Service) buildResolutionContextForFile(ctx context.Context, file *storage.File) context.Context {
	// Add organization and user information from auth context
	orgID, _ := auth.GetOrganizationIDFromContext(ctx)
	ctx = cp.WithValue(ctx, orgID)

	subjectID, _ := auth.GetSubjectIDFromContext(ctx)
	ctx = cp.WithValue(ctx, subjectID)

	// Add the entire file
	ctx = cp.WithValue(ctx, file)

	return ctx
}
