package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/samber/mo"

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

// Provider implements the storagetypes.Provider interface for Amazon S3
type Provider struct {
	client              *s3.Client
	options             *storage.ProviderOptions
	presignClient       *s3.PresignClient
	downloader          *manager.Downloader
	uploader            *manager.Uploader
	objExistsWaiter     *s3.ObjectExistsWaiter
	objNotExistsWaiter  *s3.ObjectNotExistsWaiter
	acl                 types.ObjectCannedACL
	region              string
	proxyPresignEnabled bool
	proxyConfig         *storage.ProxyPresignConfig
}

// providerConfig holds configuration for the S3 provider
type providerConfig struct {
	options      *storage.ProviderOptions
	usePathStyle bool
	debugMode    bool
	awsConfig    *aws.Config
	acl          types.ObjectCannedACL
}

// Option configures the S3 provider during construction
type Option func(*providerConfig)

// WithUsePathStyle configures the S3 client to use path-style addressing
func WithUsePathStyle(use bool) Option {
	return func(cfg *providerConfig) {
		cfg.usePathStyle = use
	}
}

// WithDebugMode enables AWS client debug logging
func WithDebugMode(enabled bool) Option {
	return func(cfg *providerConfig) {
		cfg.debugMode = enabled
	}
}

// WithAWSConfig provides a pre-configured AWS config
func WithAWSConfig(c aws.Config) Option {
	return func(cfg *providerConfig) {
		cfg.awsConfig = &c
	}
}

// WithACL sets the canned ACL applied during uploads
func WithACL(acl types.ObjectCannedACL) Option {
	return func(cfg *providerConfig) {
		cfg.acl = acl
	}
}

// NewS3Provider creates a new S3 provider instance
func NewS3Provider(options *storage.ProviderOptions, opts ...Option) (*Provider, error) {
	return NewS3ProviderResult(options, opts...).Get()
}

// NewS3ProviderResult creates a new S3 provider instance with mo.Result error handling
func NewS3ProviderResult(options *storage.ProviderOptions, opts ...Option) mo.Result[*Provider] {
	config := defaultProviderConfig(options)
	for _, opt := range opts {
		if opt != nil {
			opt(&config)
		}
	}

	validated := validateS3Config(config)
	if validated.IsError() {
		return mo.Err[*Provider](validated.Error())
	}

	return createS3Provider(validated.MustGet())
}

func defaultProviderConfig(options *storage.ProviderOptions) providerConfig {
	return providerConfig{
		options: options.Clone(),
	}
}

func validateS3Config(cfg providerConfig) mo.Result[providerConfig] {
	if cfg.options == nil || cfg.options.Bucket == "" {
		return mo.Err[providerConfig](ErrS3BucketRequired)
	}
	if cfg.options.Region == "" {
		return mo.Err[providerConfig](ErrS3CredentialsRequired)
	}

	return mo.Ok(cfg)
}

// createS3Provider creates the S3 provider after configuration is validated
func createS3Provider(cfg providerConfig) mo.Result[*Provider] {
	// Check credentials similar to original implementation
	if lo.IsEmpty(cfg.options.Credentials.AccessKeyID) || lo.IsEmpty(cfg.options.Credentials.SecretAccessKey) {
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
	creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(cfg.options.Credentials.AccessKeyID, cfg.options.Credentials.SecretAccessKey, ""))

	awsConfig := aws.Config{}
	if cfg.awsConfig != nil {
		awsConfig = *cfg.awsConfig
	}

	if awsConfig.Region == "" {
		awsConfig.Region = cfg.options.Region
	}

	awsConfig.Credentials = creds

	if cfg.options.Endpoint != "" {
		awsConfig.BaseEndpoint = aws.String(cfg.options.Endpoint)
		if !cfg.usePathStyle {
			cfg.usePathStyle = true
		}
	}

	// Create S3 client with same configuration as original
	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = cfg.usePathStyle
		o.Region = cfg.options.Region
		o.Credentials = creds

		if cfg.debugMode {
			o.ClientLogMode = aws.LogSigning | aws.LogRequest | aws.LogResponseWithBody
		}
	})

	provider := &Provider{
		client:             client,
		options:            cfg.options.Clone(),
		downloader:         manager.NewDownloader(client),
		uploader:           manager.NewUploader(client),
		presignClient:      s3.NewPresignClient(client),
		objExistsWaiter:    s3.NewObjectExistsWaiter(client),
		objNotExistsWaiter: s3.NewObjectNotExistsWaiter(client),
		acl:                cfg.acl,
		region:             awsConfig.Region,
	}

	provider.proxyPresignEnabled = cfg.options.ProxyPresignEnabled
	provider.proxyConfig = cfg.options.ProxyPresignConfig

	return mo.Ok(provider)
}

func (p *Provider) ProviderType() storagetypes.ProviderType {
	return storagetypes.S3Provider
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
		ACL:         p.acl,
	})
	if err != nil {
		return nil, err
	}

	metrics.RecordStorageUpload(string(storagetypes.S3Provider), size)

	return &storagetypes.UploadedFileMetadata{
		FileMetadata: storagetypes.FileMetadata{
			Key:          objectKey,
			Size:         size,
			Folder:       opts.FolderDestination,
			Bucket:       p.options.Bucket,
			Region:       p.options.Region,
			ContentType:  opts.ContentType,
			ProviderType: storagetypes.S3Provider,
			FullURI:      fmt.Sprintf("s3://%s/%s", p.options.Bucket, objectKey),
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
	bucket := p.options.Bucket
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

	downloadedSize := int64(len(w.Bytes()))
	metrics.RecordStorageDownload(string(storagetypes.S3Provider), downloadedSize)

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

	metrics.RecordStorageDelete("s3")

	return nil
}

// GetPresignedURL implements storagetypes.Provider
func (p *Provider) GetPresignedURL(ctx context.Context, file *storagetypes.File, opts *storagetypes.PresignedURLOptions) (string, error) {
	if opts == nil {
		opts = &storagetypes.PresignedURLOptions{}
	}

	if p.proxyPresignEnabled && p.proxyConfig != nil && p.proxyConfig.TokenManager != nil {
		dur := opts.Duration

		url, err := proxy.GenerateDownloadURL(ctx, file, dur, p.proxyConfig)
		if err == nil {
			return url, nil
		}

		if !errors.Is(err, proxy.ErrTokenManagerRequired) && !errors.Is(err, proxy.ErrEntClientRequired) {
			return "", err
		}
	}

	if opts.Duration == 0 {
		opts.Duration = DefaultPresignedURLExpiry
	}

	region := file.Region
	if region == "" {
		region = p.region
	}

	presignURL, err := p.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(file.Bucket),
		Key:                        aws.String(file.Key),
		ResponseContentType:        aws.String(file.ContentType),
		ResponseContentDisposition: aws.String("attachment"),
	}, func(s3opts *s3.PresignOptions) {
		s3opts.Expires = opts.Duration
		s3opts.ClientOptions = []func(*s3.Options){
			func(o *s3.Options) {
				o.Region = region
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
		Bucket: aws.String(p.options.Bucket),
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
		Bucket: aws.String(p.options.Bucket),
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
		Bucket: aws.String(p.options.Bucket),
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
