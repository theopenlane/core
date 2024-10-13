package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/objects"
)

var (
	presignedURLTimeout = 15 * time.Minute
)

// S3Options is used to configure the S3Store
type S3Options struct {
	// Bucket to store objects in
	Bucket string
	// DebugMode will log all requests and responses
	DebugMode bool
	// UsePathStyle allows you to enable the client to use path-style addressing, i.e., https://s3.amazonaws.com/BUCKET/KEY .
	// by default, the S3 client will use virtual hosted bucket addressing when possible( https://BUCKET.s3.amazonaws.com/KEY ).
	UsePathStyle bool
	// ACL should only be used if the bucket supports ACL
	ACL types.ObjectCannedACL
	// KeyNamespace is used to prefix all keys with a namespace
	KeyNamespace string
}

// S3Store is a store that uses S3 as the backend
type S3Store struct {
	Client             *s3.Client
	Opts               S3Options
	PresignClient      *s3.PresignClient
	Downloader         *manager.Downloader
	Uploader           *manager.Uploader
	ObjExistsWaiter    *s3.ObjectExistsWaiter
	ObjNotExistsWaiter *s3.ObjectNotExistsWaiter
	ACL                types.ObjectCannedACL
	CacheControl       string
	Scheme             string
}

// NewS3FromConfig creates a new S3Store from the provided configuration
func NewS3FromConfig(cfg aws.Config, opts S3Options) (*S3Store, error) {
	if isStringEmpty(opts.Bucket) {
		return nil, ErrInvalidS3Bucket
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = opts.UsePathStyle

		if opts.DebugMode {
			o.ClientLogMode = aws.LogSigning | aws.LogRequest | aws.LogResponseWithBody
		}
	})

	return &S3Store{
		Client:     client,
		Opts:       opts,
		Downloader: manager.NewDownloader(client),
		Uploader:   manager.NewUploader(client),
		Scheme:     "s3://",
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

// Close the S3Store satisfying the Storage interface
func (s *S3Store) Close() error { return nil }

// ManagerUpload uploads multiple files to S3
func (s *S3Store) ManagerUpload(ctx context.Context, files [][]byte) error {
	for i, file := range files {
		_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(s.Opts.Bucket),
			Key:    aws.String(fmt.Sprintf("file%d", i)),
			Body:   bytes.NewReader(file),
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

// Download an object from S3 and return the metadata and a reader
func (s *S3Store) Download(ctx context.Context, key string, opts *objects.DownloadFileOptions) (*objects.DownloadFileMetadata, io.ReadCloser, error) {
	output, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Opts.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, nil, err
	}

	return &objects.DownloadFileMetadata{
		FolderDestination: s.Opts.Bucket,
		Key:               key,
		Size:              *output.ContentLength,
	}, output.Body, nil
}

// PresignedURL returns a URL that provides access to a file for 15 minutes
func (s *S3Store) GetPresignedURL(ctx context.Context, key string) (string, error) {
	client := s3.NewPresignClient(s.Client)

	presignURL, err := client.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket:                     aws.String(s.Opts.Bucket),
		Key:                        aws.String(key),
		ResponseContentDisposition: toPointer("attachment"),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = presignedURLTimeout
	})
	if err != nil {
		return "", err
	}

	log.Debug().Str("presigned_url", presignURL.URL).Msg("presigned URL created")

	return presignURL.URL, nil
}
