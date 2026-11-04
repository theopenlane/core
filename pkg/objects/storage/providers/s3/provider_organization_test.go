package s3_test

import (
	"bytes"
	"context"
	"path"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	s3provider "github.com/theopenlane/core/v2/pkg/objects/storage/providers/s3"
	"github.com/theopenlane/core/v2/pkg/pdftext"
)

// TestS3Provider_OrganizationFolderStructure verifies that files are stored under organization-specific directories
func TestS3Provider_OrganizationFolderStructure(t *testing.T) {
	ctx := context.Background()
	orgID := "01HYQZ5YTVJ0P2R2HF7N3W3MQZ"
	parentID := "01J1PARENTXYZABCD1234"
	fileID := "01J1FILEXYZABCD5678"
	fileName := "document.pdf"
	folder := path.Join(orgID, parentID, fileID)

	endpoint, terminate := startMinio(t, ctx)
	t.Cleanup(func() { _ = terminate(context.Background()) })

	createBucket(t, ctx, endpoint, minioBucket)

	builder := s3provider.NewS3Builder().WithOptions(s3provider.WithUsePathStyle(true))

	options := storage.NewProviderOptions(
		storage.WithBucket(minioBucket),
		storage.WithEndpoint(endpoint),
		storage.WithRegion(testRegion),
		storage.WithCredentials(storage.ProviderCredentials{
			AccessKeyID:     minioUser,
			SecretAccessKey: minioSecret,
		}),
	)

	providerInterface, err := builder.Build(ctx, storage.ProviderCredentials{
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
	expectedURI := "s3://" + path.Join(minioBucket, folder, fileName)
	assert.Assert(t, is.Equal(expectedURI, uploadedFile.FullURI), "full URI should include hierarchical path")

	// Verify file can be downloaded using the organization-prefixed key
	downloadedFile, err := provider.Download(ctx, &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			Key:    expectedKey,
			Bucket: minioBucket,
		},
	}, &storagetypes.DownloadFileOptions{})
	assert.NilError(t, err)
	assert.Assert(t, is.DeepEqual(fileContent, downloadedFile.File))
}

// TestS3Provider_WithoutOrganizationFolder verifies behavior when no organization folder is specified
func TestS3Provider_WithoutOrganizationFolder(t *testing.T) {
	ctx := context.Background()
	fileName := "standalone-document.pdf"

	endpoint, terminate := startMinio(t, ctx)
	t.Cleanup(func() { _ = terminate(context.Background()) })

	createBucket(t, ctx, endpoint, minioBucket)

	builder := s3provider.NewS3Builder().WithOptions(s3provider.WithUsePathStyle(true))

	options := storage.NewProviderOptions(
		storage.WithBucket(minioBucket),
		storage.WithEndpoint(endpoint),
		storage.WithRegion(testRegion),
		storage.WithCredentials(storage.ProviderCredentials{
			AccessKeyID:     minioUser,
			SecretAccessKey: minioSecret,
		}),
	)

	providerInterface, err := builder.Build(ctx, storage.ProviderCredentials{
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
	expectedURI := "s3://" + minioBucket + "/" + fileName
	assert.Assert(t, is.Equal(expectedURI, uploadedFile.FullURI), "full URI should not include organization path")
}
