package storage

type DiskOption func(*DiskOptions)

type DiskOptions struct {
	Bucket string
	Key    string
}

// WithLocalBucket is a DiskOption that sets the bucket for the disk storage which equates to a folder on the file system
func WithLocalBucket(bucket string) DiskOption {
	return func(d *DiskOptions) {
		d.Bucket = bucket
	}
}

// WithLocalKey specifies the name of the file in the local folder
func WithLocalKey(key string) DiskOption {
	return func(d *DiskOptions) {
		d.Key = key
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
