package r2_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	storage "github.com/theopenlane/core/pkg/objects/storage"
	r2provider "github.com/theopenlane/core/pkg/objects/storage/providers/r2"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func r2Options() *storage.ProviderOptions {
	return storage.NewProviderOptions(
		storage.WithBucket("test-bucket"),
		storage.WithEndpoint("https://account.r2.cloudflarestorage.com"),
		storage.WithCredentials(storage.ProviderCredentials{
			AccountID:       "account",
			AccessKeyID:     "access",
			SecretAccessKey: "secret",
		}),
	)
}

func TestNewR2Provider(t *testing.T) {
	tests := []struct {
		name        string
		options     *storage.ProviderOptions
		expectError bool
	}{
		{
			name:    "valid configuration",
			options: r2Options(),
		},
		{
			name:        "missing bucket",
			options:     storage.NewProviderOptions(storage.WithCredentials(storage.ProviderCredentials{AccountID: "account", AccessKeyID: "access", SecretAccessKey: "secret"})),
			expectError: true,
		},
		{
			name: "missing account",
			options: storage.NewProviderOptions(
				storage.WithBucket("bucket"),
				storage.WithCredentials(storage.ProviderCredentials{AccessKeyID: "access", SecretAccessKey: "secret"}),
			),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := r2provider.NewR2Provider(tt.options)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				if err != nil {
					t.Skip("Skipping test due to missing R2-compatible environment")
				}
				assert.NotNil(t, provider)
				assert.NoError(t, provider.Close())
			}
		})
	}
}

func TestR2ProviderMethods(t *testing.T) {
	provider, err := r2provider.NewR2Provider(r2Options())
	if err != nil {
		t.Skip("Skipping test due to missing R2-compatible environment")
	}
	defer provider.Close()

	t.Run("ProviderType", func(t *testing.T) {
		assert.Equal(t, storagetypes.R2Provider, provider.ProviderType())
	})

	t.Run("GetScheme", func(t *testing.T) {
		scheme := provider.GetScheme()
		assert.NotNil(t, scheme)
		assert.Equal(t, "r2://", *scheme)
	})
}

func TestR2ProviderPresignedURL(t *testing.T) {
	provider, err := r2provider.NewR2Provider(r2Options())
	if err != nil {
		t.Skip("Skipping test due to missing R2-compatible environment")
	}
	defer provider.Close()

	file := &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "test-file.txt"}}
	url, err := provider.GetPresignedURL(context.Background(), file, &storagetypes.PresignedURLOptions{Duration: time.Minute})
	if err != nil {
		t.Skip("Skipping presigned URL test without R2 environment")
	}
	assert.True(t, strings.Contains(url, "test-file.txt"))
}

func TestNewR2ProviderFromCredentials(t *testing.T) {
	creds := storage.ProviderCredentials{AccountID: "account", AccessKeyID: "access", SecretAccessKey: "secret"}
	opts := storage.NewProviderOptions(storage.WithBucket("bucket"))

	result := r2provider.NewR2ProviderFromCredentials(creds, opts)
	if result.IsError() {
		t.Skip("Skipping test due to missing R2-compatible environment")
	}

	assert.NotNil(t, result.MustGet())
}

func TestR2Provider_Upload_Errors(t *testing.T) {
	t.Run("empty file name", func(t *testing.T) {
		opts := r2Options()
		provider, err := r2provider.NewR2Provider(opts)
		if err != nil {
			t.Skip("Skipping test due to missing R2 credentials or environment")
		}

		reader := strings.NewReader("test content")
		uploadOpts := &storagetypes.UploadFileOptions{
			FileName: "",
		}
		_, err = provider.Upload(context.Background(), reader, uploadOpts)
		assert.Error(t, err)
	})
}

func TestR2Provider_Download_Errors(t *testing.T) {
	t.Run("empty file key", func(t *testing.T) {
		opts := r2Options()
		provider, err := r2provider.NewR2Provider(opts)
		if err != nil {
			t.Skip("Skipping test due to missing R2 credentials or environment")
		}

		file := &storagetypes.File{
			FileMetadata: storagetypes.FileMetadata{
				Key: "",
			},
		}
		downloadOpts := &storagetypes.DownloadFileOptions{
			FileName: "test.txt",
		}
		_, err = provider.Download(context.Background(), file, downloadOpts)
		assert.Error(t, err)
	})
}

func TestR2Provider_Delete_Errors(t *testing.T) {
	t.Run("empty file key", func(t *testing.T) {
		opts := r2Options()
		provider, err := r2provider.NewR2Provider(opts)
		if err != nil {
			t.Skip("Skipping test due to missing R2 credentials or environment")
		}

		file := &storagetypes.File{
			FileMetadata: storagetypes.FileMetadata{
				Key: "",
			},
		}
		err = provider.Delete(context.Background(), file, nil)
		assert.Error(t, err)
	})
}

func TestR2Provider_Exists_Errors(t *testing.T) {
	t.Run("empty file key", func(t *testing.T) {
		opts := r2Options()
		provider, err := r2provider.NewR2Provider(opts)
		if err != nil {
			t.Skip("Skipping test due to missing R2 credentials or environment")
		}

		file := &storagetypes.File{
			FileMetadata: storagetypes.FileMetadata{
				Key: "",
			},
		}
		_, err = provider.Exists(context.Background(), file)
		assert.Error(t, err)
	})
}

func TestR2Provider_ListBuckets(t *testing.T) {
	t.Run("list buckets", func(t *testing.T) {
		opts := r2Options()
		provider, err := r2provider.NewR2Provider(opts)
		if err != nil {
			t.Skip("Skipping test due to missing R2 credentials or environment")
		}

		_, err = provider.ListBuckets()
		assert.Error(t, err)
	})
}
