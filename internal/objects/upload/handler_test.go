package upload

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/models"
	pkgobjects "github.com/theopenlane/core/v2/pkg/objects"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/iam/auth"
)

func TestBuildUploadOptionsInitialisesHints(t *testing.T) {
	file := &pkgobjects.File{
		OriginalName:         "report.pdf",
		FieldName:            "evidence",
		CorrelatedObjectType: "trustCenter",
		Metadata: map[string]any{
			"custom": "value",
		},
		FileMetadata: pkgobjects.FileMetadata{
			Bucket: "bucket",
			Folder: "folder",
		},
	}

	ctx := auth.WithCaller(context.Background(), &auth.Caller{
		OrganizationID: "01HYQZ5YTVJ0P2R2HF7N3W3MQZ",
	})

	opts := BuildUploadOptions(ctx, file)
	assert.Check(t, opts != nil)
	assert.Check(t, is.Equal(file.OriginalName, opts.FileName))
	assert.Check(t, is.Equal(file.FieldName, opts.FileMetadata.Key))
	assert.Check(t, opts.ProviderHints != nil)
	moduleValue, ok := opts.ProviderHints.Module.(models.OrgModule)
	assert.Check(t, ok)
	assert.Check(t, is.Equal(models.CatalogTrustCenterModule, moduleValue))
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
	assert.Check(t, is.Equal("text/plain; charset=utf-8", opts.ContentType))
	assert.Check(t, is.Equal("text/plain; charset=utf-8", file.ContentType))
}

func TestBuildUploadOptionsBufferedDetection(t *testing.T) {
	file := &pkgobjects.File{
		OriginalName: "data.bin",
		FieldName:    "upload",
		RawFile:      strings.NewReader("000000"),
		FileMetadata: pkgobjects.FileMetadata{ContentType: "application/octet-stream"},
	}

	opts := BuildUploadOptions(context.Background(), file)
	assert.Check(t, opts != nil)
	assert.Check(t, is.Equal(file.OriginalName, opts.FileName))
	assert.Check(t, opts.ProviderHints != nil)
}

func TestBuildUploadOptionsNilFile(t *testing.T) {
	opts := BuildUploadOptions(context.Background(), nil)
	assert.Check(t, opts != nil)
	assert.Check(t, is.Equal(opts.FileName, ""))
	assert.Check(t, is.Equal(opts.ContentType, ""))
}

func TestBuildUploadOptionsPreservesExplicitContentType(t *testing.T) {
	file := &pkgobjects.File{
		OriginalName: "photo.jpg",
		FieldName:    "avatarFile",
		FileMetadata: pkgobjects.FileMetadata{ContentType: "image/jpeg"},
	}

	opts := BuildUploadOptions(context.Background(), file)
	assert.Check(t, is.Equal("image/jpeg", opts.ContentType))
	assert.Check(t, is.Equal("image/jpeg", file.ContentType))
}

func TestHandleUploadsNoFiles(t *testing.T) {
	ctx := context.Background()
	newCtx, files, err := HandleUploads(ctx, nil, nil)
	assert.Check(t, err)
	assert.Check(t, is.Equal(ctx, newCtx))
	assert.Check(t, is.Nil(files))
}

func TestMergeUploadedFileMetadata(t *testing.T) {
	src := pkgobjects.File{
		FieldName:            "avatar",
		Parent:               storage.ParentObject{ID: "user-id", Type: "user"},
		CorrelatedObjectID:   "object-id",
		CorrelatedObjectType: "user",
		Metadata: map[string]any{
			"keep": "yes",
		},
	}

	dest := &pkgobjects.File{
		Metadata: map[string]any{},
	}

	mergeUploadedFileMetadata(dest, "file-id", src)

	assert.Check(t, is.Equal("file-id", dest.ID))
	assert.Check(t, is.Equal(src.FieldName, dest.FieldName))
	assert.Check(t, is.Equal(src.Parent, dest.Parent))
	assert.Check(t, is.Equal(src.CorrelatedObjectID, dest.CorrelatedObjectID))
	assert.Check(t, is.Equal(src.CorrelatedObjectType, dest.CorrelatedObjectType))
	assert.Check(t, len(dest.Metadata) > 0)
}
