package upload

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

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
	require.NotNil(t, opts)
	require.Equal(t, file.OriginalName, opts.FileName)
	require.Equal(t, file.FieldName, opts.FileMetadata.Key)
	require.NotNil(t, opts.ProviderHints)
	moduleValue, ok := opts.ProviderHints.Module.(models.OrgModule)
	require.True(t, ok)
	require.Equal(t, models.CatalogTrustCenterModule, moduleValue)
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
	require.Equal(t, "text/plain; charset=utf-8", opts.ContentType)
	require.Equal(t, "text/plain; charset=utf-8", file.ContentType)
}

func TestBuildUploadOptionsBufferedDetection(t *testing.T) {
	file := &pkgobjects.File{
		OriginalName: "data.bin",
		FieldName:    "upload",
		RawFile:      strings.NewReader("000000"),
		FileMetadata: pkgobjects.FileMetadata{ContentType: "application/octet-stream"},
	}

	opts := BuildUploadOptions(context.Background(), file)
	require.NotNil(t, opts)
	require.Equal(t, file.OriginalName, opts.FileName)
	require.NotNil(t, opts.ProviderHints)
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

	require.Equal(t, "file-id", dest.ID)
	require.Equal(t, src.FieldName, dest.FieldName)
	require.Equal(t, src.Parent, dest.Parent)
	require.Equal(t, src.CorrelatedObjectID, dest.CorrelatedObjectID)
	require.Equal(t, src.CorrelatedObjectType, dest.CorrelatedObjectType)
	require.NotEmpty(t, dest.Metadata)
}
