package r2

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/metrics"
	"github.com/theopenlane/core/pkg/objects"
	storage "github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
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
	client             *s3.Client
	options            *storage.ProviderOptions
	presignClient      *s3.PresignClient
	downloader         *manager.Downloader
	uploader           *manager.Uploader
	objExistsWaiter    *s3.ObjectExistsWaiter
	objNotExistsWaiter *s3.ObjectNotExistsWaiter
}

// NewR2Provider creates a new R2 provider instance
func NewR2Provider(options *storage.ProviderOptions) (*Provider, error) {
	if options == nil {
		return nil, ErrR2BucketRequired
	}
	if options.Bucket == "" {
		return nil, ErrR2BucketRequired
	}
	if options.Credentials.AccountID == "" {
		return nil, ErrR2AccountIDRequired
	}
	if options.Credentials.AccessKeyID == "" || options.Credentials.SecretAccessKey == "" {
		return nil, ErrR2CredentialsMissing
	}

	endpoint := options.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", options.Credentials.AccountID)
	}

	creds := credentials.NewStaticCredentialsProvider(options.Credentials.AccessKeyID, options.Credentials.SecretAccessKey, "")

	awsConfig := aws.Config{
		Region:       "auto",
		Credentials:  creds,
		BaseEndpoint: aws.String(endpoint),
	}

	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.Region = "auto"
		o.Credentials = creds
	})

	return &Provider{
		client:             client,
		options:            options.Clone(),
		downloader:         manager.NewDownloader(client),
		uploader:           manager.NewUploader(client),
		presignClient:      s3.NewPresignClient(client),
		objExistsWaiter:    s3.NewObjectExistsWaiter(client),
		objNotExistsWaiter: s3.NewObjectNotExistsWaiter(client),
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

	_, err = p.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(p.options.Bucket),
		Key:         aws.String(opts.FileName),
		Body:        seeker,
		ContentType: aws.String(opts.ContentType),
	})
	if err != nil {
		return nil, err
	}

	metrics.RecordStorageUpload("r2", size)

	return &storagetypes.UploadedFileMetadata{
		FileMetadata: storagetypes.FileMetadata{
			Key:    opts.FileName,
			Size:   size,
			Folder: opts.FolderDestination,
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

	_, err = p.downloader.Download(ctx, w, &s3.GetObjectInput{
		Bucket: aws.String(p.options.Bucket),
		Key:    aws.String(file.Key),
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
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.options.Bucket),
		Key:    aws.String(file.Key),
	})
	if err != nil {
		return err
	}

	metrics.RecordStorageDelete("r2")

	return nil
}

// GetPresignedURL implements storagetypes.Provider
func (p *Provider) GetPresignedURL(_ context.Context, file *storagetypes.File, opts *storagetypes.PresignedURLOptions) (string, error) {
	expires := opts.Duration
	if expires == 0 {
		expires = DefaultPresignedURLExpiry
	}

	presignURL, err := p.presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket:                     aws.String(p.options.Bucket),
		Key:                        aws.String(file.Key),
		ResponseContentType:        aws.String("application/octet-stream"),
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
