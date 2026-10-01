package gcs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"time"

	gstorage "cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/metrics"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

const (
	// DefaultPresignedURLExpiry defines the default expiry time for signed URLs
	DefaultPresignedURLExpiry = 15 * time.Minute
	// scheme is the URI scheme for Google Cloud Storage objects
	scheme = "gs://"
)

// Provider implements storagetypes.Provider against Google Cloud Storage
type Provider struct {
	client    *gstorage.Client
	bucket    string
	projectID string
	options   *storage.ProviderOptions
}

// providerConfig holds configuration for the GCS provider
type providerConfig struct {
	clientOptions []option.ClientOption
}

// Option configures the GCS provider during construction
type Option func(*providerConfig)

// WithClientOptions supplies Google API client options such as a credential token source; omitted, the client uses application default credentials
func WithClientOptions(opts ...option.ClientOption) Option {
	return func(cfg *providerConfig) {
		cfg.clientOptions = append(cfg.clientOptions, opts...)
	}
}

// NewProvider creates a GCS provider for the bucket in options, authenticating with application default credentials unless client options say otherwise
func NewProvider(ctx context.Context, options *storage.ProviderOptions, opts ...Option) (*Provider, error) {
	if options == nil || options.Bucket == "" {
		return nil, ErrBucketRequired
	}

	cfg := providerConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	if options.Endpoint != "" {
		cfg.clientOptions = append(cfg.clientOptions, option.WithEndpoint(options.Endpoint))
	}

	client, err := gstorage.NewClient(ctx, cfg.clientOptions...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientCreate, err)
	}

	extra, _ := options.Extra(storage.GCSProjectIDExtraKey)
	projectID, _ := extra.(string)

	return &Provider{
		client:    client,
		bucket:    options.Bucket,
		projectID: projectID,
		options:   options.Clone(),
	}, nil
}

// ProviderType implements storagetypes.Provider
func (p *Provider) ProviderType() storagetypes.ProviderType {
	return storage.GCSProvider
}

// Upload implements storagetypes.Provider
func (p *Provider) Upload(ctx context.Context, reader io.Reader, opts *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
	objectKey := opts.FileName
	if opts.FolderDestination != "" {
		objectKey = path.Join(opts.FolderDestination, opts.FileName)
	}

	writeCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	w := p.client.Bucket(p.bucket).Object(objectKey).NewWriter(writeCtx)
	w.ContentType = opts.ContentType

	size, err := io.Copy(w, reader)
	if err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	metrics.RecordStorageUpload(string(storage.GCSProvider), size)

	return &storagetypes.UploadedFileMetadata{
		FileMetadata: storagetypes.FileMetadata{
			Key:          objectKey,
			Size:         size,
			Folder:       opts.FolderDestination,
			Bucket:       p.bucket,
			ContentType:  opts.ContentType,
			ProviderType: storage.GCSProvider,
			FullURI:      scheme + p.bucket + "/" + objectKey,
		},
	}, nil
}

// Download implements storagetypes.Provider
func (p *Provider) Download(ctx context.Context, file *storagetypes.File, _ *storagetypes.DownloadFileOptions) (*storagetypes.DownloadedFileMetadata, error) {
	r, err := p.client.Bucket(p.bucketFor(file)).Object(file.Key).NewReader(ctx)
	if err != nil {
		return nil, err
	}

	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	size := int64(len(data))

	metrics.RecordStorageDownload(string(storage.GCSProvider), size)

	return &storagetypes.DownloadedFileMetadata{
		File: data,
		Size: size,
	}, nil
}

// Delete implements storagetypes.Provider
func (p *Provider) Delete(ctx context.Context, file *storagetypes.File, _ *storagetypes.DeleteFileOptions) error {
	err := p.client.Bucket(p.bucketFor(file)).Object(file.Key).Delete(ctx)
	if err != nil && !errors.Is(err, gstorage.ErrObjectNotExist) {
		return err
	}

	metrics.RecordStorageDelete(string(storage.GCSProvider))

	return nil
}

// GetPresignedURL implements storagetypes.Provider
func (p *Provider) GetPresignedURL(_ context.Context, file *storagetypes.File, opts *storagetypes.PresignedURLOptions) (string, error) {
	if opts == nil {
		opts = &storagetypes.PresignedURLOptions{}
	}

	duration := opts.Duration
	if duration == 0 {
		duration = DefaultPresignedURLExpiry
	}

	return p.client.Bucket(p.bucketFor(file)).SignedURL(file.Key, &gstorage.SignedURLOptions{
		Method:  http.MethodGet,
		Expires: time.Now().Add(duration),
		Scheme:  gstorage.SigningSchemeV4,
	})
}

// Exists implements storagetypes.Provider
func (p *Provider) Exists(ctx context.Context, file *storagetypes.File) (bool, error) {
	_, err := p.client.Bucket(p.bucketFor(file)).Object(file.Key).Attrs(ctx)

	switch {
	case errors.Is(err, gstorage.ErrObjectNotExist):
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

// GetScheme implements storagetypes.Provider
func (p *Provider) GetScheme() *string {
	s := scheme

	return &s
}

// ListBuckets implements storagetypes.Provider
func (p *Provider) ListBuckets() ([]string, error) {
	if p.projectID == "" {
		return nil, ErrProjectIDRequired
	}

	var buckets []string

	it := p.client.Buckets(context.Background(), p.projectID)

	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			return buckets, nil
		}

		if err != nil {
			return nil, err
		}

		buckets = append(buckets, attrs.Name)
	}
}

// Bucket returns the bucket the provider reads and writes by default
func (p *Provider) Bucket() string {
	return p.bucket
}

// ListObjects returns the object keys under prefix, at most limit when limit is positive
func (p *Provider) ListObjects(ctx context.Context, prefix string, limit int) ([]string, error) {
	var keys []string

	it := p.client.Bucket(p.bucket).Objects(ctx, &gstorage.Query{Prefix: prefix})

	for limit <= 0 || len(keys) < limit {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, err
		}

		keys = append(keys, attrs.Name)
	}

	return keys, nil
}

// Close implements storagetypes.Provider
func (p *Provider) Close() error {
	return p.client.Close()
}

// bucketFor returns the bucket recorded on the file, falling back to the provider bucket
func (p *Provider) bucketFor(file *storagetypes.File) string {
	if file.Bucket != "" {
		return file.Bucket
	}

	return p.bucket
}
