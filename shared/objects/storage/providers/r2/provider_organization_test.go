package r2_test

import (
	"bytes"
	"context"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	storage "github.com/theopenlane/shared/objects/storage"
	r2provider "github.com/theopenlane/shared/objects/storage/providers/r2"
	storagetypes "github.com/theopenlane/shared/objects/storage/types"
)

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
	require.NoError(t, err)
	provider := providerInterface
	t.Cleanup(func() { _ = provider.Close() })

	fileContent := []byte("test content for organization folder structure")

	// Upload file with hierarchical folder destination
	uploadedFile, err := provider.Upload(ctx, bytes.NewReader(fileContent), &storagetypes.UploadFileOptions{
		FileName:          fileName,
		FolderDestination: folder,
		ContentType:       "application/pdf",
	})
	require.NoError(t, err)
	require.NotNil(t, uploadedFile)

	// Verify the key includes the organization/parent/file hierarchy
	expectedKey := path.Join(folder, fileName)
	require.Equal(t, expectedKey, uploadedFile.Key, "file key should be orgID/parentID/fileID/filename")
	require.Equal(t, folder, uploadedFile.Folder, "folder should be set to orgID/parentID/fileID")

	// Verify the full URI reflects the organization structure
	expectedURI := "r2://" + path.Join(minioBucket, folder, fileName)
	require.Equal(t, expectedURI, uploadedFile.FullURI, "full URI should include hierarchical path")

	// Verify file can be downloaded using the organization-prefixed key
	downloadedFile, err := provider.Download(ctx, &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			Key:    expectedKey,
			Bucket: minioBucket,
		},
	}, &storagetypes.DownloadFileOptions{})
	require.NoError(t, err)
	require.Equal(t, fileContent, downloadedFile.File)
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
	require.NoError(t, err)
	provider := providerInterface
	t.Cleanup(func() { _ = provider.Close() })

	fileContent := []byte("standalone file content")

	// Upload file without organization folder destination
	uploadedFile, err := provider.Upload(ctx, bytes.NewReader(fileContent), &storagetypes.UploadFileOptions{
		FileName:          fileName,
		FolderDestination: "",
		ContentType:       "application/pdf",
	})
	require.NoError(t, err)
	require.NotNil(t, uploadedFile)

	// Verify the key is just the filename (no organization prefix)
	require.Equal(t, fileName, uploadedFile.Key, "file key should be just the filename when no folder is specified")
	require.Equal(t, "", uploadedFile.Folder, "folder should be empty")

	// Verify the full URI does not include organization path
	expectedURI := "r2://" + minioBucket + "/" + fileName
	require.Equal(t, expectedURI, uploadedFile.FullURI, "full URI should not include organization path")
}
