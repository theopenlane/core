package storage

import "google.golang.org/api/option"

// GCSOptions is used to configure the GCSStore
// Credentials should be provided via environment variables or client options
// when creating the client
type GCSOptions struct {
	// Bucket is the bucket to store objects in
	Bucket string
	// ProjectID is required when listing buckets
	ProjectID string
	// GoogleAccessID is used for generating signed URLs
	GoogleAccessID string
	// PrivateKey is used for generating signed URLs
	PrivateKey []byte
	// ClientOptions are passed directly to storage.NewClient
	ClientOptions []option.ClientOption
}

// GCSOption is a function that modifies GCSOptions
type GCSOption func(*GCSOptions)

// WithGCSBucket sets the bucket
func WithGCSBucket(b string) GCSOption {
	return func(o *GCSOptions) { o.Bucket = b }
}

// WithGCSProjectID sets the project ID used when listing buckets
func WithGCSProjectID(id string) GCSOption {
	return func(o *GCSOptions) { o.ProjectID = id }
}

// WithGCSGoogleAccessID sets the service account email used when signing URLs
func WithGCSGoogleAccessID(id string) GCSOption {
	return func(o *GCSOptions) { o.GoogleAccessID = id }
}

// WithGCSPrivateKey sets the private key used when signing URLs
func WithGCSPrivateKey(key []byte) GCSOption {
	return func(o *GCSOptions) { o.PrivateKey = key }
}

// WithGCSClientOptions provides additional client options to storage.NewClient
func WithGCSClientOptions(opts ...option.ClientOption) GCSOption {
	return func(o *GCSOptions) { o.ClientOptions = append(o.ClientOptions, opts...) }
}

// NewGCSOptions creates a new GCSOptions instance
func NewGCSOptions(opts ...GCSOption) *GCSOptions {
	o := &GCSOptions{}

	for _, opt := range opts {
		opt(o)
	}

	return o
}
