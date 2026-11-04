package r2_test

import (
	"bytes"
	"context"
	"path"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	r2provider "github.com/theopenlane/core/v2/pkg/objects/storage/providers/r2"
	"github.com/theopenlane/core/v2/pkg/pdftext"
)

func TestR2Provider_UploadAndDownload_EmptyFile(t *testing.T) {
	ctx := context.Background()
	orgID := "01HYQZ5YTVJ0P2R2HF7N3W3MQZ"
	fileID := "01J1FILEEMPTY"
	fileName := "empty.txt"
	folder := path.Join(orgID, fileID)

	endpoint, terminate := startMinio(t, ctx)
	t.Cleanup(func() { _ = terminate(context.Background()) })
	createBucket(t, ctx, endpoint, minioBucket)

	builder := r2provider.NewR2Builder().WithOptions(r2provider.WithUsePathStyle(true))
	options := storage.NewProviderOptions(
		storage.WithBucket(minioBucket),
		storage.WithEndpoint(endpoint),
		storage.WithCredentials(storage.ProviderCredentials{
			AccountID:       "test-account",
			AccessKeyID:     minioUser,
			SecretAccessKey: minioSecret,
		}),
	)
	providerInterface, err := builder.Build(ctx, storage.ProviderCredentials{
		AccountID:       "test-account",
		AccessKeyID:     minioUser,
		SecretAccessKey: minioSecret,
	}, options)
	assert.NilError(t, err)
	provider := providerInterface
	t.Cleanup(func() { _ = provider.Close() })

	fileContent := []byte{}
	uploadedFile, err := provider.Upload(ctx, bytes.NewReader(fileContent), &storagetypes.UploadFileOptions{
		FileName:          fileName,
		FolderDestination: folder,
		ContentType:       "text/plain",
	})
	assert.NilError(t, err)
	assert.Assert(t, uploadedFile != nil)

	// Download and verify empty content
	downloadedFile, err := provider.Download(ctx, &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			Key:    path.Join(folder, fileName),
			Bucket: minioBucket,
		},
	}, &storagetypes.DownloadFileOptions{
		FileName: fileName,
	})
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(string(fileContent), string(downloadedFile.File)))
}

func TestR2Provider_Download_NonExistentFile(t *testing.T) {
	ctx := context.Background()
	endpoint, terminate := startMinio(t, ctx)
	t.Cleanup(func() { _ = terminate(context.Background()) })
	createBucket(t, ctx, endpoint, minioBucket)

	builder := r2provider.NewR2Builder().WithOptions(r2provider.WithUsePathStyle(true))
	options := storage.NewProviderOptions(
		storage.WithBucket(minioBucket),
		storage.WithEndpoint(endpoint),
		storage.WithCredentials(storage.ProviderCredentials{
			AccountID:       "test-account",
			AccessKeyID:     minioUser,
			SecretAccessKey: minioSecret,
		}),
	)
	providerInterface, err := builder.Build(ctx, storage.ProviderCredentials{
		AccountID:       "test-account",
		AccessKeyID:     minioUser,
		SecretAccessKey: minioSecret,
	}, options)
	assert.NilError(t, err)
	provider := providerInterface
	t.Cleanup(func() { _ = provider.Close() })

	// Attempt to download a non-existent file
	_, err = provider.Download(ctx, &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			Key:    "does/not/exist.txt",
			Bucket: minioBucket,
		},
	}, &storagetypes.DownloadFileOptions{
		FileName: "does-not-exist.txt",
	})
	assert.Assert(t, err != nil, "expected error when downloading non-existent file")
}

