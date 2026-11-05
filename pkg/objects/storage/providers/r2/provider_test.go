package r2_test

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
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	r2provider "github.com/theopenlane/core/v2/pkg/objects/storage/providers/r2"
	"github.com/theopenlane/core/v2/pkg/objects/storage/proxy"
	"github.com/theopenlane/iam/tokens"
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
				assert.Check(t, err != nil)
				assert.Check(t, is.Nil(provider))
			} else {
				if err != nil {
					t.Skip("Skipping test due to missing R2-compatible environment")
				}
				assert.Check(t, provider != nil)
				assert.Check(t, provider.Close())
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
		assert.Check(t, is.Equal(storagetypes.R2Provider, provider.ProviderType()))
	})

	t.Run("GetScheme", func(t *testing.T) {
		scheme := provider.GetScheme()
		assert.Check(t, scheme != nil)
		assert.Check(t, is.Equal("r2://", *scheme))
	})
}

func TestR2ProviderPresignedURL(t *testing.T) {
	provider, err := r2provider.NewR2Provider(r2Options())
	if err != nil {
		t.Skip("Skipping test due to missing R2-compatible environment")
	}
	defer provider.Close()

	file := &storagetypes.File{Key: "test-file.txt"}
	url, err := provider.GetPresignedURL(context.Background(), file, &storagetypes.PresignedURLOptions{Duration: time.Minute})
	if err != nil {
		t.Skip("Skipping presigned URL test without R2 environment")
	}
	assert.Check(t, strings.Contains(url, "test-file.txt"))
}

func TestNewR2ProviderFromCredentials(t *testing.T) {
	creds := storage.ProviderCredentials{AccountID: "account", AccessKeyID: "access", SecretAccessKey: "secret"}
	opts := storage.NewProviderOptions(storage.WithBucket("bucket"))

	result := r2provider.NewR2ProviderFromCredentials(creds, opts)
	if result.IsError() {
		t.Skip("Skipping test due to missing R2-compatible environment")
	}

	assert.Check(t, result.MustGet() != nil)
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
		assert.Check(t, err != nil)
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
			Key: "",
		}
		downloadOpts := &storagetypes.DownloadFileOptions{
			FileName: "test.txt",
		}
		_, err = provider.Download(context.Background(), file, downloadOpts)
		assert.Check(t, err != nil)
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
			Key: "",
		}
		err = provider.Delete(context.Background(), file, nil)
		assert.Check(t, err != nil)
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
			Key: "",
		}
		_, err = provider.Exists(context.Background(), file)
		assert.Check(t, err != nil)
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
		assert.Check(t, err != nil)
	})
}

const (
	minioImage  = "pgsty/silo:latest"
	minioUser   = "provider1"
	minioSecret = "provider1secret"
	minioBucket = "provider1-bucket"
	testRegion  = "auto"
)

func TestR2ProviderPresignURLProxyDisabled(t *testing.T) {
	ctx := context.Background()
	endpoint, terminate := startMinio(t, ctx)
	t.Cleanup(func() { _ = terminate(context.Background()) })

	createBucket(t, ctx, endpoint, minioBucket)

	builder := r2provider.NewR2Builder().WithOptions(r2provider.WithUsePathStyle(true))

	options := storage.NewProviderOptions(
		storage.WithBucket(minioBucket),
		storage.WithEndpoint(endpoint),
		storage.WithCredentials(storage.ProviderCredentials{
			AccountID:       "test-account",
			AccessKeyID:     minioUser,
			SecretAccessKey: minioSecret,
		}),
		storage.WithProxyPresignEnabled(false),
	)

	providerInterface, err := builder.Build(ctx, storage.ProviderCredentials{
		AccountID:       "test-account",
		AccessKeyID:     minioUser,
		SecretAccessKey: minioSecret,
	}, options)
	assert.NilError(t, err)
	provider := providerInterface
	t.Cleanup(func() { _ = provider.Close() })

	content := []byte("hello, minio from r2")
	_, err = provider.Upload(ctx, bytes.NewReader(content), &storagetypes.UploadFileOptions{
		FileName:    "folder/object.txt",
		ContentType: "text/plain",
	})
	assert.NilError(t, err)

	url, err := provider.GetPresignedURL(ctx, &storagetypes.File{
		Bucket: minioBucket,
		Key:    "folder/object.txt",
	}, &storagetypes.PresignedURLOptions{Duration: time.Minute})
	assert.NilError(t, err)
	assert.Assert(t, strings.HasPrefix(url, "http://"), "expected upstream R2/S3 URL, got %q", url)
	assert.Assert(t, is.Contains(url, "X-Amz-Algorithm=AWS4-HMAC-SHA256"), "URL should contain AWS signature")
	assert.Assert(t, is.Contains(url, "X-Amz-Credential="), "URL should contain AWS credentials")
	assert.Assert(t, is.Contains(url, "X-Amz-Signature="), "URL should contain AWS signature")
	assert.Assert(t, is.Contains(url, "folder/object.txt"), "URL should contain object key")

	resp, err := http.Get(url)
	assert.NilError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(http.StatusOK, resp.StatusCode))
	assert.Assert(t, is.Equal(string(content), string(body)))
}

