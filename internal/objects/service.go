package objects

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
	"github.com/theopenlane/eddy"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/ulids"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	defaultPresignedURLDuration = 10 * time.Minute
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
	resolver        *eddy.Resolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	clientService   *eddy.ClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	objectService   *storage.ObjectService
	tokenProvider   func() *tokens.TokenManager
	tokenIssuer     string
	tokenAudience   string
	downloadSecrets sync.Map
}

// Config holds configuration for creating a new Service
type Config struct {
	Resolver       *eddy.Resolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	ClientService  *eddy.ClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	ValidationFunc storage.ValidationFunc
	TokenManager   func() *tokens.TokenManager
	TokenIssuer    string
	TokenAudience  string
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
		tokenProvider: cfg.TokenManager,
		tokenIssuer:   cfg.TokenIssuer,
		tokenAudience: cfg.TokenAudience,
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
	if s.tokenProvider == nil {
		provider, err := s.resolveDownloadProvider(ctx, file)
		if err != nil {
			return "", err
		}

		opts := &storagetypes.PresignedURLOptions{Duration: duration}
		return s.objectService.GetPresignedURL(ctx, provider, file, opts)
	}

	if file == nil || file.ID == "" {
		return "", ErrMissingFileID
	}

	// ensure provider resolution succeeds before issuing token
	if _, err := s.resolveDownloadProvider(ctx, file); err != nil {
		return "", err
	}

	if duration <= 0 {
		duration = defaultPresignedURLDuration
	}

	objectURI := buildDownloadObjectURI(file.FileMetadata.ProviderType, file.FileMetadata.Bucket, file.FileMetadata.Key) // nolint:staticcheck
	options := []tokens.DownloadTokenOption{
		tokens.WithDownloadTokenExpiresIn(duration),
		tokens.WithDownloadTokenContentType(file.FileMetadata.ContentType), // nolint:staticcheck
	}

	if file.OriginalName != "" {
		options = append(options, tokens.WithDownloadTokenFileName(file.OriginalName))
	}

	if authUser, ok := auth.AuthenticatedUserFromContext(ctx); ok && authUser != nil {
		if userID, err := ulid.Parse(authUser.SubjectID); err == nil {
			options = append(options, tokens.WithDownloadTokenUserID(userID))
		}
		if authUser.OrganizationID != "" {
			if orgID, err := ulid.Parse(authUser.OrganizationID); err == nil {
				options = append(options, tokens.WithDownloadTokenOrgID(orgID))
			}
		}
	}

	downloadToken, err := tokens.NewDownloadToken(objectURI, options...)
	if err != nil {
		return "", err
	}

	signature, secret, err := downloadToken.Sign()
	if err != nil {
		return "", err
	}

	s.storeDownloadSecret(downloadToken.TokenID, secret, downloadToken.ExpiresAt)

	payload, err := msgpack.Marshal(downloadToken)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	combined := fmt.Sprintf("%s.%s", signature, encodedPayload)
	escaped := url.QueryEscape(combined)

	return fmt.Sprintf("/v1/files/%s/download?token=%s", url.PathEscape(file.ID), escaped), nil
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

type downloadSecret struct {
	secret    []byte
	expiresAt time.Time
}

func (s *Service) storeDownloadSecret(tokenID ulid.ULID, secret []byte, expiresAt time.Time) {
	if ulids.IsZero(tokenID) || len(secret) == 0 {
		return
	}

	copySecret := make([]byte, len(secret))
	copy(copySecret, secret)

	key := tokenID.String()
	s.downloadSecrets.Store(key, downloadSecret{secret: copySecret, expiresAt: expiresAt})

	if ttl := time.Until(expiresAt); ttl > 0 {
		time.AfterFunc(ttl, func() {
			s.downloadSecrets.Delete(key)
		})
	}
}

func (s *Service) LookupDownloadSecret(tokenID ulid.ULID) ([]byte, bool) {
	if ulids.IsZero(tokenID) {
		return nil, false
	}

	value, ok := s.downloadSecrets.Load(tokenID.String())
	if !ok {
		return nil, false
	}

	ds := value.(downloadSecret)
	if time.Now().After(ds.expiresAt) {
		s.downloadSecrets.Delete(tokenID.String())
		return nil, false
	}

	return ds.secret, true
}

func buildDownloadObjectURI(provider storagetypes.ProviderType, bucket, key string) string {
	return fmt.Sprintf("%s:%s:%s", string(provider), bucket, key)
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
		zerolog.Ctx(ctx).Error().Msg("storage provider resolution failed: no provider resolved")
		return nil, ErrProviderResolutionFailed
	}

	res := result.MustGet()
	if res.Builder == nil {
		zerolog.Ctx(ctx).Error().Msg("storage provider resolution failed: resolved provider missing builder")
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
		zerolog.Ctx(ctx).Error().Str("integration_type", res.Builder.ProviderType()).Msg("storage provider resolution failed: provider client unavailable")
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
		zerolog.Ctx(ctx).Error().Msgf("storage provider resolution failed for file %s", file.ID)
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
