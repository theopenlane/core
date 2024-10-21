package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/theopenlane/core/pkg/objects"
)

type Disk struct {
	destinationFolder string
	// Scheme is the scheme of the storage backend
	Scheme string
	// Bucket is the local folder to store files in
	Bucket string
	// Key is the name of the file in the local folder
	Key string
	// Opts is the options for the disk storage
	Opts DiskOption
}

// ensure Disk satisfies the Storage interface
var _ objects.Storage = &Disk{}

func NewDiskStorage(opts DiskOptions) (*Disk, error) {
	if isStringEmpty(opts.Bucket) {
		return nil, ErrInvalidFolderPath
	}

	return &Disk{
		Bucket: opts.Bucket,
		Scheme: "file://",
	}, nil
}

func (d *Disk) Close() error { return nil }

func (d *Disk) Upload(ctx context.Context, r io.Reader, opts *objects.UploadFileOptions) (*objects.UploadedFileMetadata, error) {
	f, err := os.Create(filepath.Join(opts.Bucket, opts.FileName))
	if err != nil {
		return nil, err
	}

	defer f.Close()

	n, err := io.Copy(f, r)

	return &objects.UploadedFileMetadata{
		FolderDestination: d.destinationFolder,
		Size:              n,
		Key:               opts.FileName,
	}, err
}

// GetScheme returns the scheme of the storage backend
func (d *Disk) GetScheme() *string {
	return &d.Scheme
}

// ManagerUpload uploads multiple files to disk
// TODO: Implement this method
func (d *Disk) ManagerUpload(ctx context.Context, files [][]byte) error {
	return nil
}

// Download is used to download a file from the storage backend
// TODO: Implement this method
func (d *Disk) Download(ctx context.Context, opts *objects.DownloadFileOptions) (*objects.DownloadFileMetadata, error) {
	file, err := os.ReadFile(filepath.Join(opts.Bucket, opts.FileName))
	if err != nil {
		return nil, err
	}

	return &objects.DownloadFileMetadata{
		File: file,
		Size: int64(len(file)),
	}, nil
}

// GetPresignedURL is used to get a presigned URL for a file in the storage backend
// TODO: Implement this method
func (d *Disk) GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	return "", nil
}
