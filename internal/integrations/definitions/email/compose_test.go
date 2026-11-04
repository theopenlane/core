package email

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
)

func TestStaticAttachmentsFromFiles_EmptyList(t *testing.T) {
	attachments := staticAttachmentsFromFiles(context.Background(), nil)
	assert.Check(t, is.Len(attachments, 0))
}

func TestStaticAttachmentsFromFiles_SkipsEmptyContent(t *testing.T) {
	files := []*generated.File{
		{ProvidedFileName: "empty", FileContents: nil},
	}

	attachments := staticAttachmentsFromFiles(context.Background(), files)
	assert.Check(t, is.Len(attachments, 0))
}

func TestStaticAttachmentsFromFiles_WithExtension(t *testing.T) {
	files := []*generated.File{
		{
			ProvidedFileName:      "report",
			ProvidedFileExtension: "pdf",
			DetectedMimeType:      "application/pdf",
			FileContents:          []byte("pdf-content"),
		},
	}

	attachments := staticAttachmentsFromFiles(context.Background(), files)

	assert.Assert(t, is.Len(attachments, 1))
	assert.Check(t, is.Equal("report.pdf", attachments[0].Filename))
	assert.Check(t, is.Equal("application/pdf", attachments[0].ContentType))
	assert.Check(t, is.DeepEqual([]byte("pdf-content"), attachments[0].Content))
}

func TestStaticAttachmentsFromFiles_WithoutExtension(t *testing.T) {
	files := []*generated.File{
		{
			ProvidedFileName: "invoice",
			FileContents:     []byte("data"),
		},
	}

	attachments := staticAttachmentsFromFiles(context.Background(), files)

	assert.Assert(t, is.Len(attachments, 1))
	assert.Check(t, is.Equal("invoice", attachments[0].Filename))
}

func TestStaticAttachmentsFromFiles_MixedContent(t *testing.T) {
	files := []*generated.File{
		{ProvidedFileName: "empty", FileContents: nil},
		{
			ProvidedFileName:      "doc",
			ProvidedFileExtension: "txt",
			FileContents:          []byte("text"),
		},
	}

	attachments := staticAttachmentsFromFiles(context.Background(), files)

	assert.Assert(t, is.Len(attachments, 1))
	assert.Check(t, is.Equal("doc.txt", attachments[0].Filename))
}
