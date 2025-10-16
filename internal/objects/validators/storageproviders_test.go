package validators

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func TestValidateConfiguredStorageProvidersDisabled(t *testing.T) {
	errs := ValidateConfiguredStorageProviders(context.Background(), storage.ProviderConfig{
		Enabled: false,
	})
	require.Empty(t, errs)
}

func TestValidateConfiguredStorageProvidersDevModeDisk(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	bucket := filepath.Join(tempDir, "uploads")

	cfg := storage.ProviderConfig{
		Enabled: true,
		DevMode: true,
		Providers: storage.Providers{
			Disk: storage.ProviderConfigs{
				Enabled: true,
				Bucket:  bucket,
			},
		},
	}

	errs := ValidateConfiguredStorageProviders(context.Background(), cfg)
	require.Empty(t, errs)

	_, err := os.Stat(bucket)
	require.NoError(t, err, "expected dev-mode validation to create bucket directory")
}

func TestValidateAvailabilityByProviderDevMode(t *testing.T) {
	cfg := storage.ProviderConfig{
		Enabled: true,
		DevMode: true,
		Providers: storage.Providers{
			Disk: storage.ProviderConfigs{
				Enabled:         true,
				EnsureAvailable: true,
			},
		},
	}

	errs := ValidateAvailabilityByProvider(context.Background(), cfg)
	require.Empty(t, errs)
}

func TestValidateAvailabilityByProviderDisk(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	bucket := filepath.Join(tempDir, "bucket")

	cfg := storage.ProviderConfig{
		Enabled: true,
		Providers: storage.Providers{
			Disk: storage.ProviderConfigs{
				Enabled:         true,
				EnsureAvailable: true,
				Bucket:          bucket,
			},
		},
	}

	errs := ValidateAvailabilityByProvider(context.Background(), cfg)
	require.Empty(t, errs)
}

func TestValidateBuckets(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		provider := &stubProvider{buckets: []string{"primary"}}
		require.NoError(t, validateBuckets("disk", provider, "primary"))
	})

	t.Run("missing bucket", func(t *testing.T) {
		provider := &stubProvider{buckets: []string{"secondary"}}
		err := validateBuckets("disk", provider, "primary")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrBucketNotFound)
	})

	t.Run("list error", func(t *testing.T) {
		provider := &stubProvider{listErr: errors.New("boom")}
		err := validateBuckets("disk", provider, "")
		require.Error(t, err)
	})
}

func TestEnsureDirectoryExists(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "nested", "path")

	require.NoError(t, ensureDirectoryExists(target))
	info, err := os.Stat(target)
	require.NoError(t, err)
	require.True(t, info.IsDir())
}

func TestStorageAvailabilityCheck(t *testing.T) {
	check := StorageAvailabilityCheck(func() storage.ProviderConfig {
		return storage.ProviderConfig{Enabled: false}
	})

	require.NoError(t, check(context.Background()))
}

type stubProvider struct {
	buckets []string
	listErr error
}

func (s *stubProvider) Upload(context.Context, io.Reader, *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
	return nil, nil
}

func (s *stubProvider) Download(context.Context, *storagetypes.File, *storagetypes.DownloadFileOptions) (*storagetypes.DownloadedFileMetadata, error) {
	return nil, nil
}

func (s *stubProvider) Delete(context.Context, *storagetypes.File, *storagetypes.DeleteFileOptions) error {
	return nil
}

func (s *stubProvider) GetPresignedURL(context.Context, *storagetypes.File, *storagetypes.PresignedURLOptions) (string, error) {
	return "", nil
}

func (s *stubProvider) Exists(context.Context, *storagetypes.File) (bool, error) {
	return false, nil
}

func (s *stubProvider) GetScheme() *string {
	return nil
}

func (s *stubProvider) ListBuckets() ([]string, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.buckets, nil
}

func (s *stubProvider) ProviderType() storagetypes.ProviderType {
	return storagetypes.DiskProvider
}

func (s *stubProvider) Close() error {
	return nil
}

// ensure stubProvider satisfies storagetypes.Provider at compile-time.
var _ storagetypes.Provider = (*stubProvider)(nil)
