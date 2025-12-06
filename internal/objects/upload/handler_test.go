package upload

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/models"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/iam/auth"
)

func TestBuildUploadOptionsInitialisesHints(t *testing.T) {
	file := &pkgobjects.File{
		OriginalName:         "report.pdf",
		FieldName:            "evidence",
		CorrelatedObjectType: "trustCenter",
		Metadata: map[string]string{
			"custom": "value",
		},
		FileMetadata: pkgobjects.FileMetadata{
			Bucket: "bucket",
			Folder: "folder",
		},
	}

	ctx := auth.WithAuthenticatedUser(context.Background(), &auth.AuthenticatedUser{
		OrganizationID: "01HYQZ5YTVJ0P2R2HF7N3W3MQZ",
	})

	opts := BuildUploadOptions(ctx, file)
	assert.NotNil(t, opts)
	assert.Equal(t, file.OriginalName, opts.FileName)
	assert.Equal(t, file.FieldName, opts.FileMetadata.Key)
	assert.NotNil(t, opts.ProviderHints)
	moduleValue, ok := opts.ProviderHints.Module.(models.OrgModule)
	assert.True(t, ok)
	assert.Equal(t, models.CatalogTrustCenterModule, moduleValue)
}

func TestBuildUploadOptionsDetectsContentType(t *testing.T) {
	content := "Sample text payload"
	file := &pkgobjects.File{
		OriginalName: "notes.txt",
		FieldName:    "notes",
		RawFile:      bytes.NewReader([]byte(content)),
		FileMetadata: pkgobjects.FileMetadata{ContentType: ""},
	}

	opts := BuildUploadOptions(context.Background(), file)
	assert.Equal(t, "text/plain; charset=utf-8", opts.ContentType)
	assert.Equal(t, "text/plain; charset=utf-8", file.ContentType)
}

func TestBuildUploadOptionsBufferedDetection(t *testing.T) {
	file := &pkgobjects.File{
		OriginalName: "data.bin",
		FieldName:    "upload",
		RawFile:      strings.NewReader("000000"),
		FileMetadata: pkgobjects.FileMetadata{ContentType: "application/octet-stream"},
	}

	opts := BuildUploadOptions(context.Background(), file)
	assert.NotNil(t, opts)
	assert.Equal(t, file.OriginalName, opts.FileName)
	assert.NotNil(t, opts.ProviderHints)
}

func TestBuildUploadOptionsNilFile(t *testing.T) {
	opts := BuildUploadOptions(context.Background(), nil)
	assert.NotNil(t, opts)
	assert.Empty(t, opts.FileName)
	assert.Empty(t, opts.ContentType)
}

func TestBuildUploadOptionsPreservesExplicitContentType(t *testing.T) {
	file := &pkgobjects.File{
		OriginalName: "photo.jpg",
		FieldName:    "avatarFile",
		FileMetadata: pkgobjects.FileMetadata{ContentType: "image/jpeg"},
	}

	opts := BuildUploadOptions(context.Background(), file)
	assert.Equal(t, "image/jpeg", opts.ContentType)
	assert.Equal(t, "image/jpeg", file.ContentType)
}

func TestHandleUploadsNoFiles(t *testing.T) {
	ctx := context.Background()
	newCtx, files, err := HandleUploads(ctx, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, ctx, newCtx)
	assert.Nil(t, files)
}

func TestMergeUploadedFileMetadata(t *testing.T) {
	src := pkgobjects.File{
		FieldName:            "avatar",
		Parent:               storage.ParentObject{ID: "user-id", Type: "user"},
		CorrelatedObjectID:   "object-id",
		CorrelatedObjectType: "user",
		Metadata: map[string]string{
			"keep": "yes",
		},
	}

	dest := &pkgobjects.File{
		Metadata: map[string]string{},
	}

	mergeUploadedFileMetadata(dest, "file-id", src)

	assert.Equal(t, "file-id", dest.ID)
	assert.Equal(t, src.FieldName, dest.FieldName)
	assert.Equal(t, src.Parent, dest.Parent)
	assert.Equal(t, src.CorrelatedObjectID, dest.CorrelatedObjectID)
	assert.Equal(t, src.CorrelatedObjectType, dest.CorrelatedObjectType)
	assert.NotEmpty(t, dest.Metadata)
}