func TestR2ProviderPresignURLProxyEnabled(t *testing.T) {
	ctx := context.Background()
	endpoint, terminate := startMinio(t, ctx)
	t.Cleanup(func() { _ = terminate(context.Background()) })

	createBucket(t, ctx, endpoint, minioBucket)

	tm := newTestTokenManager(t)
	baseURL := "http://localhost:17608"

	builder := r2provider.NewR2Builder().WithOptions(r2provider.WithUsePathStyle(true))

	options := storage.NewProviderOptions(
		storage.WithBucket(minioBucket),
		storage.WithEndpoint(endpoint),
		storage.WithCredentials(storage.ProviderCredentials{
			AccountID:       "test-account",
			AccessKeyID:     minioUser,
			SecretAccessKey: minioSecret,
		}),
		storage.WithProxyPresignEnabled(true),
		storage.WithProxyPresignConfig(&storage.ProxyPresignConfig{
			TokenManager: tm,
			BaseURL:      baseURL,
		}),
	)

	providerInterface, err := builder.Build(ctx, storage.ProviderCredentials{
		AccountID:       "test-account",
		AccessKeyID:     minioUser,
		SecretAccessKey: minioSecret,
	}, options)
	assert.NilError(t, err)
	provider := providerInterface
	t.Cleanup(func() { _ = provider.Close() })

	content := []byte("hello, minio from r2 with proxy enabled")
	_, err = provider.Upload(ctx, bytes.NewReader(content), &storagetypes.UploadFileOptions{
		FileName:    "folder/object.txt",
		ContentType: "text/plain",
	})
	assert.NilError(t, err)

	file := &storagetypes.File{
		ID:      "01HZZTESTFILER2BBBBBB",
		Bucket:  minioBucket,
		Key:     "folder/object.txt",
		FullURI: "r2://" + minioBucket + "/folder/object.txt",
	}

	secret := make([]byte, 128)
	_, err = rand.Read(secret)
	assert.NilError(t, err)

	url, err := proxy.GenerateDownloadURLWithSecret(file, secret, time.Minute, &storage.ProxyPresignConfig{
		TokenManager: tm,
		BaseURL:      baseURL,
	})
	assert.NilError(t, err, "with proxy presign enabled, should generate proxy URL")
	assert.Assert(t, len(url) > 0)
	assert.Assert(t, strings.HasPrefix(url, baseURL), "URL should start with BaseURL, got %q", url)
	assert.Assert(t, is.Contains(url, file.ID), "URL should contain file ID")
	assert.Assert(t, is.Contains(url, "token="), "URL should contain token parameter")
	assert.Assert(t, !strings.Contains(url, "X-Amz-Algorithm"), "proxy URL should not contain AWS signature parameters")
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
	assert.NilError(t, err)

	host, err := container.Host(ctx)
	assert.NilError(t, err)
	port, err := container.MappedPort(ctx, "9000/tcp")
	assert.NilError(t, err)

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
		if _, ok := errors.AsType[*types.BucketAlreadyOwnedByYou](err); !ok {
			assert.NilError(t, err)
		}
	}

	waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	waiter := awss3.NewBucketExistsWaiter(client)
	assert.NilError(t, waiter.Wait(waitCtx, &awss3.HeadBucketInput{Bucket: aws.String(bucket)}, 1*time.Second))
}

func newTestTokenManager(t *testing.T) *tokens.TokenManager {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	assert.NilError(t, err)
	conf := tokens.Config{
		Audience:        "http://localhost:17608",
		Issuer:          "http://localhost:17608",
		AccessDuration:  time.Hour,
		RefreshDuration: 2 * time.Hour,
		RefreshOverlap:  -15 * time.Minute,
	}
	tm, err := tokens.NewWithKey(key, conf)
	assert.NilError(t, err)
	return tm
}
