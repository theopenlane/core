package objects

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
)

// mockProvider implements storagetypes.Provider for testing
type mockProvider struct {
	name string
}

// ProviderType implements storagetypes.Provider.
func (m *mockProvider) ProviderType() storagetypes.ProviderType {
	panic("unimplemented")
}

func (m *mockProvider) Upload(ctx context.Context, reader io.Reader, opts *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
	return &storagetypes.UploadedFileMetadata{
		FileMetadata: storagetypes.FileMetadata{
			Key:          "test-file-key",
			Size:         100,
			ContentType:  opts.ContentType,
			ProviderType: storagetypes.ProviderType(m.name),
		},
		ProviderHints: &storagetypes.ProviderHints{
			IntegrationID:  "hint-integration",
			HushID:         "hint-hush",
			OrganizationID: "hint-org",
		},
	}, nil
}

func (m *mockProvider) Download(ctx context.Context, opts *storagetypes.DownloadFileOptions) (*storagetypes.DownloadedFileMetadata, error) {
	// Return a fixed size download to match test expectations
	content := make([]byte, 100) // 100 bytes to match test
	for i := range content {
		content[i] = byte('A' + (i % 26)) // Fill with letters
	}
	return &storagetypes.DownloadedFileMetadata{
		File: content,
		Size: 100,
	}, nil
}

func (m *mockProvider) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockProvider) GetPresignedURL(key string, expires time.Duration) (string, error) {
	return fmt.Sprintf("https://%s.example.com/presigned/%s", m.name, key), nil
}

func (m *mockProvider) Exists(ctx context.Context, key string) (bool, error) {
	return true, nil
}

func (m *mockProvider) GetScheme() *string {
	scheme := fmt.Sprintf("%s://", m.name)
	return &scheme
}

func (m *mockProvider) ListBuckets() ([]string, error) {
	// Return a mock list of buckets for testing
	return []string{"bucket1", "bucket2"}, nil
}

func (m *mockProvider) Close() error {
	return nil
}

// mockBuilder implements cp.ClientBuilder for testing
type mockBuilder struct {
	provider *mockProvider
}

func (m *mockBuilder) WithCredentials(credentials map[string]string) cp.ClientBuilder[storage.Provider] {
	return m
}

func (m *mockBuilder) WithConfig(config map[string]any) cp.ClientBuilder[storage.Provider] {
	return m
}

func (m *mockBuilder) Build(ctx context.Context) (storage.Provider, error) {
	return m.provider, nil
}

func (m *mockBuilder) ClientType() cp.ProviderType {
	return cp.ProviderType(m.provider.name)
}

// createTestService creates a test service with typed context resolution rules
func createTestService() *Service {
	// Create resolver with typed context rules
	resolver := cp.NewResolver[storage.Provider]()

	// Trust Center Module → R2 Provider
	trustCenterRule := cp.NewRule[storage.Provider]().
		WhenFunc(func(ctx context.Context) bool {
			return cp.GetValueEquals(ctx, models.CatalogTrustCenterModule)
		}).
		Resolve(func(ctx context.Context) (*cp.ResolvedProvider, error) {
			return &cp.ResolvedProvider{
				Type: "r2",
				Credentials: map[string]string{
					"integration_id": "r2-integration",
					"hush_id":        "r2-hush",
					"system_org_id":  ulids.New().String(), // Generate valid ULID
				},
				Config: map[string]any{
					"region": "us-east-1",
				},
			}, nil
		})

	// Compliance Module → S3 Provider
	complianceRule := cp.NewRule[storage.Provider]().
		WhenFunc(func(ctx context.Context) bool {
			return cp.GetValueEquals(ctx, models.CatalogComplianceModule)
		}).
		Resolve(func(ctx context.Context) (*cp.ResolvedProvider, error) {
			return &cp.ResolvedProvider{
				Type: "s3",
				Credentials: map[string]string{
					"integration_id": "s3-integration",
					"hush_id":        "s3-hush",
					"system_org_id":  ulids.New().String(), // Generate valid ULID
				},
				Config: map[string]any{
					"region": "us-west-2",
				},
			}, nil
		})

	// Default rule for fallback
	defaultRule := cp.DefaultRule[storage.Provider](cp.Resolution{
		ClientType: "disk",
		Credentials: map[string]string{
			"integration_id": "disk-integration",
			"hush_id":        "disk-hush",
			"system_org_id":  ulids.New().String(), // Generate valid ULID
		},
		Config: map[string]any{
			"base_path": "/tmp",
		},
	})

	resolver.AddRule(trustCenterRule)
	resolver.AddRule(complianceRule)
	resolver.AddRule(defaultRule)

	// Create client service with mock builders
	pool := cp.NewClientPool[storage.Provider](5 * time.Minute)
	clientService := cp.NewClientService(pool)

	// Register mock builders
	clientService.RegisterBuilder("r2", &mockBuilder{provider: &mockProvider{name: "r2"}})
	clientService.RegisterBuilder("s3", &mockBuilder{provider: &mockProvider{name: "s3"}})
	clientService.RegisterBuilder("disk", &mockBuilder{provider: &mockProvider{name: "disk"}})

	return NewService(resolver, clientService)
}

