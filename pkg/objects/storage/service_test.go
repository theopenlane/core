package storage_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/pkg/objects/storage"
)

func TestNewObjectService(t *testing.T) {
	service := storage.NewObjectService()

	assert.NotNil(t, service)
	assert.Equal(t, int64(storage.DefaultMaxFileSize), service.MaxSize())
	assert.Equal(t, int64(storage.DefaultMaxMemory), service.MaxMemory())
	assert.Equal(t, []string{storage.DefaultUploadFileKey}, service.Keys())
	assert.False(t, service.IgnoreNonExistentKeys())
	assert.NotNil(t, service.Skipper())
	assert.NotNil(t, service.ErrorResponseHandler())
}

func TestObjectServiceAccessors(t *testing.T) {
	service := storage.NewObjectService()

	// Test all accessor methods
	assert.Equal(t, int64(storage.DefaultMaxFileSize), service.MaxSize())
	assert.Equal(t, int64(storage.DefaultMaxMemory), service.MaxMemory())
	assert.Equal(t, []string{storage.DefaultUploadFileKey}, service.Keys())
	assert.False(t, service.IgnoreNonExistentKeys())
	assert.NotNil(t, service.Skipper())
	assert.NotNil(t, service.ErrorResponseHandler())

	// Test skipper function
	skipper := service.Skipper()
	assert.False(t, skipper(nil)) // DefaultSkipper always returns false

	// Test error response handler
	handler := service.ErrorResponseHandler()
	assert.NotNil(t, handler)
}

type mockProvider struct {
	uploadFunc       func(ctx context.Context, reader io.Reader, opts *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error)
	downloadFunc     func(ctx context.Context, file *storagetypes.File, opts *storagetypes.DownloadFileOptions) (*storagetypes.DownloadedFileMetadata, error)
	deleteFunc       func(ctx context.Context, file *storagetypes.File, opts *storagetypes.DeleteFileOptions) error
	getPresignedFunc func(ctx context.Context, file *storagetypes.File, opts *storagetypes.PresignedURLOptions) (string, error)
	existsFunc       func(ctx context.Context, file *storagetypes.File) (bool, error)
	providerType     storagetypes.ProviderType
}

func (m *mockProvider) Upload(ctx context.Context, reader io.Reader, opts *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
	if m.uploadFunc != nil {
		return m.uploadFunc(ctx, reader, opts)
	}
	return &storagetypes.UploadedFileMetadata{
		TimeUploaded: time.Now(),
		FileMetadata: storagetypes.FileMetadata{
			Key:  "test-key",
			Size: 100,
		},
	}, nil
}

func (m *mockProvider) Download(ctx context.Context, file *storagetypes.File, opts *storagetypes.DownloadFileOptions) (*storagetypes.DownloadedFileMetadata, error) {
	if m.downloadFunc != nil {
		return m.downloadFunc(ctx, file, opts)
	}
	return &storagetypes.DownloadedFileMetadata{
		File:           []byte("test content"),
		Size:           12,
		TimeDownloaded: time.Now(),
	}, nil
}

func (m *mockProvider) Delete(ctx context.Context, file *storagetypes.File, opts *storagetypes.DeleteFileOptions) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, file, opts)
	}
	return nil
}

func (m *mockProvider) GetPresignedURL(ctx context.Context, file *storagetypes.File, opts *storagetypes.PresignedURLOptions) (string, error) {
	if m.getPresignedFunc != nil {
		return m.getPresignedFunc(ctx, file, opts)
	}
	return "https://example.com/presigned-url", nil
}

func (m *mockProvider) Exists(ctx context.Context, file *storagetypes.File) (bool, error) {
	if m.existsFunc != nil {
		return m.existsFunc(ctx, file)
	}
	return true, nil
}

func (m *mockProvider) GetScheme() *string {
	scheme := "mock"
	return &scheme
}

func (m *mockProvider) ListBuckets() ([]string, error) {
	return []string{"bucket1", "bucket2"}, nil
}

func (m *mockProvider) ProviderType() storagetypes.ProviderType {
	if m.providerType != "" {
		return m.providerType
	}
	return "mock"
}

