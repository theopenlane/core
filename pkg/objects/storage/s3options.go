package storage

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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
	// Region is the region to use for the S3 client
	Region string
	// AccessKeyID is the access key ID to use for the S3 client
	AccessKeyID string
	// SecretAccessKey is the secret access key to use for the S3 client
	SecretAccessKey string
	// Endpoint is the endpoint to use for the S3 client
	Endpoint string
	// UseSSL is a flag to determine if the S3 client should use SSL
	UseSSL bool
	// PresignURLTimeout is the timeout for presigned URLs
	PresignedURLTimeout int
	// AWSConfig is the AWS configuration to use for the S3 client
	AWSConfig aws.Config
}

// S3Option is a function that modifies S3Options
type S3Option func(*S3Options)

// WithRegion sets the region for S3Options
func WithRegion(region string) S3Option {
	return func(o *S3Options) {
		o.Region = region
	}
}

// WithBucket sets the bucket for S3Options
func WithBucket(bucket string) S3Option {
	return func(o *S3Options) {
		o.Bucket = bucket
	}
}

// WithAccessKeyID sets the access key ID for S3Options
func WithAccessKeyID(accessKeyID string) S3Option {
	return func(o *S3Options) {
		o.AccessKeyID = accessKeyID
	}
}

// WithSecretAccessKey sets the secret access key for S3Options
func WithSecretAccessKey(secretAccessKey string) S3Option {
	return func(o *S3Options) {
		o.SecretAccessKey = secretAccessKey
	}
}

// WithEndpoint sets the endpoint for S3Options
func WithEndpoint(endpoint string) S3Option {
	return func(o *S3Options) {
		o.Endpoint = endpoint
	}
}

// WithUseSSL sets the use SSL flag for S3Options
func WithUseSSL(useSSL bool) S3Option {
	return func(o *S3Options) {
		o.UseSSL = useSSL
	}
}

// WithPresignedURLTimeout sets the presigned URL timeout for S3Options
func WithPresignedURLTimeout(timeout int) S3Option {
	return func(o *S3Options) {
		o.PresignedURLTimeout = timeout
	}
}

// WithAWSConfig sets the AWS configuration for S3Options
func WithAWSConfig(cfg aws.Config) S3Option {
	return func(o *S3Options) {
		o.AWSConfig = cfg
	}
}

// WithPathStyle allows you set the path style. This is useful for
// other compatible s3 storage systems
func WithPathStyle(v bool) S3Option {
	return func(o *S3Options) {
		o.UsePathStyle = v
	}
}

// NewS3Options creates a new S3Options instance with the provided options
func NewS3Options(opts ...S3Option) *S3Options {
	options := &S3Options{}
	for _, opt := range opts {
		opt(options)
	}

	return options
}
