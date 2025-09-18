package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog/log"
	"github.com/samber/mo"

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

// Provider implements the storagetypes.Provider interface for Amazon S3
type Provider struct {
	client             *s3.Client
	config             *Config
	presignClient      *s3.PresignClient
	downloader         *manager.Downloader
	uploader           *manager.Uploader
	objExistsWaiter    *s3.ObjectExistsWaiter
	objNotExistsWaiter *s3.ObjectNotExistsWaiter
}

// Config contains configuration for S3 provider
type Config struct {
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string
	UsePathStyle    bool
	DebugMode       bool
}

// NewS3Provider creates a new S3 provider instance
func NewS3Provider(cfg *Config) (*Provider, error) {
	return NewS3ProviderResult(cfg).Get()
}

// NewS3ProviderResult creates a new S3 provider instance with mo.Result error handling
func NewS3ProviderResult(cfg *Config) mo.Result[*Provider] {
	configResult := validateS3Config(cfg)
	if configResult.IsError() {
		return mo.Err[*Provider](configResult.Error())
	}
	return createS3Provider(configResult.MustGet())
}

func validateS3Config(cfg *Config) mo.Result[*Config] {
	if cfg.Bucket == "" {
		return mo.Err[*Config](ErrS3BucketRequired)
	}
	return mo.Ok(cfg)
}

func createS3Provider(cfg *Config) mo.Result[*Provider] {
	// Set up AWS credentials
	var creds aws.CredentialsProvider
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		creds = credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")
	} else {
		// Try environment variables
		awsConfig, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return mo.Err[*Provider](fmt.Errorf("%w: %w", ErrS3LoadCredentials, err))
		}
		creds = awsConfig.Credentials
	}

	// Create AWS config
	awsConfig := aws.Config{
		Region:      cfg.Region,
		Credentials: creds,
	}

	if cfg.Endpoint != "" {
		awsConfig.BaseEndpoint = aws.String(cfg.Endpoint)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
		o.Region = cfg.Region
		o.Credentials = creds

		if cfg.DebugMode {
			o.ClientLogMode = aws.LogSigning | aws.LogRequest | aws.LogResponseWithBody
		}
	})

	provider := &Provider{
		client:             client,
		config:             cfg,
		downloader:         manager.NewDownloader(client),
		uploader:           manager.NewUploader(client),
		presignClient:      s3.NewPresignClient(client),
		objExistsWaiter:    s3.NewObjectExistsWaiter(client),
		objNotExistsWaiter: s3.NewObjectNotExistsWaiter(client),
	}

	return mo.Ok(provider)
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
		Metadata:    opts.Metadata,
	})
	if err != nil {
		return nil, err
	}

	return &storagetypes.UploadedFileMetadata{
		FileStorageMetadata: storagetypes.FileStorageMetadata{
			Key:  opts.FileName,
			Size: n,
		},
		FolderDestination: p.config.Bucket,
	}, nil
}

// Download implements storagetypes.Provider
func (p *Provider) Download(ctx context.Context, opts *storagetypes.DownloadFileOptions) (*storagetypes.DownloadFileMetadata, error) {
	head, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(opts.FileName),
	})
	if err != nil {
		return nil, err
	}

	buf := make([]byte, int(*head.ContentLength))
	w := manager.NewWriteAtBuffer(buf)

	_, err = p.downloader.Download(ctx, w, &s3.GetObjectInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(opts.FileName),
	})
	if err != nil {
		return nil, err
	}

	return &storagetypes.DownloadFileMetadata{
		File:   w.Bytes(),
		Size:   int64(len(w.Bytes())),
		Writer: w,
	}, nil
}

// Delete implements storagetypes.Provider
func (p *Provider) Delete(ctx context.Context, key string) error {
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(key),
	})

	return err
}

// GetPresignedURL implements storagetypes.Provider
func (p *Provider) GetPresignedURL(key string, expires time.Duration) (string, error) {
	if expires == 0 {
		expires = DefaultPresignedURLExpiry
	}

	presignURL, err := p.presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket:                     aws.String(p.config.Bucket),
		Key:                        aws.String(key),
		ResponseContentType:        aws.String("application/octet-stream"),
		ResponseContentDisposition: aws.String("attachment"),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expires
		opts.ClientOptions = []func(*s3.Options){
			func(o *s3.Options) {
				o.Region = p.config.Region
			},
		}
	})
	if err != nil {
		return "", err
	}

	log.Debug().Str("presigned_url", presignURL.URL).Msg("S3 presigned URL created")

	return presignURL.URL, nil
}

// Exists checks if an object exists in S3
func (p *Provider) Exists(ctx context.Context, key string) (bool, error) {
	_, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(key),
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

// GetScheme returns the URI scheme for S3
func (p *Provider) GetScheme() *string {
	scheme := "s3://"

	return &scheme
}

// Close cleans up resources
func (p *Provider) Close() error {
	return nil
}
