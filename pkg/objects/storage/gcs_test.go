package storage

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/objects"
)

func TestGCSStoreUploadDownload(t *testing.T) {
	srv := fakestorage.NewServer(nil)
	srv.CreateBucketWithOpts(fakestorage.CreateBucketOpts{Name: "test-bucket"})
	defer srv.Stop()

	ctx := context.Background()
	opts := NewGCSOptions(
		WithGCSBucket("test-bucket"),
		WithGCSProjectID("proj"),
	)

	client := srv.Client()
	store := &GCSStore{Client: client, Bucket: client.Bucket(opts.Bucket), Opts: opts, Scheme: "gs://"}

	// Upload
	_, err := store.Upload(ctx, io.NopCloser(strings.NewReader("hello")), &objects.UploadFileOptions{FileName: "foo.txt"})
	require.NoError(t, err)

	// Download
	meta, err := store.Download(ctx, &objects.DownloadFileOptions{Bucket: opts.Bucket, FileName: "foo.txt"})
	require.NoError(t, err)
	assert.Equal(t, int64(5), meta.Size)
	assert.Equal(t, []byte("hello"), meta.File)
}

func TestGCSStorePresignedURL(t *testing.T) {
	srv := fakestorage.NewServer(nil)
	srv.CreateBucketWithOpts(fakestorage.CreateBucketOpts{Name: "test-bucket"})
	defer srv.Stop()

	key := []byte("dummy")
	opts := NewGCSOptions(
		WithGCSBucket("test-bucket"),
		WithGCSGoogleAccessID("test@test.com"),
		WithGCSPrivateKey(key),
	)

	client := srv.Client()
	store := &GCSStore{Client: client, Bucket: client.Bucket(opts.Bucket), Opts: opts, Scheme: "gs://"}

	_, err := store.GetPresignedURL("somekey", time.Minute)
	require.Error(t, err)
}

func TestGCSStoreDelete(t *testing.T) {
	srv := fakestorage.NewServer(nil)
	srv.CreateBucketWithOpts(fakestorage.CreateBucketOpts{Name: "test-bucket"})
	defer srv.Stop()

	ctx := context.Background()
	opts := NewGCSOptions(
		WithGCSBucket("test-bucket"),
		WithGCSProjectID("proj"),
	)

	client := srv.Client()

	store := &GCSStore{Client: client, Bucket: client.Bucket(opts.Bucket), Opts: opts, Scheme: "gs://"}

	_, err := store.Upload(ctx, io.NopCloser(strings.NewReader("bye")), &objects.UploadFileOptions{FileName: "del.txt"})
	require.NoError(t, err)

	err = store.Delete(ctx, "del.txt")
	require.NoError(t, err)

	exists, err := store.Exists(ctx, "del.txt")
	assert.False(t, exists)
}
