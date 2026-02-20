package r2

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/pkg/metrics"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/objects/storage/proxy"
)

const (
	// DefaultPresignedURLExpiry defines the default expiry time for presigned URLs
	DefaultPresignedURLExpiry = 15 * time.Minute
	// DefaultPartSize defines the default part size for multipart uploads (64MB)
	DefaultPartSize = 64 * 1024 * 1024
	// DefaultConcurrency defines the default concurrency for uploads
	DefaultConcurrency = 5
)

// Provider implements the storagetypes.Provider interface for Cloudflare R2
type Provider struct {
	client              *s3.Client
	options             *storage.ProviderOptions
	presignClient       *s3.PresignClient
	downloader          *transfermanager.Client
	objExistsWaiter     *s3.ObjectExistsWaiter
	objNotExistsWaiter  *s3.ObjectNotExistsWaiter
	proxyPresignEnabled bool
	proxyConfig         *storage.ProxyPresignConfig
}

// providerConfig holds configuration for the R2 provider
type providerConfig struct {
	options      *storage.ProviderOptions
	usePathStyle bool
}

// Option configures the R2 provider during construction
type Option func(*providerConfig)

// WithUsePathStyle configures the R2 client to use path-style addressing
func WithUsePathStyle(use bool) Option {
	return func(cfg *providerConfig) {
		cfg.usePathStyle = use
	}
}

// NewR2Provider creates a new R2 provider instance
func NewR2Provider(options *storage.ProviderOptions, opts ...Option) (*Provider, error) {
	config := providerConfig{
		options: options,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&config)
		}
	}

	if config.options == nil {
		return nil, ErrR2BucketRequired
	}

	if config.options.Bucket == "" {
		return nil, ErrR2BucketRequired
	}

	if config.options.Credentials.AccountID == "" && config.options.Endpoint == "" {
		return nil, ErrR2AccountIDRequired
	}

	if config.options.Credentials.AccessKeyID == "" || config.options.Credentials.SecretAccessKey == "" {
		return nil, ErrR2CredentialsMissing
	}

	endpoint := config.options.Endpoint
	if endpoint == "" && config.options.Credentials.AccountID != "" {
		endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", config.options.Credentials.AccountID)
	}

	creds := credentials.NewStaticCredentialsProvider(config.options.Credentials.AccessKeyID, config.options.Credentials.SecretAccessKey, "")

	awsConfig := aws.Config{
		Region:       "auto",
		Credentials:  creds,
		BaseEndpoint: aws.String(endpoint),
	}

	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.Region = "auto"
		o.Credentials = creds
		o.UsePathStyle = config.usePathStyle
	})

	return &Provider{
		client:              client,
		options:             config.options.Clone(),
		downloader:          transfermanager.New(client),
		presignClient:       s3.NewPresignClient(client),
		objExistsWaiter:     s3.NewObjectExistsWaiter(client),
		objNotExistsWaiter:  s3.NewObjectNotExistsWaiter(client),
		proxyPresignEnabled: config.options.ProxyPresignEnabled,
		proxyConfig:         config.options.ProxyPresignConfig,
	}, nil
}

// Upload implements storagetypes.Provider
func (p *Provider) Upload(ctx context.Context, reader io.Reader, opts *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
	// Try to infer size from reader if available
	size, sizeKnown := objects.InferReaderSize(reader)

	// Convert reader to ReadSeeker using BufferedReader for small files or temp file for large files
	seeker, err := objects.ReaderToSeeker(reader)
	if err != nil {
		return nil, err
	}

	// If size wasn't known upfront, get it from the seeker
	if !sizeKnown {
		if sized, ok := seeker.(objects.SizedReader); ok {
			size = sized.Size()
		} else {
			// Fall back to seeking to end to get size
			endPos, seekErr := seeker.Seek(0, io.SeekEnd)
			if seekErr != nil {
				return nil, seekErr
			}
			size = endPos
			// Reset to beginning
			_, seekErr = seeker.Seek(0, io.SeekStart)
			if seekErr != nil {
				return nil, seekErr
			}
		}
	}

	objectKey := opts.FileName
	if opts.FolderDestination != "" {
		objectKey = path.Join(opts.FolderDestination, opts.FileName)
	}

	_, err = p.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(p.options.Bucket),
		Key:         aws.String(objectKey),
		Body:        seeker,
		ContentType: aws.String(opts.ContentType),
	})
	if err != nil {
		return nil, err
	}

	metrics.RecordStorageUpload(string(storagetypes.R2Provider), size)

	return &storagetypes.UploadedFileMetadata{
		FileMetadata: storagetypes.FileMetadata{
			Key:          objectKey,
			Size:         size,
			Folder:       opts.FolderDestination,
			Bucket:       p.options.Bucket,
			Region:       p.options.Region,
			ContentType:  opts.ContentType,
			ProviderType: storagetypes.R2Provider,
			FullURI:      fmt.Sprintf("r2://%s/%s", p.options.Bucket, objectKey),
		},
	}, nil
}

