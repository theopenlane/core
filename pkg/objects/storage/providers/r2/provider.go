package r2

import (
	"bytes"
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
	config             *Config
	presignClient      *s3.PresignClient
	downloader         *manager.Downloader
	uploader           *manager.Uploader
	objExistsWaiter    *s3.ObjectExistsWaiter
	objNotExistsWaiter *s3.ObjectNotExistsWaiter
}

// Config contains configuration for R2 provider
type Config struct {
	Bucket          string
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	APIToken        string
	Endpoint        string
	Region          string
}

// NewR2Provider creates a new R2 provider instance
func NewR2Provider(cfg *Config) (*Provider, error) {
	if cfg == nil {
		return nil, ErrR2BucketRequired
	}
	if cfg.Bucket == "" {
		return nil, ErrR2BucketRequired
	}
	if cfg.AccountID == "" {
		return nil, ErrR2AccountIDRequired
	}
	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, ErrR2CredentialsMissing
	}

	// Set up R2-specific endpoint
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)
	if cfg.Endpoint != "" {
		endpoint = cfg.Endpoint
	}

	creds := credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")

	// Create AWS config with R2 endpoint
	awsConfig := aws.Config{
		Region:       "auto", // R2 uses "auto" as the region
		Credentials:  creds,
		BaseEndpoint: aws.String(endpoint),
	}

	// Create S3 client configured for R2
	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.Region = "auto"
		o.Credentials = creds
	})

	return &Provider{
		client:             client,
		config:             cfg,
		downloader:         manager.NewDownloader(client),
		uploader:           manager.NewUploader(client),
		presignClient:      s3.NewPresignClient(client),
		objExistsWaiter:    s3.NewObjectExistsWaiter(client),
		objNotExistsWaiter: s3.NewObjectNotExistsWaiter(client),
	}, nil
}

// Upload implements storagetypes.Provider
func (p *Provider) Upload(ctx context.Context, reader io.Reader, opts *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
	b := new(bytes.Buffer)
	reader = io.TeeReader(reader, b)

	n, err := io.Copy(io.Discard, reader)
	if err != nil {
		return nil, err
	}

	seeker := bytes.NewReader(b.Bytes())

	_, err = p.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(p.config.Bucket),
		Key:         aws.String(opts.FileName),
		Body:        seeker,
		ContentType: aws.String(opts.ContentType),
	})
	if err != nil {
		return nil, err
	}

	return &storagetypes.UploadedFileMetadata{
		FileMetadata: storagetypes.FileMetadata{
			Key:    opts.FileName,
			Size:   n,
			Folder: opts.FolderDestination,
		},
	}, nil
}

// Download implements storagetypes.Provider
func (p *Provider) Download(ctx context.Context, file *storagetypes.File, opts *storagetypes.DownloadFileOptions) (*storagetypes.DownloadedFileMetadata, error) {
	head, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(file.Key),
	})
	if err != nil {
		return nil, err
	}

	buf := make([]byte, int(*head.ContentLength))
	w := manager.NewWriteAtBuffer(buf)

	_, err = p.downloader.Download(ctx, w, &s3.GetObjectInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(file.Key),
	})
	if err != nil {
		return nil, err
	}

	return &storagetypes.DownloadedFileMetadata{
		File: w.Bytes(),
		Size: int64(len(w.Bytes())),
	}, nil
}

// Delete implements storagetypes.Provider
func (p *Provider) Delete(ctx context.Context, file *storagetypes.File, opts *storagetypes.DeleteFileOptions) error {
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(file.Key),
	})

	return err
}

// GetPresignedURL implements storagetypes.Provider
func (p *Provider) GetPresignedURL(ctx context.Context, file *storagetypes.File, opts *storagetypes.PresignedURLOptions) (string, error) {
	expires := opts.Duration
	if expires == 0 {
		expires = DefaultPresignedURLExpiry
	}

	presignURL, err := p.presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket:                     aws.String(p.config.Bucket),
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
		Bucket: aws.String(p.config.Bucket),
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
