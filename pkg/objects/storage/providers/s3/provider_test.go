package s3_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"

	storage "github.com/theopenlane/core/pkg/objects/storage"
	s3provider "github.com/theopenlane/core/pkg/objects/storage/providers/s3"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func providerOptions() *storage.ProviderOptions {
	return storage.NewProviderOptions(
		storage.WithBucket("test-bucket"),
		storage.WithRegion("us-east-1"),
		storage.WithCredentials(storage.ProviderCredentials{
			AccessKeyID:     "test-access-key",
			SecretAccessKey: "test-secret-key",
		}),
	)
}

func TestNewS3Provider(t *testing.T) {
	opts := providerOptions()

	provider, err := s3provider.NewS3Provider(opts)
	if err != nil {
		t.Skip("Skipping test due to missing AWS credentials or environment")
	}

	assert.NotNil(t, provider)
}

func TestNewS3ProviderWithOptions(t *testing.T) {
	opts := providerOptions()
	provider, err := s3provider.NewS3Provider(opts, s3provider.WithUsePathStyle(true), s3provider.WithDebugMode(true), s3provider.WithAWSConfig(aws.Config{Region: "us-east-1"}))
	if err != nil {
		t.Skip("Skipping test due to missing AWS credentials or environment")
	}

	assert.NotNil(t, provider)
}

func TestNewS3ProviderMissingBucket(t *testing.T) {
	opts := providerOptions()
	opts.Bucket = ""

	_, err := s3provider.NewS3Provider(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bucket")
}

func TestNewS3ProviderMissingRegion(t *testing.T) {
	opts := providerOptions()
	opts.Region = ""

	_, err := s3provider.NewS3Provider(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required S3 credentials")
}

func TestS3ProviderConstants(t *testing.T) {
	assert.Equal(t, 15*time.Minute, s3provider.DefaultPresignedURLExpiry)
	assert.Equal(t, 64*1024*1024, s3provider.DefaultPartSize)
	assert.Equal(t, 5, s3provider.DefaultConcurrency)
}

func TestS3ProviderMethods(t *testing.T) {
	opts := providerOptions()

	provider, err := s3provider.NewS3Provider(opts)
	if err != nil {
		t.Skip("Skipping test due to missing AWS credentials or environment")
	}

	t.Run("ProviderType", func(t *testing.T) {
		assert.Equal(t, storagetypes.S3Provider, provider.ProviderType())
	})

	t.Run("GetScheme", func(t *testing.T) {
		scheme := provider.GetScheme()
		assert.NotNil(t, scheme)
		assert.Equal(t, "s3://", *scheme)
	})

	t.Run("Close", func(t *testing.T) {
		assert.NoError(t, provider.Close())
	})
}

func TestNewS3ProviderResult(t *testing.T) {
	opts := providerOptions()
	result := s3provider.NewS3ProviderResult(opts)
	if result.IsError() {
		t.Skip("Skipping test due to missing AWS credentials or environment")
	}

	provider := result.MustGet()
	assert.NotNil(t, provider)
}

func TestNewS3ProviderFromCredentials(t *testing.T) {
	creds := storage.ProviderCredentials{AccessKeyID: "key", SecretAccessKey: "secret"}
	opts := storage.NewProviderOptions(storage.WithBucket("bucket"), storage.WithRegion("us-east-1"))

	result := s3provider.NewS3ProviderFromCredentials(creds, opts)
	if result.IsError() {
		t.Skip("Skipping test due to missing AWS credentials or environment")
	}

	assert.NotNil(t, result.MustGet())

	t.Run("missing options", func(t *testing.T) {
		errResult := s3provider.NewS3ProviderFromCredentials(creds, nil)
		assert.True(t, errResult.IsError())
	})
}

func TestWithACL(t *testing.T) {
	t.Run("set ACL option", func(t *testing.T) {
		opts := providerOptions()
		provider, err := s3provider.NewS3Provider(opts, s3provider.WithACL("public-read"))
		if err != nil {
			t.Skip("Skipping test due to missing AWS credentials or environment")
		}

		assert.NotNil(t, provider)
	})

	t.Run("private ACL", func(t *testing.T) {
		opts := providerOptions()
		provider, err := s3provider.NewS3Provider(opts, s3provider.WithACL("private"))
		if err != nil {
			t.Skip("Skipping test due to missing AWS credentials or environment")
		}

		assert.NotNil(t, provider)
	})
}

func TestS3Provider_Upload_Errors(t *testing.T) {
	t.Run("empty file name", func(t *testing.T) {
		opts := providerOptions()
		provider, err := s3provider.NewS3Provider(opts)
		if err != nil {
			t.Skip("Skipping test due to missing AWS credentials or environment")
		}

		reader := strings.NewReader("test content")
		uploadOpts := &storagetypes.UploadFileOptions{
			FileName: "",
		}
		_, err = provider.Upload(context.Background(), reader, uploadOpts)
		assert.Error(t, err)
	})
}

func TestS3Provider_Download_Errors(t *testing.T) {
	t.Run("empty file key", func(t *testing.T) {
		opts := providerOptions()
		provider, err := s3provider.NewS3Provider(opts)
		if err != nil {
			t.Skip("Skipping test due to missing AWS credentials or environment")
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

func TestS3Provider_Delete_Errors(t *testing.T) {
	t.Run("empty file key", func(t *testing.T) {
		opts := providerOptions()
		provider, err := s3provider.NewS3Provider(opts)
		if err != nil {
			t.Skip("Skipping test due to missing AWS credentials or environment")
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

func TestS3Provider_GetPresignedURL_Errors(t *testing.T) {
	t.Run("empty file key", func(t *testing.T) {
		opts := providerOptions()
		provider, err := s3provider.NewS3Provider(opts)
		if err != nil {
			t.Skip("Skipping test due to missing AWS credentials or environment")
		}

		file := &storagetypes.File{
			FileMetadata: storagetypes.FileMetadata{
				Key: "",
			},
		}
		urlOpts := &storagetypes.PresignedURLOptions{
			Duration: 15 * time.Minute,
		}
		_, err = provider.GetPresignedURL(context.Background(), file, urlOpts)
		assert.Error(t, err)
	})
}

func TestS3Provider_Exists_Errors(t *testing.T) {
	t.Run("empty file key", func(t *testing.T) {
		opts := providerOptions()
		provider, err := s3provider.NewS3Provider(opts)
		if err != nil {
			t.Skip("Skipping test due to missing AWS credentials or environment")
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

func TestS3Provider_ListBuckets(t *testing.T) {
	t.Run("list buckets", func(t *testing.T) {
		opts := providerOptions()
		provider, err := s3provider.NewS3Provider(opts)
		if err != nil {
			t.Skip("Skipping test due to missing AWS credentials or environment")
		}

		_, err = provider.ListBuckets()
		assert.Error(t, err)
	})
}

func TestWithOptions_Builder(t *testing.T) {
	t.Run("apply provider options", func(t *testing.T) {
		builder := s3provider.NewS3Builder()

		builder = builder.WithOptions(s3provider.WithACL("private"), s3provider.WithUsePathStyle(true), s3provider.WithDebugMode(false))
		assert.NotNil(t, builder)
	})
}
