package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"

	"github.com/theopenlane/core/pkg/objects"
)

// ensure GCSStore satisfies the Storage interface
var _ objects.Storage = &GCSStore{}

// ProviderGCS is the provider for Google Cloud Storage
var ProviderGCS = "gcs"

// GCSStore is a store that uses GCS as the backend
// it exposes a limited subset of features required by objects.Storage interface
type GCSStore struct {
	// Client is the GCS client
	Client *storage.Client
	// Bucket is the GCS bucket handle
	Bucket *storage.BucketHandle
	// Opts are the options for the GCS store
	Opts *GCSOptions
	// Scheme is the storage backend
	Scheme string
}

// NewGCSFromConfig creates a new GCSStore from the provided configuration.
// Authentication is handled via the client options passed in or the
// GOOGLE_APPLICATION_CREDENTIALS environment variable. The provided context is
// used for client initialization
func NewGCSFromConfig(ctx context.Context, opts *GCSOptions) (*GCSStore, error) {
	client, err := storage.NewClient(ctx, opts.ClientOptions...)
	if err != nil {
		return nil, err
	}

	bucket := client.Bucket(opts.Bucket)

	return &GCSStore{
		Client: client,
		Bucket: bucket,
		Opts:   opts,
		Scheme: "gs://",
	}, nil
}

// Exists checks if an object exists in GCS
func (s *GCSStore) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.Bucket.Object(key).Attrs(ctx)

	if errors.Is(err, storage.ErrObjectNotExist) {
		return false, nil
	}

	return err == nil, err
}

// HeadObj returns the object attributes
func (s *GCSStore) HeadObj(ctx context.Context, key string) (*storage.ObjectAttrs, error) {
	return s.Bucket.Object(key).Attrs(ctx)
}

// Close closes the underlying client
func (s *GCSStore) Close() error { return s.Client.Close() }

// ManagerUpload uploads multiple files sequentially
func (s *GCSStore) ManagerUpload(ctx context.Context, files [][]byte) error {
	for i, f := range files {
		w := s.Bucket.Object("file" + toString(i)).NewWriter(ctx)

		if _, err := w.Write(f); err != nil {
			_ = w.Close()
			return err
		}

		if err := w.Close(); err != nil {
			return err
		}
	}

	return nil
}

func toString(i int) string { return fmt.Sprintf("%d", i) }

// Upload uploads a single file
func (s *GCSStore) Upload(ctx context.Context, r io.Reader, opts *objects.UploadFileOptions) (*objects.UploadedFileMetadata, error) {
	w := s.Bucket.Object(opts.FileName).NewWriter(ctx)

	w.ContentType = opts.ContentType
	w.Metadata = opts.Metadata

	n, err := io.Copy(w, r)
	if err != nil {
		_ = w.Close()
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return &objects.UploadedFileMetadata{
		FolderDestination: s.Opts.Bucket,
		Size:              n,
		Key:               opts.FileName,
	}, nil
}

// Download retrieves an object
func (s *GCSStore) Download(ctx context.Context, opts *objects.DownloadFileOptions) (*objects.DownloadFileMetadata, error) {
	r, err := s.Bucket.Object(opts.FileName).NewReader(ctx)
	if err != nil {
		return nil, err
	}

	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return &objects.DownloadFileMetadata{File: data, Size: int64(len(data))}, nil
}

// GetPresignedURL returns a signed URL that expires after the given duration
func (s *GCSStore) GetPresignedURL(key string, expires time.Duration) (string, error) {
	if expires == 0 {
		expires = 15 * time.Minute // nolint: mnd
	}

	u, err := storage.SignedURL(s.Opts.Bucket, key, &storage.SignedURLOptions{
		GoogleAccessID: s.Opts.GoogleAccessID,
		PrivateKey:     s.Opts.PrivateKey,
		Method:         "GET",
		Expires:        time.Now().Add(expires),
		Scheme:         storage.SigningSchemeV4,
	})
	if err != nil {
		return "", err
	}

	log.Debug().Str("presigned_url", u).Msg("gcs presigned URL created")

	return u, nil
}

// GetScheme returns the scheme of the storage backend
func (s *GCSStore) GetScheme() *string { return &s.Scheme }

// Delete deletes an object
func (s *GCSStore) Delete(ctx context.Context, key string) error {
	return s.Bucket.Object(key).Delete(ctx)
}

// Tag updates an object's metadata
func (s *GCSStore) Tag(ctx context.Context, key string, tags map[string]string) error {
	_, err := s.Bucket.Object(key).Update(ctx, storage.ObjectAttrsToUpdate{Metadata: tags})

	return err
}

// GetTags retrieves object metadata
func (s *GCSStore) GetTags(ctx context.Context, key string) (map[string]string, error) {
	attrs, err := s.Bucket.Object(key).Attrs(ctx)
	if err != nil {
		return nil, err
	}

	return attrs.Metadata, nil
}

// ListBuckets lists buckets within the configured project
func (s *GCSStore) ListBuckets() ([]string, error) {
	ctx := context.Background()

	it := s.Client.Buckets(ctx, s.Opts.ProjectID)
	var buckets []string

	for {
		b, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, err
		}

		buckets = append(buckets, b.Name)
	}

	return buckets, nil
}