func TestService_Upload_TypedContextResolution(t *testing.T) {
	service := createTestService()

	tests := []struct {
		name             string
		setupContext     func() context.Context
		expectedProvider string
	}{
		{
			name: "TrustCenter module resolves to R2",
			setupContext: func() context.Context {
				// Create authenticated user context with organization ID
				ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())
				// Add module information for provider resolution
				ctx = cp.WithValue(ctx, models.CatalogTrustCenterModule)
				return ctx
			},
			expectedProvider: "r2",
		},
		{
			name: "Compliance module resolves to S3",
			setupContext: func() context.Context {
				ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())
				ctx = cp.WithValue(ctx, models.CatalogComplianceModule)
				return ctx
			},
			expectedProvider: "s3",
		},
		{
			name: "Unknown module falls back to disk",
			setupContext: func() context.Context {
				ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())
				ctx = cp.WithValue(ctx, "unknown-module")
				return ctx
			},
			expectedProvider: "disk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			reader := strings.NewReader("test content")
			opts := &storage.UploadOptions{
				FileName:      "test-file.txt",
				ContentType:   "text/plain",
				ProviderHints: &storagetypes.ProviderHints{},
			}

			file, err := service.Upload(ctx, reader, opts)
			require.NoError(t, err)
			assert.Equal(t, "test-file.txt", file.Name)
			assert.Equal(t, tt.expectedProvider, string(file.ProviderType))
		})
	}
}

func TestService_Download_TypedContextResolution(t *testing.T) {
	service := createTestService()

	// Create a test file with metadata
	testFile := &storage.File{
		ID:           "test-file-id",
		OriginalName: "test-file.txt",
		FileMetadata: storagetypes.FileMetadata{
			Key:         "test-file-id", // Use same value as ID for testing
			Size:        100,
			ContentType: "text/plain",
		},
	}

	ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())

	result, err := service.Download(ctx, testFile)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify the download result
	assert.Equal(t, int64(100), result.Size)
	assert.NotNil(t, result.File)
}

func TestService_GetPresignedURL_TypedContextResolution(t *testing.T) {
	service := createTestService()

	testFile := &storage.File{
		ID:           "test-file-id",
		OriginalName: "test-file.txt",
		FileMetadata: storagetypes.FileMetadata{
			Key:            "test-file-id", // Use same value as ID for testing
			Size:           100,
			ContentType:    "text/plain",
			IntegrationID:  "test-integration",
			HushID:         "test-hush",
			OrganizationID: "test-org",
		},
	}

	ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())

	url, err := service.GetPresignedURL(ctx, testFile, time.Hour)
	require.NoError(t, err)
	assert.Contains(t, url, "presigned")
	assert.Contains(t, url, testFile.ID)
}

