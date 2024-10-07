package storage

import (
	"bytes"
	"context"
	"crypto/rsa"
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

type S3Options struct {
	Bucket string
	// If true, this will log request and responses
	DebugMode bool

	UsePathStyle bool

	// Only use if the bucket supports ACL
	ACL            types.ObjectCannedACL
	Keynamespace   string
	requestTimeout time.Duration
	Privatekey     *rsa.PrivateKey
}

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
}

func NewS3FromConfig(cfg aws.Config, opts S3Options) (*S3Store, error) {
	if IsStringEmpty(opts.Bucket) {
		return nil, ErrProvideValidS3Bucket
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
	}, nil
}

func (s *S3Store) Close() error { return nil }

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

	_, err = s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.Opts.Bucket),
		Metadata:    opts.Metadata,
		Key:         aws.String(opts.FileName),
		ACL:         s.Opts.ACL,
		Body:        seeker,
		ContentType: aws.String(opts.ContentType),
	})
	if err != nil {
		return nil, err
	}

	return &objects.UploadedFileMetadata{
		FolderDestination: s.Opts.Bucket,
		Size:              n,
		Key:               opts.FileName,
	}, nil
}

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
func (s *S3Store) GetPresignedURL(cntext context.Context, key string) string {
	presignClient := s3.NewPresignClient(s.Client)

	presignurl, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket:                     aws.String(s.Opts.Bucket),
		Key:                        aws.String(key),
		ResponseContentDisposition: StringPointer("attachment"),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute // nolint:mnd
	})
	if err != nil {
		return ""
	}

	log.Info().Str("presigned_url", presignurl.URL).Msg("HAY MATT presigned URL")

	return presignurl.URL
}
