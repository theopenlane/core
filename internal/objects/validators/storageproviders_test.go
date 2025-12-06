package validators

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func TestValidateConfiguredStorageProvidersDevModeDisk(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	bucket := filepath.Join(tempDir, "uploads")

	cfg := storage.ProviderConfig{
		Enabled: true,
		DevMode: true,
		Providers: storage.Providers{
			Disk: storage.ProviderConfigs{
				Enabled: false,
				Bucket:  bucket,
			},
		},
	}

	errs := ValidateAvailabilityByProvider(context.Background(), cfg)
	require.Empty(t, errs)

	// devMode uses the defaultDevStorageBucket - so even if Disk provider is created, this is the directory we want to ensure exists
	_, err := os.Stat(objects.DefaultDevStorageBucket)
	require.NoError(t, err, "expected dev-mode validation to create bucket directory")
}

func TestValidateAvailabilityByProviderDevMode(t *testing.T) {
	cfg := storage.ProviderConfig{
		Enabled: true,
		DevMode: true,
		Providers: storage.Providers{
			// devmode should work regardless of individual disk configuration being on / off
			Disk: storage.ProviderConfigs{
				Enabled:         false,
				EnsureAvailable: false,
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

func TestConfigKeyMismatchLeavesR2Unpopulated(t *testing.T) {
	yamlWithCloudflareR2Key := `
enabled: true
providers:
  s3:
    enabled: true
    bucket: "opln"
    region: "us-east-2"
  cloudflarer2:
    enabled: true
    bucket: "ol-trust-center"
    region: "WNAM"
`

	yamlWithCorrectR2Key := `
enabled: true
providers:
  s3:
    enabled: true
    bucket: "opln"
    region: "us-east-2"
  r2:
    enabled: true
    bucket: "ol-trust-center"
    region: "WNAM"
`

	t.Run("cloudflarer2 YAML key does not populate R2 struct field with r2 koanf tag", func(t *testing.T) {
		var cfg storage.ProviderConfig
		err := yaml.Unmarshal([]byte(yamlWithCloudflareR2Key), &cfg)
		require.NoError(t, err)

		require.True(t, cfg.Providers.S3.Enabled, "S3 config should be populated")
		require.Equal(t, "opln", cfg.Providers.S3.Bucket)
		require.Equal(t, "us-east-2", cfg.Providers.S3.Region)

		require.False(t, cfg.Providers.R2.Enabled, "R2.Enabled defaults to false when YAML uses cloudflarer2 key but struct expects r2")
		require.Empty(t, cfg.Providers.R2.Bucket, "R2.Bucket empty when YAML key mismatch")
		require.Empty(t, cfg.Providers.R2.Region, "R2.Region empty when YAML key mismatch")
	})

	t.Run("r2 YAML key correctly populates R2 struct field with r2 koanf tag", func(t *testing.T) {
		var cfg storage.ProviderConfig
		err := yaml.Unmarshal([]byte(yamlWithCorrectR2Key), &cfg)
		require.NoError(t, err)

		require.True(t, cfg.Providers.S3.Enabled)
		require.Equal(t, "opln", cfg.Providers.S3.Bucket)

		require.True(t, cfg.Providers.R2.Enabled, "R2.Enabled true when YAML r2 key matches struct r2 koanf tag")
		require.Equal(t, "ol-trust-center", cfg.Providers.R2.Bucket, "R2.Bucket populated when keys match")
		require.Equal(t, "WNAM", cfg.Providers.R2.Region, "R2.Region populated when keys match")
	})
}
