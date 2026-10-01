package gcs_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	gstorage "cloud.google.com/go/storage"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/core/v2/pkg/objects/storage/providers/gcs"
)

const (
	testBucket  = "test-bucket"
	otherBucket = "other-bucket"
	testProject = "test-project"
	rsaKeyBits  = 2048
)

var errReadFailed = errors.New("read failed")

func newFakeServer(t *testing.T, objects ...fakestorage.Object) *fakestorage.Server {
	t.Helper()

	server, err := fakestorage.NewServerWithOptions(fakestorage.Options{NoListener: true, InitialObjects: objects})
	require.NoError(t, err)
	t.Cleanup(server.Stop)

	server.CreateBucketWithOpts(fakestorage.CreateBucketOpts{Name: testBucket})
	server.CreateBucketWithOpts(fakestorage.CreateBucketOpts{Name: otherBucket})

	return server
}

func fakeObject(bucket, name, content string) fakestorage.Object {
	return fakestorage.Object{
		ObjectAttrs: fakestorage.ObjectAttrs{BucketName: bucket, Name: name, ContentType: "text/plain"},
		Content:     []byte(content),
	}
}

func fakeClientOptions(server *fakestorage.Server, auth option.ClientOption) gcs.Option {
	return gcs.WithClientOptions(option.WithHTTPClient(server.HTTPClient()), auth)
}

func gcsOptions(opts ...storage.ProviderOption) *storage.ProviderOptions {
	return storage.NewProviderOptions(append([]storage.ProviderOption{
		storage.WithBucket(testBucket),
		storage.WithExtra(storage.GCSProjectIDExtraKey, testProject),
	}, opts...)...)
}

func newFakeProvider(t *testing.T, server *fakestorage.Server, opts ...storage.ProviderOption) *gcs.Provider {
	t.Helper()

	provider, err := gcs.NewProvider(context.Background(), gcsOptions(opts...), fakeClientOptions(server, option.WithoutAuthentication()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close() })

	return provider
}

func serviceAccountJSON(t *testing.T) []byte {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
	require.NoError(t, err)

	pemKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	sa, err := json.Marshal(map[string]string{
		"type":         "service_account",
		"project_id":   testProject,
		"client_email": "signer@" + testProject + ".iam.gserviceaccount.com",
		"private_key":  string(pemKey),
		"token_uri":    "https://oauth2.googleapis.com/token",
	})
	require.NoError(t, err)

	return sa
}

func TestNewProviderValidation(t *testing.T) {
	tests := []struct {
		name      string
		options   *storage.ProviderOptions
		opts      []gcs.Option
		expectErr error
	}{
		{
			name:      "nil options",
			options:   nil,
			expectErr: gcs.ErrBucketRequired,
		},
		{
			name:      "empty bucket",
			options:   storage.NewProviderOptions(storage.WithBucket("")),
			expectErr: gcs.ErrBucketRequired,
		},
		{
			name:      "conflicting client options",
			options:   gcsOptions(),
			opts:      []gcs.Option{gcs.WithClientOptions(option.WithoutAuthentication(), option.WithCredentialsJSON([]byte("{}")))},
			expectErr: gcs.ErrClientCreate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := gcs.NewProvider(context.Background(), tt.options, tt.opts...)
			assert.ErrorIs(t, err, tt.expectErr)
			assert.Nil(t, provider)
		})
	}
}

func TestNewProvider(t *testing.T) {
	server := newFakeServer(t, fakeObject(testBucket, "seeded.txt", "seeded"))

	tests := []struct {
		name    string
		options *storage.ProviderOptions
	}{
		{
			name:    "valid configuration",
			options: gcsOptions(),
		},
		{
			name:    "endpoint override",
			options: gcsOptions(storage.WithEndpoint("http://gcs.local/storage/v1/")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := gcs.NewProvider(context.Background(), tt.options, fakeClientOptions(server, option.WithoutAuthentication()))
			require.NoError(t, err)
			require.NotNil(t, provider)

			exists, err := provider.Exists(context.Background(), &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "seeded.txt"}})
			require.NoError(t, err)
			assert.True(t, exists)
			assert.NoError(t, provider.Close())
		})
	}
}