func (m *mockProvider) Close() error {
	return nil
}

func TestObjectService_Upload(t *testing.T) {
	t.Run("successful upload", func(t *testing.T) {
		service := storage.NewObjectService()
		provider := &mockProvider{providerType: "s3"}
		reader := strings.NewReader("test content")

		opts := &storage.UploadOptions{
			FileName:    "test.txt",
			ContentType: "text/plain",
			Bucket:      "test-bucket",
		}
		opts.Key = "files"

		file, err := service.Upload(context.Background(), provider, reader, opts)
		require.NoError(t, err)
		require.NotNil(t, file)
		assert.Equal(t, "test.txt", file.OriginalName)
		assert.Equal(t, storagetypes.ProviderType("s3"), file.ProviderType)
	})

	t.Run("upload with content type detection", func(t *testing.T) {
		service := storage.NewObjectService()
		provider := &mockProvider{}
		reader := bytes.NewReader([]byte(`{"key": "value"}`))

		opts := &storage.UploadOptions{
			FileName: "test.json",
			Bucket:   "test-bucket",
		}
		opts.Key = "files"

		file, err := service.Upload(context.Background(), provider, reader, opts)
		require.NoError(t, err)
		require.NotNil(t, file)
		assert.Contains(t, file.ContentType, "json")
	})

	t.Run("upload with validation failure", func(t *testing.T) {
		service := storage.NewObjectService()
		validationErr := errors.New("validation failed")
		service = service.WithValidation(func(f storage.File) error {
			return validationErr
		})

		provider := &mockProvider{}
		reader := strings.NewReader("test content")

		opts := &storage.UploadOptions{
			FileName:    "test.txt",
			ContentType: "text/plain",
		}
		opts.Key = "files"

		file, err := service.Upload(context.Background(), provider, reader, opts)
		assert.Error(t, err)
		assert.Nil(t, file)
		assert.Equal(t, validationErr, err)
	})

	t.Run("upload provider error", func(t *testing.T) {
		service := storage.NewObjectService()
		uploadErr := errors.New("upload failed")
		provider := &mockProvider{
			uploadFunc: func(_ context.Context, _ io.Reader, _ *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
				return nil, uploadErr
			},
		}
		reader := strings.NewReader("test content")

		opts := &storage.UploadOptions{
			FileName: "test.txt",
		}
		opts.Key = "files"

		file, err := service.Upload(context.Background(), provider, reader, opts)
		assert.Error(t, err)
		assert.Nil(t, file)
		assert.Equal(t, uploadErr, err)
	})
}

func TestObjectService_Download(t *testing.T) {
	t.Run("successful download", func(t *testing.T) {
		service := storage.NewObjectService()
		provider := &mockProvider{}

		file := &storagetypes.File{
			ID: "test-id",
			FileMetadata: storage.FileMetadata{
				Key: "test-key",
			},
		}

		opts := &storagetypes.DownloadFileOptions{
			FileName: "download.txt",
		}

		metadata, err := service.Download(context.Background(), provider, file, opts)
		require.NoError(t, err)
		require.NotNil(t, metadata)
		assert.Equal(t, []byte("test content"), metadata.File)
	})

	t.Run("download error", func(t *testing.T) {
		service := storage.NewObjectService()
		downloadErr := errors.New("download failed")
		provider := &mockProvider{
			downloadFunc: func(_ context.Context, _ *storagetypes.File, _ *storagetypes.DownloadFileOptions) (*storagetypes.DownloadedFileMetadata, error) {
				return nil, downloadErr
			},
		}

		file := &storagetypes.File{ID: "test-id"}
		opts := &storagetypes.DownloadFileOptions{FileName: "test.txt"}

		metadata, err := service.Download(context.Background(), provider, file, opts)
		assert.Error(t, err)
		assert.Nil(t, metadata)
		assert.Equal(t, downloadErr, err)
	})
}

