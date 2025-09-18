package disk

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

const (
	// DefaultDirPermissions defines the default permissions for created directories
	DefaultDirPermissions = 0755
	// DefaultFilePermissions defines the default permissions for created files
	DefaultFilePermissions = 0644
)

// Provider implements the storagetypes.Provider interface for local filesystem storage
type Provider struct {
	config            *Config
	Scheme            string
	destinationFolder string
}

// Config contains configuration for disk provider
type Config struct {
	BasePath string
	Bucket   string
	Key      string
	// LocalURL is the URL to use for the "presigned" URL for the file
	// e.g for local development, this can be http://localhost:17608/files/
	LocalURL string
}

// NewDiskProvider creates a new disk provider instance
func NewDiskProvider(cfg *Config) (*Provider, error) {
	if lo.IsEmpty(cfg.Bucket) {
		return nil, ErrInvalidFolderPath
	}

	disk := &Provider{
		config: cfg,
		Scheme: "file://",
	}

	// create directory if it does not exist
	if _, err := disk.ListBuckets(); os.IsNotExist(err) {
		log.Info().Str("folder", cfg.Bucket).Msg("directory does not exist, creating directory")

		if err := os.MkdirAll(cfg.Bucket, os.ModePerm); err != nil {
			return nil, fmt.Errorf("%w: failed to create directory", ErrInvalidFolderPath)
		}
	}

	disk.destinationFolder = cfg.Bucket

	return disk, nil
}

// Upload implements storagetypes.Provider
func (p *Provider) Upload(_ context.Context, reader io.Reader, opts *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
	f, err := os.Create(filepath.Join(p.config.Bucket, opts.FileName))
	if err != nil {
		return nil, err
	}

	defer f.Close()

	n, err := io.Copy(f, reader)

	return &storagetypes.UploadedFileMetadata{
		FileMetadata: storagetypes.FileMetadata{
			Key:    opts.FileName,
			Size:   n,
			Folder: opts.FolderDestination,
		},
	}, err
}

// Download implements storagetypes.Provider
func (p *Provider) Download(_ context.Context, file *storagetypes.File, opts *storagetypes.DownloadFileOptions) (*storagetypes.DownloadedFileMetadata, error) {
	filePath := filepath.Join(p.config.Bucket, file.Key)
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return &storagetypes.DownloadedFileMetadata{
		File: fileData,
		Size: int64(len(fileData)),
	}, nil
}

// Delete implements storagetypes.Provider
func (p *Provider) Delete(_ context.Context, file *storagetypes.File, opts *storagetypes.DeleteFileOptions) error {
	return os.Remove(filepath.Join(p.config.Bucket, file.Key))
}

// GetPresignedURL implements storagetypes.Provider
func (p *Provider) GetPresignedURL(_ context.Context, file *storagetypes.File, opts *storagetypes.PresignedURLOptions) (string, error) {
	if p.config.LocalURL == "" {
		return "", ErrMissingLocalURL
	}

	base := strings.TrimRight(p.config.LocalURL, "/")

	return fmt.Sprintf("%s/%s", base, file.Key), nil
}

// Exists checks if a file exists on disk
func (p *Provider) Exists(_ context.Context, file *storagetypes.File) (bool, error) {
	fullPath := filepath.Join(p.config.Bucket, file.Key)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("%w %s: %w", ErrDiskCheckExists, fullPath, err)
	}

	return true, nil
}

// GetScheme returns the URI scheme for disk
func (p *Provider) GetScheme() *string {
	scheme := "file://"
	return &scheme
}

// Close cleans up resources
func (p *Provider) Close() error {
	return nil
}

// ListBuckets lists the local bucket if it exists
func (p *Provider) ListBuckets() ([]string, error) {
	if _, err := os.Stat(p.config.Bucket); err != nil {
		return nil, err
	}

	return []string{p.config.Bucket}, nil
}

func (p *Provider) ProviderType() storagetypes.ProviderType {
	return storagetypes.DiskProvider
}