func TestProviderMethods(t *testing.T) {
	provider := newFakeProvider(t, newFakeServer(t))

	t.Run("ProviderType", func(t *testing.T) {
		assert.Equal(t, storage.GCSProvider, provider.ProviderType())
	})

	t.Run("GetScheme", func(t *testing.T) {
		scheme := provider.GetScheme()
		assert.NotNil(t, scheme)
		assert.Equal(t, "gs://", *scheme)
	})

	t.Run("Bucket", func(t *testing.T) {
		assert.Equal(t, testBucket, provider.Bucket())
	})

	t.Run("Close", func(t *testing.T) {
		assert.NoError(t, provider.Close())
	})
}

func TestProviderUpload(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		uploadOpts  *storagetypes.UploadFileOptions
		expectedKey string
	}{
		{
			name:        "root object",
			content:     "hello, gcs",
			uploadOpts:  &storagetypes.UploadFileOptions{FileName: "file.txt", ContentType: "text/plain"},
			expectedKey: "file.txt",
		},
		{
			name:        "folder destination",
			content:     "hello, folder",
			uploadOpts:  &storagetypes.UploadFileOptions{FileName: "file.txt", FolderDestination: "org/parent", ContentType: "application/json"},
			expectedKey: "org/parent/file.txt",
		},
		{
			name:        "empty content",
			content:     "",
			uploadOpts:  &storagetypes.UploadFileOptions{FileName: "empty.txt", ContentType: "text/plain"},
			expectedKey: "empty.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newFakeServer(t)
			provider := newFakeProvider(t, server)

			uploaded, err := provider.Upload(context.Background(), strings.NewReader(tt.content), tt.uploadOpts)
			require.NoError(t, err)
			require.NotNil(t, uploaded)

			assert.Equal(t, tt.expectedKey, uploaded.Key)
			assert.Equal(t, int64(len(tt.content)), uploaded.Size)
			assert.Equal(t, tt.uploadOpts.FolderDestination, uploaded.Folder)
			assert.Equal(t, testBucket, uploaded.Bucket)
			assert.Equal(t, tt.uploadOpts.ContentType, uploaded.ContentType)
			assert.Equal(t, storage.GCSProvider, uploaded.ProviderType)
			assert.Equal(t, "gs://"+testBucket+"/"+tt.expectedKey, uploaded.FullURI)

			object, err := server.GetObject(testBucket, tt.expectedKey)
			require.NoError(t, err)
			assert.Equal(t, []byte(tt.content), object.Content)
			assert.Equal(t, tt.uploadOpts.ContentType, object.ContentType)
		})
	}
}

func TestProviderUploadReaderFailure(t *testing.T) {
	provider := newFakeProvider(t, newFakeServer(t))

	uploaded, err := provider.Upload(context.Background(), iotest.ErrReader(errReadFailed), &storagetypes.UploadFileOptions{FileName: "broken.txt"})
	assert.ErrorIs(t, err, errReadFailed)
	assert.Nil(t, uploaded)

	exists, err := provider.Exists(context.Background(), &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "broken.txt"}})
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestProviderUploadMissingBucket(t *testing.T) {
	provider := newFakeProvider(t, newFakeServer(t), storage.WithBucket("missing-bucket"))

	uploaded, err := provider.Upload(context.Background(), strings.NewReader("content"), &storagetypes.UploadFileOptions{FileName: "file.txt"})
	assert.Error(t, err)
	assert.Nil(t, uploaded)
}