// TestR2Provider_OrganizationFolderStructure verifies that files are stored under organization-specific directories
func TestR2Provider_OrganizationFolderStructure(t *testing.T) {
	ctx := context.Background()
	orgID := "01HYQZ5YTVJ0P2R2HF7N3W3MQZ"
	parentID := "01J1PARENTXYZABCD1234"
	fileID := "01J1FILEXYZABCD5678"
	fileName := "document.pdf"
	folder := path.Join(orgID, parentID, fileID)

	endpoint, terminate := startMinio(t, ctx)
	t.Cleanup(func() { _ = terminate(context.Background()) })

	createBucket(t, ctx, endpoint, minioBucket)

	builder := r2provider.NewR2Builder().WithOptions(r2provider.WithUsePathStyle(true))

	options := storage.NewProviderOptions(
		storage.WithBucket(minioBucket),
		storage.WithEndpoint(endpoint),
		storage.WithCredentials(storage.ProviderCredentials{
			AccountID:       "test-account",
			AccessKeyID:     minioUser,
			SecretAccessKey: minioSecret,
		}),
	)

	providerInterface, err := builder.Build(ctx, storage.ProviderCredentials{
		AccountID:       "test-account",
		AccessKeyID:     minioUser,
		SecretAccessKey: minioSecret,
	}, options)
	assert.NilError(t, err)
	provider := providerInterface
	t.Cleanup(func() { _ = provider.Close() })

	fileContent := []byte("test content for organization folder structure")

	// Upload file with hierarchical folder destination
	uploadedFile, err := provider.Upload(ctx, bytes.NewReader(fileContent), &storagetypes.UploadFileOptions{
		FileName:          fileName,
		FolderDestination: folder,
		ContentType:       pdftext.ContentType,
	})
	assert.NilError(t, err)
	assert.Assert(t, uploadedFile != nil)

	// Verify the key includes the organization/parent/file hierarchy
	expectedKey := path.Join(folder, fileName)
	assert.Assert(t, is.Equal(expectedKey, uploadedFile.Key), "file key should be orgID/parentID/fileID/filename")
	assert.Assert(t, is.Equal(folder, uploadedFile.Folder), "folder should be set to orgID/parentID/fileID")

	// Verify the full URI reflects the organization structure
	expectedURI := "r2://" + path.Join(minioBucket, folder, fileName)
	assert.Assert(t, is.Equal(expectedURI, uploadedFile.FullURI), "full URI should include hierarchical path")

	// Verify file can be downloaded using the organization-prefixed key
	downloadedFile, err := provider.Download(ctx, &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			Key:    expectedKey,
			Bucket: minioBucket,
		},
	}, &storagetypes.DownloadFileOptions{})
	assert.NilError(t, err)
	assert.Assert(t, downloadedFile != nil)
	assert.Assert(t, is.DeepEqual(fileContent, downloadedFile.File))
}

// TestR2Provider_WithoutOrganizationFolder verifies behavior when no organization folder is specified
func TestR2Provider_WithoutOrganizationFolder(t *testing.T) {
	ctx := context.Background()
	fileName := "standalone-document.pdf"

	endpoint, terminate := startMinio(t, ctx)
	t.Cleanup(func() { _ = terminate(context.Background()) })

	createBucket(t, ctx, endpoint, minioBucket)

	builder := r2provider.NewR2Builder().WithOptions(r2provider.WithUsePathStyle(true))

	options := storage.NewProviderOptions(
		storage.WithBucket(minioBucket),
		storage.WithEndpoint(endpoint),
		storage.WithCredentials(storage.ProviderCredentials{
			AccountID:       "test-account",
			AccessKeyID:     minioUser,
			SecretAccessKey: minioSecret,
		}),
	)

	providerInterface, err := builder.Build(ctx, storage.ProviderCredentials{
		AccountID:       "test-account",
		AccessKeyID:     minioUser,
		SecretAccessKey: minioSecret,
	}, options)
	assert.NilError(t, err)
	provider := providerInterface
	t.Cleanup(func() { _ = provider.Close() })

	fileContent := []byte("standalone file content")

	// Upload file without organization folder destination
	uploadedFile, err := provider.Upload(ctx, bytes.NewReader(fileContent), &storagetypes.UploadFileOptions{
		FileName:          fileName,
		FolderDestination: "",
		ContentType:       pdftext.ContentType,
	})
	assert.NilError(t, err)
	assert.Assert(t, uploadedFile != nil)

	// Verify the key is just the filename (no organization prefix)
	assert.Assert(t, is.Equal(fileName, uploadedFile.Key), "file key should be just the filename when no folder is specified")
	assert.Assert(t, is.Equal("", uploadedFile.Folder), "folder should be empty")

	// Verify the full URI does not include organization path
	expectedURI := "r2://" + minioBucket + "/" + fileName
	assert.Assert(t, is.Equal(expectedURI, uploadedFile.FullURI), "full URI should not include organization path")
}
