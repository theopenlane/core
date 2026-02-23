//go:build examples

package app

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// lifecycleConfig configures the standard upload/download lifecycle.
type lifecycleConfig struct {
	FileName        string
	ContentType     string
	Bucket          string
	Reader          io.Reader
	ProviderLabel   string
	PresignDuration time.Duration
	AfterDownload   func(*storage.DownloadedMetadata) error
}

// runLifecycle performs the upload → download → presign → delete workflow.
func runLifecycle(ctx context.Context, out io.Writer, svc *storage.ObjectService, provider storage.Provider, cfg lifecycleConfig) (*storage.File, error) {
	if cfg.Reader == nil {
		return nil, fmt.Errorf("lifecycle reader is required")
	}

	if cfg.ProviderLabel == "" {
		cfg.ProviderLabel = string(provider.ProviderType())
	}

	if cfg.PresignDuration <= 0 {
		cfg.PresignDuration = 15 * time.Minute
	}

	if closer, ok := cfg.Reader.(io.Closer); ok {
		defer closer.Close()
	}

	fmt.Fprintln(out, "1. Uploading file...")
	uploadOpts := &storage.UploadOptions{
		FileName:    cfg.FileName,
		ContentType: cfg.ContentType,
		Bucket:      cfg.Bucket,
	}

	uploadedFile, err := svc.Upload(ctx, provider, cfg.Reader, uploadOpts)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	fmt.Fprintf(out, "Uploaded: %s (size: %d bytes)\n", uploadedFile.Key, uploadedFile.Size)

	fmt.Fprintln(out, "\n2. Downloading file...")
	storageFile := &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			Key:         uploadedFile.Key,
			Bucket:      uploadedFile.Bucket,
			Size:        uploadedFile.Size,
			ContentType: uploadedFile.ContentType,
		},
	}

	downloaded, err := svc.Download(ctx, provider, storageFile, &storage.DownloadOptions{})
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}

	fmt.Fprintf(out, "Downloaded: %s (%d bytes)\n", uploadedFile.Key, len(downloaded.File))
	if cfg.AfterDownload != nil {
		if err := cfg.AfterDownload(downloaded); err != nil {
			return nil, fmt.Errorf("post-download handler failed: %w", err)
		}
	}

	fmt.Fprintln(out, "\n3. Getting presigned URL...")
	presignedURL, err := svc.GetPresignedURL(ctx, provider, storageFile, &storagetypes.PresignedURLOptions{Duration: cfg.PresignDuration})
	if err != nil {
		return nil, fmt.Errorf("failed to get presigned URL: %w", err)
	}
	fmt.Fprintf(out, "URL: %s\n", presignedURL)

	fmt.Fprintf(out, "\n4. Checking %s file existence...\n", cfg.ProviderLabel)
	exists, err := provider.Exists(ctx, storageFile)
	if err != nil {
		return nil, fmt.Errorf("exists check failed: %w", err)
	}
	fmt.Fprintf(out, "File exists: %v\n", exists)

	fmt.Fprintln(out, "\n5. Deleting file...")
	if err := svc.Delete(ctx, provider, storageFile, &storagetypes.DeleteFileOptions{}); err != nil {
		return nil, fmt.Errorf("delete failed: %w", err)
	}
	fmt.Fprintf(out, "File deleted successfully from %s\n", cfg.ProviderLabel)

	exists, err = provider.Exists(ctx, storageFile)
	if err != nil {
		return nil, fmt.Errorf("post-delete exists check failed: %w", err)
	}
	fmt.Fprintf(out, "File exists after deletion: %v\n", exists)

	return uploadedFile, nil
}