func TestProviderDownload(t *testing.T) {
	tests := []struct {
		name      string
		file      *storagetypes.File
		stopped   bool
		expected  string
		expectErr error
	}{
		{
			name:     "provider bucket",
			file:     &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "file.txt"}},
			expected: "default bucket content",
		},
		{
			name:     "file bucket",
			file:     &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "file.txt", Bucket: otherBucket}},
			expected: "other bucket content",
		},
		{
			name:      "missing object",
			file:      &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "missing.txt"}},
			expectErr: gstorage.ErrObjectNotExist,
		},
		{
			name:    "stopped server",
			file:    &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "file.txt"}},
			stopped: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newFakeServer(t,
				fakeObject(testBucket, "file.txt", "default bucket content"),
				fakeObject(otherBucket, "file.txt", "other bucket content"),
			)
			provider := newFakeProvider(t, server)

			if tt.stopped {
				server.Stop()
			}

			downloaded, err := provider.Download(context.Background(), tt.file, &storagetypes.DownloadFileOptions{})
			if tt.stopped || tt.expectErr != nil {
				assert.Error(t, err)
				assert.Nil(t, downloaded)

				if tt.expectErr != nil {
					assert.ErrorIs(t, err, tt.expectErr)
				}

				return
			}

			require.NoError(t, err)
			require.NotNil(t, downloaded)
			assert.Equal(t, []byte(tt.expected), downloaded.File)
			assert.Equal(t, int64(len(tt.expected)), downloaded.Size)
		})
	}
}

func TestProviderDelete(t *testing.T) {
	tests := []struct {
		name    string
		file    *storagetypes.File
		stopped bool
	}{
		{
			name: "existing object",
			file: &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "file.txt"}},
		},
		{
			name: "file bucket",
			file: &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "file.txt", Bucket: otherBucket}},
		},
		{
			name: "missing object",
			file: &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "missing.txt"}},
		},
		{
			name:    "stopped server",
			file:    &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "file.txt"}},
			stopped: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newFakeServer(t,
				fakeObject(testBucket, "file.txt", "default bucket content"),
				fakeObject(otherBucket, "file.txt", "other bucket content"),
			)
			provider := newFakeProvider(t, server)

			if tt.stopped {
				server.Stop()
			}

			err := provider.Delete(context.Background(), tt.file, &storagetypes.DeleteFileOptions{})
			if tt.stopped {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)

			exists, err := provider.Exists(context.Background(), tt.file)
			require.NoError(t, err)
			assert.False(t, exists)
		})
	}
}

func TestProviderExists(t *testing.T) {
	tests := []struct {
		name     string
		file     *storagetypes.File
		stopped  bool
		expected bool
	}{
		{
			name:     "existing object",
			file:     &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "file.txt"}},
			expected: true,
		},
		{
			name:     "missing object",
			file:     &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "missing.txt"}},
			expected: false,
		},
		{
			name:     "missing bucket",
			file:     &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "file.txt", Bucket: "missing-bucket"}},
			expected: false,
		},
		{
			name:    "stopped server",
			file:    &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "file.txt"}},
			stopped: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newFakeServer(t, fakeObject(testBucket, "file.txt", "content"))
			provider := newFakeProvider(t, server)

			if tt.stopped {
				server.Stop()
			}

			exists, err := provider.Exists(context.Background(), tt.file)
			if tt.stopped {
				assert.Error(t, err)
				assert.False(t, exists)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, exists)
		})
	}
}

func TestProviderListBuckets(t *testing.T) {
	t.Run("with project id", func(t *testing.T) {
		provider := newFakeProvider(t, newFakeServer(t))

		buckets, err := provider.ListBuckets()
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{testBucket, otherBucket}, buckets)
	})

	t.Run("without project id", func(t *testing.T) {
		server := newFakeServer(t)
		provider, err := gcs.NewProvider(context.Background(), storage.NewProviderOptions(storage.WithBucket(testBucket)), fakeClientOptions(server, option.WithoutAuthentication()))
		require.NoError(t, err)
		t.Cleanup(func() { _ = provider.Close() })

		buckets, err := provider.ListBuckets()
		assert.ErrorIs(t, err, gcs.ErrProjectIDRequired)
		assert.Nil(t, buckets)
	})

	t.Run("stopped server", func(t *testing.T) {
		server := newFakeServer(t)
		provider := newFakeProvider(t, server)
		server.Stop()

		buckets, err := provider.ListBuckets()
		assert.Error(t, err)
		assert.Nil(t, buckets)
	})
}

