package gcs_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	gstorage "cloud.google.com/go/storage"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"google.golang.org/api/option"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

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
	assert.NilError(t, err)
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
	assert.NilError(t, err)
	t.Cleanup(func() { _ = provider.Close() })

	return provider
}

func serviceAccountJSON(t *testing.T) []byte {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
	assert.NilError(t, err)

	pemKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	sa, err := json.Marshal(map[string]string{
		"type":         "service_account",
		"project_id":   testProject,
		"client_email": "signer@" + testProject + ".iam.gserviceaccount.com",
		"private_key":  string(pemKey),
		"token_uri":    "https://oauth2.googleapis.com/token",
	})
	assert.NilError(t, err)

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
			assert.Check(t, is.ErrorIs(err, tt.expectErr))
			assert.Check(t, is.Nil(provider))
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
			assert.NilError(t, err)
			assert.Assert(t, provider != nil)

			exists, err := provider.Exists(context.Background(), &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "seeded.txt"}})
			assert.NilError(t, err)
			assert.Check(t, exists)
			assert.Check(t, provider.Close())
		})
	}
}

func TestProviderMethods(t *testing.T) {
	provider := newFakeProvider(t, newFakeServer(t))

	t.Run("ProviderType", func(t *testing.T) {
		assert.Check(t, is.Equal(storage.GCSProvider, provider.ProviderType()))
	})

	t.Run("GetScheme", func(t *testing.T) {
		scheme := provider.GetScheme()
		assert.Check(t, scheme != nil)
		assert.Check(t, is.Equal("gs://", *scheme))
	})

	t.Run("Bucket", func(t *testing.T) {
		assert.Check(t, is.Equal(testBucket, provider.Bucket()))
	})

	t.Run("Close", func(t *testing.T) {
		assert.Check(t, provider.Close())
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
			assert.NilError(t, err)
			assert.Assert(t, uploaded != nil)

			assert.Check(t, is.Equal(tt.expectedKey, uploaded.Key))
			assert.Check(t, is.Equal(int64(len(tt.content)), uploaded.Size))
			assert.Check(t, is.Equal(tt.uploadOpts.FolderDestination, uploaded.Folder))
			assert.Check(t, is.Equal(testBucket, uploaded.Bucket))
			assert.Check(t, is.Equal(tt.uploadOpts.ContentType, uploaded.ContentType))
			assert.Check(t, is.Equal(storage.GCSProvider, uploaded.ProviderType))
			assert.Check(t, is.Equal("gs://"+testBucket+"/"+tt.expectedKey, uploaded.FullURI))

			object, err := server.GetObject(testBucket, tt.expectedKey)
			assert.NilError(t, err)
			assert.Check(t, is.Equal(tt.content, string(object.Content)))
			assert.Check(t, is.Equal(tt.uploadOpts.ContentType, object.ContentType))
		})
	}
}

func TestProviderUploadReaderFailure(t *testing.T) {
	provider := newFakeProvider(t, newFakeServer(t))

	uploaded, err := provider.Upload(context.Background(), iotest.ErrReader(errReadFailed), &storagetypes.UploadFileOptions{FileName: "broken.txt"})
	assert.Check(t, is.ErrorIs(err, errReadFailed))
	assert.Check(t, is.Nil(uploaded))

	exists, err := provider.Exists(context.Background(), &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "broken.txt"}})
	assert.NilError(t, err)
	assert.Check(t, !exists)
}

func TestProviderUploadMissingBucket(t *testing.T) {
	provider := newFakeProvider(t, newFakeServer(t), storage.WithBucket("missing-bucket"))

	uploaded, err := provider.Upload(context.Background(), strings.NewReader("content"), &storagetypes.UploadFileOptions{FileName: "file.txt"})
	assert.Check(t, err != nil)
	assert.Check(t, is.Nil(uploaded))
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
				assert.Check(t, err != nil)
				assert.Check(t, is.Nil(downloaded))

				if tt.expectErr != nil {
					assert.Check(t, is.ErrorIs(err, tt.expectErr))
				}

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, downloaded != nil)
			assert.Check(t, is.Equal(tt.expected, string(downloaded.File)))
			assert.Check(t, is.Equal(int64(len(tt.expected)), downloaded.Size))
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
				assert.Check(t, err != nil)

				return
			}

			assert.NilError(t, err)

			exists, err := provider.Exists(context.Background(), tt.file)
			assert.NilError(t, err)
			assert.Check(t, !exists)
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
				assert.Check(t, err != nil)
				assert.Check(t, !exists)

				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.Equal(tt.expected, exists))
		})
	}
}

func TestProviderListBuckets(t *testing.T) {
	t.Run("with project id", func(t *testing.T) {
		provider := newFakeProvider(t, newFakeServer(t))

		buckets, err := provider.ListBuckets()
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual(slices.Sorted(slices.Values([]string{testBucket, otherBucket})), slices.Sorted(slices.Values(buckets))))
	})

	t.Run("without project id", func(t *testing.T) {
		server := newFakeServer(t)
		provider, err := gcs.NewProvider(context.Background(), storage.NewProviderOptions(storage.WithBucket(testBucket)), fakeClientOptions(server, option.WithoutAuthentication()))
		assert.NilError(t, err)
		t.Cleanup(func() { _ = provider.Close() })

		buckets, err := provider.ListBuckets()
		assert.Check(t, is.ErrorIs(err, gcs.ErrProjectIDRequired))
		assert.Check(t, is.Nil(buckets))
	})

	t.Run("stopped server", func(t *testing.T) {
		server := newFakeServer(t)
		provider := newFakeProvider(t, server)
		server.Stop()

		buckets, err := provider.ListBuckets()
		assert.Check(t, err != nil)
		assert.Check(t, is.Nil(buckets))
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
				assert.Check(t, err != nil)
				assert.Check(t, is.Nil(keys))

				return
			}

			assert.NilError(t, err)

			if tt.expected != nil {
				assert.Check(t, is.DeepEqual(slices.Sorted(slices.Values(tt.expected)), slices.Sorted(slices.Values(keys))))
			} else {
				assert.Check(t, is.Len(keys, tt.expectedLen))
			}
		})
	}
}

func TestProviderGetPresignedURL(t *testing.T) {
	server := newFakeServer(t)

	t.Run("signing credentials", func(t *testing.T) {
		provider, err := gcs.NewProvider(context.Background(), gcsOptions(), fakeClientOptions(server, option.WithCredentialsJSON(serviceAccountJSON(t))))
		assert.NilError(t, err)
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
				assert.NilError(t, err)
				assert.Check(t, strings.HasPrefix(signed, "https://storage.googleapis.com/"+tt.expectedBucket+"/folder/object.txt?"), "unexpected URL %q", signed)

				parsed, err := url.Parse(signed)
				assert.NilError(t, err)

				query := parsed.Query()
				assert.Check(t, is.Equal("GOOG4-RSA-SHA256", query.Get("X-Goog-Algorithm")))
				assert.Check(t, len(query.Get("X-Goog-Signature")) > 0)

				expires, err := strconv.Atoi(query.Get("X-Goog-Expires"))
				assert.NilError(t, err)
				assert.Check(t, math.Abs(tt.expectedExpiry.Seconds()-float64(expires)) <= 1)
			})
		}
	})

	t.Run("without signing credentials", func(t *testing.T) {
		provider := newFakeProvider(t, server)

		signed, err := provider.GetPresignedURL(context.Background(), &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: "object.txt"}}, nil)
		assert.Check(t, err != nil)
		assert.Check(t, is.Len(signed, 0))
	})
}
