package s3_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/theopenlane/iam/tokens"
	storage "github.com/theopenlane/shared/objects/storage"
	s3provider "github.com/theopenlane/shared/objects/storage/providers/s3"
	"github.com/theopenlane/shared/objects/storage/proxy"
	storagetypes "github.com/theopenlane/shared/objects/storage/types"
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

const (
	minioImage  = "minio/minio:latest"
	minioUser   = "provider1"
	minioSecret = "provider1secret"
	minioBucket = "provider1-bucket"
	testRegion  = "us-east-1"
)

func TestS3ProviderPresignURLProxyDisabled(t *testing.T) {
	ctx := context.Background()
	endpoint, terminate := startMinio(t, ctx)
	t.Cleanup(func() { _ = terminate(context.Background()) })

	createBucket(t, ctx, endpoint, minioBucket)

	builder := s3provider.NewS3Builder().WithOptions(s3provider.WithUsePathStyle(true))

	options := storage.NewProviderOptions(
		storage.WithBucket(minioBucket),
		storage.WithRegion(testRegion),
		storage.WithEndpoint(endpoint),
		storage.WithCredentials(storage.ProviderCredentials{AccessKeyID: minioUser, SecretAccessKey: minioSecret}),
		storage.WithProxyPresignEnabled(false),
	)

	providerInterface, err := builder.Build(ctx, storage.ProviderCredentials{AccessKeyID: minioUser, SecretAccessKey: minioSecret}, options)
	require.NoError(t, err)
	provider := providerInterface
	t.Cleanup(func() { _ = provider.Close() })

	content := []byte("hello, minio")
	_, err = provider.Upload(ctx, bytes.NewReader(content), &storagetypes.UploadFileOptions{
		FileName:    "folder/object.txt",
		ContentType: "text/plain",
	})
	require.NoError(t, err)

	url, err := provider.GetPresignedURL(ctx, &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			Bucket: minioBucket,
			Key:    "folder/object.txt",
		},
	}, &storagetypes.PresignedURLOptions{Duration: time.Minute})
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(url, "http://"), "expected upstream S3 URL, got %q", url)
	require.Contains(t, url, "X-Amz-Algorithm=AWS4-HMAC-SHA256", "URL should contain AWS signature")
	require.Contains(t, url, "X-Amz-Credential=", "URL should contain AWS credentials")
	require.Contains(t, url, "X-Amz-Signature=", "URL should contain AWS signature")
	require.Contains(t, url, "folder/object.txt", "URL should contain object key")

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, string(content), string(body))
}

func TestS3ProviderPresignURLProxyEnabled(t *testing.T) {
	ctx := context.Background()
	endpoint, terminate := startMinio(t, ctx)
	t.Cleanup(func() { _ = terminate(context.Background()) })

	createBucket(t, ctx, endpoint, minioBucket)

	tm := newTestTokenManager(t)
	baseURL := "http://localhost:17608"

	builder := s3provider.NewS3Builder().WithOptions(s3provider.WithUsePathStyle(true))

	options := storage.NewProviderOptions(
		storage.WithBucket(minioBucket),
		storage.WithRegion(testRegion),
		storage.WithEndpoint(endpoint),
		storage.WithCredentials(storage.ProviderCredentials{AccessKeyID: minioUser, SecretAccessKey: minioSecret}),
		storage.WithProxyPresignEnabled(true),
		storage.WithProxyPresignConfig(&storage.ProxyPresignConfig{
			TokenManager: tm,
			BaseURL:      baseURL,
		}),
	)

	providerInterface, err := builder.Build(ctx, storage.ProviderCredentials{AccessKeyID: minioUser, SecretAccessKey: minioSecret}, options)
	require.NoError(t, err)
	provider := providerInterface
	t.Cleanup(func() { _ = provider.Close() })

	content := []byte("hello, minio with proxy enabled")
	_, err = provider.Upload(ctx, bytes.NewReader(content), &storagetypes.UploadFileOptions{
		FileName:    "folder/object.txt",
		ContentType: "text/plain",
	})
	require.NoError(t, err)

	file := &storagetypes.File{
		ID: "01HZZTESTFILE00000000000000",
		FileMetadata: storagetypes.FileMetadata{
			Bucket:  minioBucket,
			Key:     "folder/object.txt",
			FullURI: "s3://" + minioBucket + "/folder/object.txt",
		},
	}

	secret := make([]byte, 128)
	_, err = rand.Read(secret)
	require.NoError(t, err)

	url, err := proxy.GenerateDownloadURLWithSecret(file, secret, time.Minute, &storage.ProxyPresignConfig{
		TokenManager: tm,
		BaseURL:      baseURL,
	})
	require.NoError(t, err, "with proxy presign enabled, should generate proxy URL")
	require.NotEmpty(t, url)
	require.True(t, strings.HasPrefix(url, baseURL), "URL should start with BaseURL, got %q", url)
	require.Contains(t, url, file.ID, "URL should contain file ID")
	require.Contains(t, url, "token=", "URL should contain token parameter")
	require.NotContains(t, url, "X-Amz-Algorithm", "proxy URL should not contain AWS signature parameters")
}

func startMinio(t *testing.T, ctx context.Context) (string, func(context.Context, ...testcontainers.TerminateOption) error) {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        minioImage,
		ExposedPorts: []string{"9000/tcp"},
		Env: map[string]string{
			"MINIO_ROOT_USER":     minioUser,
			"MINIO_ROOT_PASSWORD": minioSecret,
		},
		Cmd:        []string{"server", "/data"},
		WaitingFor: wait.ForLog("API:").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "9000/tcp")
	require.NoError(t, err)

	endpoint := "http://" + host + ":" + port.Port()
	return endpoint, container.Terminate
}

func createBucket(t *testing.T, ctx context.Context, endpoint, bucket string) {
	awsCfg := aws.Config{
		Region:       testRegion,
		Credentials:  credentials.NewStaticCredentialsProvider(minioUser, minioSecret, ""),
		BaseEndpoint: aws.String(endpoint),
	}

	client := awss3.NewFromConfig(awsCfg, func(o *awss3.Options) {
		o.UsePathStyle = true
		o.Region = testRegion
		o.Credentials = credentials.NewStaticCredentialsProvider(minioUser, minioSecret, "")
	})

	_, err := client.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: aws.String(bucket)})
	if err != nil {
		var existsErr *types.BucketAlreadyOwnedByYou
		if !errors.As(err, &existsErr) {
			require.NoError(t, err)
		}
	}

	waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	waiter := awss3.NewBucketExistsWaiter(client)
	require.NoError(t, waiter.Wait(waitCtx, &awss3.HeadBucketInput{Bucket: aws.String(bucket)}, 1*time.Second))
}

func newTestTokenManager(t *testing.T) *tokens.TokenManager {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	conf := tokens.Config{
		Audience:        "http://localhost:17608",
		Issuer:          "http://localhost:17608",
		AccessDuration:  time.Hour,
		RefreshDuration: 2 * time.Hour,
		RefreshOverlap:  -15 * time.Minute,
	}
	tm, err := tokens.NewWithKey(key, conf)
	require.NoError(t, err)
	return tm
}