func TestProviderListObjects(t *testing.T) {
	tests := []struct {
		name        string
		prefix      string
		limit       int
		stopped     bool
		expected    []string
		expectedLen int
	}{
		{
			name:     "prefix filter",
			prefix:   "imports/",
			expected: []string{"imports/a.json", "imports/b.json"},
		},
		{
			name:        "limit",
			limit:       2,
			expectedLen: 2,
		},
		{
			name:     "no matches",
			prefix:   "missing/",
			expected: []string{},
		},
		{
			name:    "stopped server",
			stopped: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newFakeServer(t,
				fakeObject(testBucket, "imports/a.json", "{}"),
				fakeObject(testBucket, "imports/b.json", "{}"),
				fakeObject(testBucket, "exports/c.json", "{}"),
			)
			provider := newFakeProvider(t, server)

			if tt.stopped {
				server.Stop()
			}

			keys, err := provider.ListObjects(context.Background(), tt.prefix, tt.limit)
			if tt.stopped {
				assert.Error(t, err)
				assert.Nil(t, keys)

				return
			}

			require.NoError(t, err)

			if tt.expected != nil {
				assert.ElementsMatch(t, tt.expected, keys)
			} else {
				assert.Len(t, keys, tt.expectedLen)
			}
		})
	}
}

func TestProviderGetPresignedURL(t *testing.T) {
	server := newFakeServer(t)

	t.Run("signing credentials", func(t *testing.T) {
		provider, err := gcs.NewProvider(context.Background(), gcsOptions(), fakeClientOptions(server, option.WithCredentialsJSON(serviceAccountJSON(t))))
		require.NoError(t, err)
		t.Cleanup(func() { _ = provider.Close() })

		tests := []struct {
			name           string
			file           *storagetypes.File
			opts           *storagetypes.PresignedURLOptions
			expectedBucket string
			expectedExpiry time.Duration
		}{
			{
				name:           "explicit duration",
				file:           &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "folder/object.txt"}},
				opts:           &storagetypes.PresignedURLOptions{Duration: time.Minute},
				expectedBucket: testBucket,
				expectedExpiry: time.Minute,
			},
			{
				name:           "default duration",
				file:           &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "folder/object.txt", Bucket: otherBucket}},
				expectedBucket: otherBucket,
				expectedExpiry: gcs.DefaultPresignedURLExpiry,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				signed, err := provider.GetPresignedURL(context.Background(), tt.file, tt.opts)
				require.NoError(t, err)
				assert.True(t, strings.HasPrefix(signed, "https://storage.googleapis.com/"+tt.expectedBucket+"/folder/object.txt?"), "unexpected URL %q", signed)

				parsed, err := url.Parse(signed)
				require.NoError(t, err)

				query := parsed.Query()
				assert.Equal(t, "GOOG4-RSA-SHA256", query.Get("X-Goog-Algorithm"))
				assert.NotEmpty(t, query.Get("X-Goog-Signature"))

				expires, err := strconv.Atoi(query.Get("X-Goog-Expires"))
				require.NoError(t, err)
				assert.InDelta(t, tt.expectedExpiry.Seconds(), expires, 1)
			})
		}
	})

	t.Run("without signing credentials", func(t *testing.T) {
		provider := newFakeProvider(t, server)

		signed, err := provider.GetPresignedURL(context.Background(), &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "object.txt"}}, nil)
		assert.Error(t, err)
		assert.Empty(t, signed)
	})
}
