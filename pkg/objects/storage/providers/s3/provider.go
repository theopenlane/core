package s3

import (
	"bytes"
	"context"
	"errors"
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

	"github.com/theopenlane/core/pkg/objects"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"

	"github.com/samber/lo"
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
	Bucket              string
	Region              string
	AccessKeyID         string
	SecretAccessKey     string
	Endpoint            string
	UsePathStyle        bool
	DebugMode           bool
	ACL                 types.ObjectCannedACL
	UseSSL              bool
	PresignedURLTimeout int
	AWSConfig           aws.Config
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
	// Check credentials similar to original implementation
	if lo.IsEmpty(cfg.AccessKeyID) || lo.IsEmpty(cfg.SecretAccessKey) {
		log.Info().Msg("AWS credentials not provided, attempting to use environment variables")

		awsEnvConfig, err := config.NewEnvConfig()
		if err != nil {
			return mo.Err[*Provider](err)
		}

		if lo.IsEmpty(awsEnvConfig.Credentials.AccessKeyID) || lo.IsEmpty(awsEnvConfig.Credentials.SecretAccessKey) {
			log.Error().Err(err).Msg("AWS credentials not found in environment variables")
			return mo.Err[*Provider](ErrS3LoadCredentials)
		}
	}

	// Create credentials with cache as in original
	creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""))

	// Use provided AWSConfig as base if available, otherwise create new
	awsConfig := cfg.AWSConfig
	if awsConfig.Region == "" {
		awsConfig.Region = cfg.Region
	}

	awsConfig.Credentials = creds

	if cfg.Endpoint != "" {
		awsConfig.BaseEndpoint = aws.String(cfg.Endpoint)
	}

	// Create S3 client with same configuration as original
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

func (p *Provider) ProviderType() storagetypes.ProviderType {
	return storagetypes.S3Provider
}

// Upload implements storagetypes.Provider
func (p *Provider) Upload(ctx context.Context, reader io.Reader, opts *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
	b := new(bytes.Buffer)
	reader = io.TeeReader(reader, b)

	n, err := io.Copy(io.Discard, reader)
	if err != nil {
		return nil, err
	}

	// Use objects.ReaderToSeeker as in original
	seeker, err := objects.ReaderToSeeker(b)
	if err != nil {
		return nil, err
	}

	_, err = p.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(p.config.Bucket),
		Key:         aws.String(opts.FileName),
		Body:        seeker,
		ContentType: aws.String(opts.ContentType),
		ACL:         p.config.ACL,
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
	head, err := p.HeadObj(ctx, file.Key)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, int(*head.ContentLength))
	w := manager.NewWriteAtBuffer(buf)

	// Use bucket from options if provided, otherwise use default
	bucket := p.config.Bucket
	if opts.Bucket != "" {
		bucket = opts.Bucket
	}

	_, err = p.downloader.Download(ctx, w, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
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
	if opts.Duration == 0 {
		opts.Duration = DefaultPresignedURLExpiry
	}

	presignURL, err := p.presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket:                     aws.String(p.config.Bucket),
		Key:                        aws.String(file.Key),
		ResponseContentType:        lo.ToPtr("application/octet-stream"),
		ResponseContentDisposition: lo.ToPtr("attachment"),
	}, func(s3opts *s3.PresignOptions) {
		s3opts.Expires = opts.Duration
		s3opts.ClientOptions = []func(*s3.Options){
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

// GetScheme returns the URI scheme for S3
func (p *Provider) GetScheme() *string {
	scheme := "s3://"

	return &scheme
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

// HeadObj checks if an object exists in S3 and returns its metadata
func (p *Provider) HeadObj(ctx context.Context, key string) (*s3.HeadObjectOutput, error) {
	obj, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// Tag updates an existing object in a bucket with specific tags
func (p *Provider) Tag(ctx context.Context, key string, tags map[string]string) error {
	_, err := p.client.PutObjectTagging(ctx, &s3.PutObjectTaggingInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(key),
		Tagging: &types.Tagging{
			TagSet: makeTagSet(tags),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// GetTags returns the tags for an object in a bucket
func (p *Provider) GetTags(ctx context.Context, key string) (map[string]string, error) {
	output, err := p.client.GetObjectTagging(ctx, &s3.GetObjectTaggingInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return makeTagMap(output.TagSet), nil
}

func makeTagSet(tags map[string]string) []types.Tag {
	var tagSet []types.Tag

	for k, v := range tags {
		tagSet = append(tagSet, types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	return tagSet
}

func makeTagMap(tags []types.Tag) map[string]string {
	tagMap := make(map[string]string, len(tags))

	for _, tag := range tags {
		tagMap[*tag.Key] = *tag.Value
	}

	return tagMap
}
