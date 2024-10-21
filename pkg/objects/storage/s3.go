package storage

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

	"github.com/theopenlane/core/pkg/objects"
)

// ensure S3Store satisfies the Storage interface
var _ objects.Storage = &S3Store{}

// S3Store is a store that uses S3 as the backend
type S3Store struct {
	Client             *s3.Client
	Opts               S3Options
	PresignClient      *s3.PresignClient
	Downloader         *manager.Downloader
	Uploader           *manager.Uploader
	ObjExistsWaiter    *s3.ObjectExistsWaiter
	ObjNotExistsWaiter *s3.ObjectNotExistsWaiter
	Scheme             string
}

// NewS3FromConfig creates a new S3Store from the provided configuration
func NewS3FromConfig(opts S3Options) (*S3Store, error) {
	if isStringEmpty(opts.AccessKeyID) || isStringEmpty(opts.SecretAccessKey) {
		log.Info().Msg("AWS credentials not provided, attempting to use environment variables")

		awsEnvConfig, err := config.NewEnvConfig()
		if err != nil {
			return nil, err
		}

		if isStringEmpty(awsEnvConfig.Credentials.AccessKeyID) || isStringEmpty(awsEnvConfig.Credentials.SecretAccessKey) {
			log.Error().Err(err).Msg("AWS credentials not found in environment variables")
			return nil, err
		}
	}

	creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(opts.AccessKeyID, opts.SecretAccessKey, ""))

	client := s3.NewFromConfig(opts.AWSConfig, func(o *s3.Options) {
		o.UsePathStyle = opts.UsePathStyle
		o.Region = opts.Region
		o.Credentials = creds

		if opts.DebugMode {
			o.ClientLogMode = aws.LogSigning | aws.LogRequest | aws.LogResponseWithBody
		}
	})

	return &S3Store{
		Client:             client,
		Opts:               opts,
		Downloader:         manager.NewDownloader(client),
		Uploader:           manager.NewUploader(client),
		Scheme:             "s3://",
		PresignClient:      s3.NewPresignClient(client),
		ObjExistsWaiter:    s3.NewObjectExistsWaiter(client),
		ObjNotExistsWaiter: s3.NewObjectNotExistsWaiter(client),
	}, nil
}

// Exists checks if an object exists in S3
func (s *S3Store) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.Opts.Bucket),
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

// Exists checks if an object exists in S3
func (s *S3Store) HeadObj(ctx context.Context, key string) (*s3.HeadObjectOutput, error) {
	obj, err := s.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.Opts.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// Close the S3Store satisfying the Storage interface
func (s *S3Store) Close() error { return nil }

// ManagerUpload uploads multiple files to S3
func (s *S3Store) ManagerUpload(ctx context.Context, files [][]byte) error {
	for i, file := range files {
		_, err := s.Uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket: aws.String(s.Opts.Bucket),
			Key:    aws.String(fmt.Sprintf("file%d", i)),
			Body:   bytes.NewReader(file),
		}, func(o *manager.Uploader) {
			o.PartSize = 64 * 1024 * 1024 // nolint: mnd
			o.Concurrency = 5
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// Upload an object to S3 and return the metadata
func (s *S3Store) Upload(ctx context.Context, r io.Reader, opts *objects.UploadFileOptions) (*objects.UploadedFileMetadata, error) {
	b := new(bytes.Buffer)

	r = io.TeeReader(r, b)

	n, err := io.Copy(io.Discard, r)
	if err != nil {
		return nil, err
	}

	seeker, err := objects.ReaderToSeeker(b)
	if err != nil {
		return nil, err
	}

	if _, err = s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.Opts.Bucket),
		Metadata:    opts.Metadata,
		Key:         aws.String(opts.FileName),
		ACL:         s.Opts.ACL,
		Body:        seeker,
		ContentType: aws.String(opts.ContentType),
	}); err != nil {
		return nil, err
	}

	return &objects.UploadedFileMetadata{
		FolderDestination: s.Opts.Bucket,
		Size:              n,
		Key:               opts.FileName,
	}, nil
}

// Download an object from S3 and return the metadata and a reader - the reader must be closed after use and is the responsibiolity of the caller
func (s *S3Store) Download(ctx context.Context, opts *objects.DownloadFileOptions) (*objects.DownloadFileMetadata, error) {
	head, err := s.HeadObj(ctx, opts.FileName)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, int(*head.ContentLength))
	// wrap with aws.WriteAtBuffer
	w := manager.NewWriteAtBuffer(buf)
	// download file into the memories
	_, err = s.Downloader.Download(ctx, w, &s3.GetObjectInput{
		Bucket: aws.String(opts.Bucket),
		Key:    aws.String(opts.FileName),
	})

	if err != nil {
		return nil, err
	}

	return &objects.DownloadFileMetadata{
		File:   w.Bytes(),
		Writer: w,
		Size:   int64(len(w.Bytes())),
	}, nil
}

// PresignedURL returns a URL that provides access to a file for 15 minutes
func (s *S3Store) GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	presignURL, err := s.PresignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket:                     aws.String(s.Opts.Bucket),
		Key:                        aws.String(key),
		ResponseContentType:        toPointer("application/octet-stream"),
		ResponseContentDisposition: toPointer("attachment"),
	}, func(opts *s3.PresignOptions) {
		if expires == 0 {
			expires = 15 * time.Minute // nolint: mnd
		}

		opts.Expires = expires
		opts.ClientOptions = []func(*s3.Options){
			func(o *s3.Options) {
				o.Region = s.Opts.Region
			},
		}
	})
	if err != nil {
		return "", err
	}

	log.Debug().Str("presigned_url", presignURL.URL).Msg("presigned URL created")

	return presignURL.URL, nil
}

// GetScheme returns the scheme of the storage backend
func (s *S3Store) GetScheme() *string {
	return &s.Scheme
}

// Delete an object from S3
func (s *S3Store) Delete(ctx context.Context, key string) error {
	_, err := s.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.Opts.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}

	return nil
}

// Tag updates an existing object in a bucket with specific tags
func (s *S3Store) Tag(ctx context.Context, key string, tags map[string]string) error {
	_, err := s.Client.PutObjectTagging(ctx, &s3.PutObjectTaggingInput{
		Bucket: aws.String(s.Opts.Bucket),
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
func (s *S3Store) GetTags(ctx context.Context, key string) (map[string]string, error) {
	output, err := s.Client.GetObjectTagging(ctx, &s3.GetObjectTaggingInput{
		Bucket: aws.String(s.Opts.Bucket),
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

// ListBuckets lists the buckets in the current account.
func (s *S3Store) ListBuckets() ([]string, error) {
	var buckets []string

	result, err := s.Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	for _, bucket := range result.Buckets {
		buckets = append(buckets, *bucket.Name)
	}

	return buckets, err
}
