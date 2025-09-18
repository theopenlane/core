package disk

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

const (
	// DefaultDirPermissions defines the default permissions for created directories
	DefaultDirPermissions = 0755
	// DefaultFilePermissions defines the default permissions for created files
	DefaultFilePermissions = 0644
)

// Provider implements the storagetypes.Provider interface for local filesystem storage
type Provider struct {
	config *Config
}

// Config contains configuration for disk provider
type Config struct {
	BasePath string
	LocalURL string
}

// NewDiskProvider creates a new disk provider instance
func NewDiskProvider(cfg *Config) (*Provider, error) {
	if cfg == nil {
		return nil, ErrDiskBasePathRequired
	}
	if cfg.BasePath == "" {
		return nil, ErrDiskBasePathRequired
	}

	// Ensure base path exists
	if err := os.MkdirAll(cfg.BasePath, DefaultDirPermissions); err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrDiskCreateBasePath, cfg.BasePath, err)
	}

	return &Provider{
		config: cfg,
	}, nil
}

// Upload implements storagetypes.Provider
func (p *Provider) Upload(_ context.Context, reader io.Reader, opts *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
	fullPath := filepath.Join(p.config.BasePath, opts.FileName)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, DefaultDirPermissions); err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrDiskCreateDirectory, dir, err)
	}

	// Create/truncate file
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrDiskCreateFile, fullPath, err)
	}
	defer file.Close()

	// Copy data
	n, err := io.Copy(file, reader)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrDiskWriteFile, fullPath, err)
	}

	return &storagetypes.UploadedFileMetadata{
		FileStorageMetadata: storagetypes.FileStorageMetadata{
			Key:  opts.FileName,
			Size: n,
		},
		FolderDestination: p.config.BasePath,
	}, nil
}

// Download implements storagetypes.Provider
func (p *Provider) Download(_ context.Context, opts *storagetypes.DownloadFileOptions) (*storagetypes.DownloadFileMetadata, error) {
	fullPath := filepath.Join(p.config.BasePath, opts.FileName)

	// Check if file exists
	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrDiskFileNotFound, opts.FileName)
		}
		return nil, fmt.Errorf("%w %s: %w", ErrDiskStatFile, fullPath, err)
	}

	// Read file
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrDiskReadFile, fullPath, err)
	}

	return &storagetypes.DownloadFileMetadata{
		File: data,
		Size: stat.Size(),
	}, nil
}

// Delete implements storagetypes.Provider
func (p *Provider) Delete(_ context.Context, key string) error {
	fullPath := filepath.Join(p.config.BasePath, key)

	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("%w %s: %w", ErrDiskDeleteFile, fullPath, err)
	}

	return nil
}

// GetPresignedURL implements storagetypes.Provider
func (p *Provider) GetPresignedURL(key string, _ time.Duration) (string, error) {
	// For disk storage, we can return a local URL if configured
	if p.config.LocalURL != "" {
		return fmt.Sprintf("%s/%s", p.config.LocalURL, key), nil
	}

	// Otherwise, return a file:// URL
	fullPath := filepath.Join(p.config.BasePath, key)

	return fmt.Sprintf("file://%s", fullPath), nil
}

// Exists checks if a file exists on disk
func (p *Provider) Exists(_ context.Context, key string) (bool, error) {
	fullPath := filepath.Join(p.config.BasePath, key)

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
