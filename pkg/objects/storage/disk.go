package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/theopenlane/core/pkg/objects"
)

type Disk struct {
	destinationFolder string
	Scheme            string
}

// ensure Disk satisfies the Storage interface
var _ objects.Storage = &Disk{}

func NewDiskStorage(folder string) (*Disk, error) {
	if len(strings.TrimSpace(folder)) == 0 {
		return nil, fmt.Errorf("%w: please provide a valid folder path", ErrInvalidFolderPath)
	}

	return &Disk{
		destinationFolder: folder,
		Scheme:            "file://",
	}, nil
}

func (d *Disk) Close() error { return nil }

func (d *Disk) Upload(ctx context.Context, r io.Reader, opts *objects.UploadFileOptions) (*objects.UploadedFileMetadata, error) {
	f, err := os.Create(filepath.Join(d.destinationFolder, opts.FileName))
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
func (d *Disk) Download(ctx context.Context, key string, opts *objects.DownloadFileOptions) (*objects.DownloadFileMetadata, io.ReadCloser, error) {
	return nil, nil, nil
}

// GetPresignedURL is used to get a presigned URL for a file in the storage backend
// TODO: Implement this method
func (d *Disk) GetPresignedURL(ctx context.Context, key string) (string, error) {
	return "", nil
}