func TestObjectService_GetPresignedURL(t *testing.T) {
	t.Run("successful presigned URL", func(t *testing.T) {
		service := storage.NewObjectService()
		provider := &mockProvider{}

		file := &storagetypes.File{
			ID: "test-id",
			FileMetadata: storage.FileMetadata{
				Key: "test-key",
			},
		}

		opts := &storagetypes.PresignedURLOptions{
			Duration: 1 * time.Hour,
		}

		url, err := service.GetPresignedURL(context.Background(), provider, file, opts)
		require.NoError(t, err)
		assert.Equal(t, "https://example.com/presigned-url", url)
	})

	t.Run("presigned URL error", func(t *testing.T) {
		service := storage.NewObjectService()
		urlErr := errors.New("presigned URL failed")
		provider := &mockProvider{
			getPresignedFunc: func(_ context.Context, _ *storagetypes.File, _ *storagetypes.PresignedURLOptions) (string, error) {
				return "", urlErr
			},
		}

		file := &storagetypes.File{ID: "test-id"}
		opts := &storagetypes.PresignedURLOptions{Duration: 1 * time.Hour}

		url, err := service.GetPresignedURL(context.Background(), provider, file, opts)
		assert.Error(t, err)
		assert.Empty(t, url)
		assert.Equal(t, urlErr, err)
	})
}

func TestObjectService_Delete(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		service := storage.NewObjectService()
		provider := &mockProvider{}

		file := &storagetypes.File{
			ID: "test-id",
			FileMetadata: storage.FileMetadata{
				Key: "test-key",
			},
		}

		opts := &storagetypes.DeleteFileOptions{
			Reason: "cleanup",
		}

		err := service.Delete(context.Background(), provider, file, opts)
		assert.NoError(t, err)
	})

	t.Run("delete error", func(t *testing.T) {
		service := storage.NewObjectService()
		deleteErr := errors.New("delete failed")
		provider := &mockProvider{
			deleteFunc: func(_ context.Context, _ *storagetypes.File, _ *storagetypes.DeleteFileOptions) error {
				return deleteErr
			},
		}

		file := &storagetypes.File{ID: "test-id"}
		opts := &storagetypes.DeleteFileOptions{Reason: "test"}

		err := service.Delete(context.Background(), provider, file, opts)
		assert.Error(t, err)
		assert.Equal(t, deleteErr, err)
	})
}

func TestObjectService_WithUploader(t *testing.T) {
	t.Run("set custom uploader", func(t *testing.T) {
		service := storage.NewObjectService()

		customUploader := func(_ context.Context, _ *storage.ObjectService, files []storage.File) ([]storage.File, error) {
			return files, nil
		}

		newService := service.WithUploader(customUploader)
		require.NotNil(t, newService)
		assert.NotEqual(t, service, newService)
	})

	t.Run("original service unchanged", func(t *testing.T) {
		service := storage.NewObjectService()
		originalKeys := service.Keys()

		customUploader := func(_ context.Context, _ *storage.ObjectService, _ []storage.File) ([]storage.File, error) {
			return nil, nil
		}

		newService := service.WithUploader(customUploader)
		assert.Equal(t, originalKeys, service.Keys())
		assert.Equal(t, originalKeys, newService.Keys())
	})
}

func TestObjectService_WithValidation(t *testing.T) {
	t.Run("set custom validation", func(t *testing.T) {
		service := storage.NewObjectService()

		customValidation := func(f storage.File) error {
			if f.Size > 1000 {
				return errors.New("file too large")
			}
			return nil
		}

		newService := service.WithValidation(customValidation)
		require.NotNil(t, newService)
		assert.NotEqual(t, service, newService)
	})

	t.Run("validation affects upload", func(t *testing.T) {
		service := storage.NewObjectService()
		service = service.WithValidation(func(f storage.File) error {
			if f.OriginalName == "forbidden.txt" {
				return errors.New("forbidden file")
			}
			return nil
		})

		provider := &mockProvider{}
		reader := strings.NewReader("test")
		opts := &storage.UploadOptions{
			FileName: "forbidden.txt",
		}
		opts.Key = "files"

		file, err := service.Upload(context.Background(), provider, reader, opts)
		assert.Error(t, err)
		assert.Nil(t, file)
		assert.Contains(t, err.Error(), "forbidden")
	})
}