func TestService_Delete_TypedContextResolution(t *testing.T) {
	service := createTestService()

	testFile := &storage.File{
		ID:   "test-file-id",
		Name: "test-file.txt",
		FileMetadata: storagetypes.FileMetadata{
			Key:            "test-file-id", // Use same value as ID for testing
			Size:           100,
			ContentType:    "text/plain",
			IntegrationID:  "test-integration",
			HushID:         "test-hush",
			OrganizationID: "test-org",
		},
	}

	ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())

	err := service.Delete(ctx, testFile)
	require.NoError(t, err)
}

func TestService_BuildResolutionContext(t *testing.T) {
	service := createTestService()

	ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())

	opts := &storage.UploadOptions{
		FileName:    "test-file.txt",
		ContentType: "text/plain",
		ProviderHints: &storagetypes.ProviderHints{
			IntegrationID:  "hint-integration",
			HushID:         "hint-hush",
			OrganizationID: "hint-org",
		},
	}

	enrichedCtx := service.buildResolutionContext(ctx, opts)

	// Verify context contains expected values
	uploadOpts := cp.GetValue[*storage.UploadOptions](enrichedCtx)
	assert.True(t, uploadOpts.IsPresent())
	assert.Equal(t, opts, uploadOpts.MustGet())

	hints := cp.GetValue[*storagetypes.ProviderHints](enrichedCtx)
	assert.True(t, hints.IsPresent())
	assert.Equal(t, "hint-integration", hints.MustGet().IntegrationID)
}

func TestService_BuildResolutionContextForFile(t *testing.T) {
	service := createTestService()

	ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())

	file := &storage.File{
		ID:   "test-file-id",
		Name: "test-file.txt",
		FileMetadata: storagetypes.FileMetadata{
			IntegrationID:  "file-integration",
			HushID:         "file-hush",
			OrganizationID: "file-org",
		},
	}

	enrichedCtx := service.buildResolutionContextForFile(ctx, file)

	// Verify context contains expected values
	contextFile := cp.GetValue[*storage.File](enrichedCtx)
	assert.True(t, contextFile.IsPresent())
	assert.Equal(t, file, contextFile.MustGet())
}

func TestService_ResolveProvider_ErrorHandling(t *testing.T) {
	// Create resolver with no rules (should fail resolution)
	resolver := cp.NewResolver[storage.Provider]()
	pool := cp.NewClientPool[storage.Provider](5 * time.Minute)
	clientService := cp.NewClientService(pool)
	service := NewService(resolver, clientService)

	ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())

	opts := &storage.UploadOptions{
		FileName:      "test-file.txt",
		ContentType:   "text/plain",
		ProviderHints: &storagetypes.ProviderHints{},
	}

	reader := strings.NewReader("test content")
	_, err := service.Upload(ctx, reader, opts)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrProviderResolutionFailed)
}

func TestService_ResolveProviderForFile_ErrorHandling(t *testing.T) {
	// Create resolver with no rules
	resolver := cp.NewResolver[storage.Provider]()
	pool := cp.NewClientPool[storage.Provider](5 * time.Minute)
	clientService := cp.NewClientService(pool)
	service := NewService(resolver, clientService)

	ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())

	testFile := &storage.File{
		ID: "test-file-id",
		FileMetadata: storagetypes.FileMetadata{
			Key:            "test-file-id", // Use same value as ID for testing
			IntegrationID:  "test-integration",
			HushID:         "test-hush",
			OrganizationID: "test-org",
		},
	}

	_, err := service.Download(ctx, testFile)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrProviderResolutionFailed)
}

func TestAuthContext_Verification(t *testing.T) {
	// Test that auth context functions work as expected - use generated ULIDs
	testUserID := ulids.New().String()
	testOrgID := ulids.New().String()

	ctx := auth.NewTestContextWithOrgID(testUserID, testOrgID)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	require.NoError(t, err, "Should be able to get organization ID from test context")
	assert.Equal(t, testOrgID, orgID)

	subjectID, err := auth.GetSubjectIDFromContext(ctx)
	require.NoError(t, err, "Should be able to get subject ID from test context")
	assert.Equal(t, testUserID, subjectID)
}
