package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/objects"
)

type Disk struct {
	// destinationFolder is the folder where the files will be stored
	destinationFolder string
	// Scheme is the scheme of the storage backend
	Scheme string
	// Opts is the options for the disk storage
	Opts *DiskOptions
}

// ensure Disk satisfies the Storage interface
var _ objects.Storage = &Disk{}

// ProviderDisk is the provider for the disk storage
var ProviderDisk = "disk"

// NewDiskStorage creates a new Disk storage backend
func NewDiskStorage(opts *DiskOptions) (*Disk, error) {
	if isStringEmpty(opts.Bucket) {
		return nil, ErrInvalidFolderPath
	}

	disk := &Disk{
		Opts:   opts,
		Scheme: "file://",
	}

	// create directory if it does not exist
	if _, err := disk.ListBuckets(); os.IsNotExist(err) {
		log.Info().Str("folder", opts.Bucket).Msg("directory does not exist, creating directory")

		if err := os.MkdirAll(opts.Bucket, os.ModePerm); err != nil {
			return nil, fmt.Errorf("%w: failed to create directory", ErrInvalidFolderPath)
		}
	}

	return disk, nil
}

// Close satisfies the Storage interface
func (d *Disk) Close() error { return nil }

// Upload satisfies the Storage interface
func (d *Disk) Upload(_ context.Context, r io.Reader, opts *objects.UploadFileOptions) (*objects.UploadedFileMetadata, error) {
	f, err := os.Create(filepath.Join(d.Opts.Bucket, opts.FileName))
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
func (d *Disk) ManagerUpload(_ context.Context, _ [][]byte) error {
	return nil
}

// Download is used to download a file from the storage backend
func (d *Disk) Download(_ context.Context, opts *objects.DownloadFileOptions) (*objects.DownloadFileMetadata, error) {
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
func (d *Disk) GetPresignedURL(key string, _ time.Duration) (string, error) {
	return d.Opts.LocalURL + key, nil
}

// ListBuckets lists the local bucket if it exists
func (d *Disk) ListBuckets() ([]string, error) {
	if _, err := os.Stat(d.Opts.Bucket); err != nil {
		return nil, err
	}

	return []string{d.Opts.Bucket}, nil
}

// Delete removes a file from disk
func (d *Disk) Delete(_ context.Context, key string) error {
	return os.Remove(filepath.Join(d.Opts.Bucket, key))
}
