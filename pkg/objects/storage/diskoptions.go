package storage

// DiskOption is an options function for the DiskOptions
type DiskOption func(*DiskOptions)

// DiskOptions are options for the disk storage
type DiskOptions struct {
	Bucket string
	Key    string
	// LocalURL is the URL to use for the "presigned" URL for the file
	// e.g for local development, this can be http://localhost:17608/files/
	LocalURL string
}

// WithLocalBucket is a DiskOption that sets the bucket for the disk storage which equates to a folder on the file system
func WithDiskLocalBucket(bucket string) DiskOption {
	return func(d *DiskOptions) {
		d.Bucket = bucket
	}
}

// WithLocalKey specifies the name of the file in the local folder
func WithDiskLocalKey(key string) DiskOption {
	return func(d *DiskOptions) {
		d.Key = key
	}
}

// WithLocalURL specifies the URL to use for the "presigned" URL for the file
func WithDiskLocalURL(url string) DiskOption {
	return func(d *DiskOptions) {
		d.LocalURL = url
	}
}

// NewDiskOptions returns a new DiskOptions struct
func NewDiskOptions(opts ...DiskOption) *DiskOptions {
	o := &DiskOptions{}
	for _, opt := range opts {
		opt(o)
	}

	return o
}
