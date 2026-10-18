package storage

import "context"

// CopyOptions names the destination of a server-side object copy
type CopyOptions struct {
	// Bucket is the destination bucket, the provider's own bucket when empty
	Bucket string
	// Key is the destination object key
	Key string
}

// Copier is implemented by providers that copy an object inside the backing store without moving its bytes through the caller
type Copier interface {
	// Copy copies file to the destination named by opts
	Copy(ctx context.Context, file *File, opts *CopyOptions) error
}
