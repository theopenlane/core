package storage

import (
	"context"
	"io"

	"github.com/theopenlane/core/common/storagetypes"
)

// Opener is implemented by providers that stream an object body instead of buffering it in memory
type Opener interface {
	// Open returns the body of file, closed by the caller
	Open(ctx context.Context, file *File) (io.ReadCloser, error)
}

// Statter is implemented by providers that report object metadata without reading the object body
type Statter interface {
	// Stat returns the metadata of file, or ErrObjectNotFound when the object does not exist
	Stat(ctx context.Context, file *File) (*storagetypes.ObjectInfo, error)
}

// Toucher is implemented by providers that read the smallest possible part of an object, which is enough to make a lazily migrating backing store such as Sippy copy the whole object
type Toucher interface {
	// Touch reads a single byte of file and discards it
	Touch(ctx context.Context, file *File) error
}

// Copier is implemented by providers that copy an object inside the backing store without moving its bytes through the caller
type Copier interface {
	// Copy copies file to the destination named by opts
	Copy(ctx context.Context, file *File, opts *CopyOptions) error
}

// CopyOptions names the destination of a server-side object copy
type CopyOptions struct {
	// Bucket is the destination bucket, the provider's own bucket when empty
	Bucket string
	// Key is the destination object key
	Key string
}
