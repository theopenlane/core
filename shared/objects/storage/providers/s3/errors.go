package s3

import "errors"

var (
	// ErrS3BucketRequired is returned when S3 bucket is not specified
	ErrS3BucketRequired = errors.New("S3 bucket is required")
	// ErrS3CredentialsRequired is returned when required S3 credentials are missing
	ErrS3CredentialsRequired = errors.New("missing required S3 credentials: bucket, region")
	// ErrS3SecretCredentialRequired is returned when S3 secret access key is missing
	ErrS3SecretCredentialRequired = errors.New("missing required S3 secret access key, id")
	// ErrS3LoadCredentials is returned when AWS credentials fail to load
	ErrS3LoadCredentials = errors.New("failed to load AWS credentials")
)
