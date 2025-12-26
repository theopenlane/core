package disk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/pkg/metrics"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/objects/storage/proxy"
)

const (
	// DefaultDirPermissions defines the default permissions for created directories
	DefaultDirPermissions = 0755
	// DefaultFilePermissions defines the default permissions for created files
	DefaultFilePermissions = 0644
)

// Provider implements the storagetypes.Provider interface for local filesystem storage
type Provider struct {
	options             *storage.ProviderOptions
	Scheme              string
	destinationFolder   string
	proxyPresignEnabled bool
	proxyConfig         *storage.ProxyPresignConfig
}

// NewDiskProvider creates a new disk provider instance
func NewDiskProvider(options *storage.ProviderOptions) (*Provider, error) {
	if options == nil || lo.IsEmpty(options.Bucket) {
		return nil, ErrInvalidFolderPath
	}

	disk := &Provider{
		options: options.Clone(),
		Scheme:  "file://",
	}

	disk.proxyPresignEnabled = options.ProxyPresignEnabled
	disk.proxyConfig = options.ProxyPresignConfig

	if _, err := disk.ListBuckets(); os.IsNotExist(err) {
		log.Info().Str("folder", options.Bucket).Msg("directory does not exist, creating directory")

		if err := os.MkdirAll(options.Bucket, os.ModePerm); err != nil {
			return nil, fmt.Errorf("%w: failed to create directory", ErrInvalidFolderPath)
		}
	}

	disk.destinationFolder = options.Bucket

	return disk, nil
}

// Upload implements storagetypes.Provider
func (p *Provider) Upload(_ context.Context, reader io.Reader, opts *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
	// Try to infer size from reader if available for metadata purposes
	size, sizeKnown := objects.InferReaderSize(reader)

	objectKey := opts.FileName
	if opts.FolderDestination != "" {
		objectKey = filepath.Join(opts.FolderDestination, opts.FileName)
	}

	targetPath := filepath.Join(p.options.Bucket, objectKey)

	if err := os.MkdirAll(filepath.Dir(targetPath), DefaultDirPermissions); err != nil {
		return nil, err
	}

	f, err := os.Create(targetPath)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	n, err := io.Copy(f, reader)
	if err != nil {
		return nil, err
	}

	// Use actual bytes written if size wasn't known upfront
	if !sizeKnown {
		size = n
	}

	metrics.RecordStorageUpload("disk", size)

	return &storagetypes.UploadedFileMetadata{
		FileMetadata: storagetypes.FileMetadata{
			Key:          filepath.ToSlash(objectKey),
			Size:         size,
			Folder:       filepath.ToSlash(opts.FolderDestination),
			Bucket:       p.options.Bucket,
			Region:       p.options.Region,
			ContentType:  opts.ContentType,
			ProviderType: storagetypes.DiskProvider,
			FullURI:      fmt.Sprintf("%s%s", p.Scheme, filepath.ToSlash(targetPath)),
		},
	}, nil
}

// Download implements storagetypes.Provider
func (p *Provider) Download(_ context.Context, file *storagetypes.File, _ *storagetypes.DownloadFileOptions) (*storagetypes.DownloadedFileMetadata, error) {
	filePath := filepath.Join(p.options.Bucket, file.Key)
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	downloadedSize := int64(len(fileData))
	metrics.RecordStorageDownload(string(storagetypes.DiskProvider), downloadedSize)

	return &storagetypes.DownloadedFileMetadata{
		File: fileData,
		Size: downloadedSize,
	}, nil
}

// Delete implements storagetypes.Provider
func (p *Provider) Delete(_ context.Context, file *storagetypes.File, _ *storagetypes.DeleteFileOptions) error {
	bucket := file.Bucket
	if bucket == "" {
		bucket = p.options.Bucket
	}

	err := os.Remove(filepath.Join(bucket, file.Key))
	if os.IsNotExist(err) {
		metrics.RecordStorageDelete(string(storagetypes.DiskProvider))
		return nil
	}
	if err != nil {
		return err
	}

	metrics.RecordStorageDelete(string(storagetypes.DiskProvider))

	return nil
}

// GetPresignedURL implements storagetypes.Provider
func (p *Provider) GetPresignedURL(ctx context.Context, file *storagetypes.File, opts *storagetypes.PresignedURLOptions) (string, error) {
	if opts == nil {
		opts = &storagetypes.PresignedURLOptions{}
	}

	if p.proxyPresignEnabled && p.proxyConfig != nil && p.proxyConfig.TokenManager != nil {
		url, err := proxy.GenerateDownloadURL(ctx, file, opts.Duration, p.proxyConfig)
		if err == nil {
			return url, nil
		}
		if !errors.Is(err, proxy.ErrTokenManagerRequired) && !errors.Is(err, proxy.ErrEntClientRequired) {
			return "", err
		}
	}

	if p.options.LocalURL == "" {
		return "", ErrMissingLocalURL
	}

	base := strings.TrimRight(p.options.LocalURL, "/")

	return fmt.Sprintf("%s/%s", base, file.Key), nil
}

// Exists checks if a file exists on disk
func (p *Provider) Exists(_ context.Context, file *storagetypes.File) (bool, error) {
	fullPath := filepath.Join(p.options.Bucket, file.Key)

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
	if _, err := os.Stat(p.options.Bucket); err != nil {
		return nil, err
	}

	return []string{p.options.Bucket}, nil
}

func (p *Provider) ProviderType() storagetypes.ProviderType {
	return storagetypes.DiskProvider
}