// Download implements storagetypes.Provider
func (p *Provider) Download(ctx context.Context, file *storagetypes.File, _ *storagetypes.DownloadFileOptions) (*storagetypes.DownloadedFileMetadata, error) {
	head, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.options.Bucket),
		Key:    aws.String(file.Key),
	})
	if err != nil {
		return nil, err
	}

	buf := make([]byte, int(*head.ContentLength))
	w := manager.NewWriteAtBuffer(buf)

	_, err = p.downloader.DownloadObject(ctx, &transfermanager.DownloadObjectInput{
		Bucket:   aws.String(p.options.Bucket),
		Key:      aws.String(file.Key),
		WriterAt: w,
	})
	if err != nil {
		return nil, err
	}

	downloadedSize := int64(len(w.Bytes()))
	metrics.RecordStorageDownload("r2", downloadedSize)

	return &storagetypes.DownloadedFileMetadata{
		File: w.Bytes(),
		Size: downloadedSize,
	}, nil
}

// Delete implements storagetypes.Provider
func (p *Provider) Delete(ctx context.Context, file *storagetypes.File, _ *storagetypes.DeleteFileOptions) error {
	bucket := file.Bucket
	if bucket == "" {
		bucket = p.options.Bucket
	}

	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(file.Key),
	})
	if err != nil {
		return err
	}

	metrics.RecordStorageDelete("r2")

	return nil
}

// GetPresignedURL implements storagetypes.Provider
func (p *Provider) GetPresignedURL(ctx context.Context, file *storagetypes.File, opts *storagetypes.PresignedURLOptions) (string, error) {
	if opts == nil {
		opts = &storagetypes.PresignedURLOptions{}
	}

	if p.proxyPresignEnabled && p.proxyConfig != nil && p.proxyConfig.TokenManager != nil {
		url, err := proxy.GenerateDownloadURL(ctx, file, opts.Duration, p.proxyConfig)
		if err == nil {
			return url, nil
		}
		if !errors.Is(err, proxy.ErrTokenManagerRequired) && !errors.Is(err, proxy.ErrEntClientRequired) {
			return "", err
		}
	}

	expires := opts.Duration
	if expires == 0 {
		expires = DefaultPresignedURLExpiry
	}

	presignURL, err := p.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(p.options.Bucket),
		Key:                        aws.String(file.Key),
		ResponseContentType:        aws.String(file.ContentType),
		ResponseContentDisposition: aws.String("attachment"),
	}, func(s3opts *s3.PresignOptions) {
		s3opts.Expires = expires
		s3opts.ClientOptions = []func(*s3.Options){
			func(o *s3.Options) {
				o.Region = "auto"
			},
		}
	})
	if err != nil {
		return "", err
	}

	log.Debug().Str("presigned_url", presignURL.URL).Msg("R2 presigned URL created")

	return presignURL.URL, nil
}

// Exists checks if an object exists in R2
func (p *Provider) Exists(ctx context.Context, file *storagetypes.File) (bool, error) {
	_, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.options.Bucket),
		Key:    aws.String(file.Key),
	})
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetScheme returns the URI scheme for R2
func (p *Provider) GetScheme() *string {
	scheme := "r2://"

	return &scheme
}

func (p *Provider) ProviderType() storagetypes.ProviderType {
	return storagetypes.R2Provider
}

// Close cleans up resources
func (p *Provider) Close() error {
	return nil
}

// ListBuckets lists the buckets in the current account.
func (p *Provider) ListBuckets() ([]string, error) {
	var buckets []string

	result, err := p.client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	for _, bucket := range result.Buckets {
		buckets = append(buckets, *bucket.Name)
	}

	return buckets, err
}
