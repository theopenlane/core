package email

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/generated"
)

func TestStaticAttachmentsFromFiles_EmptyList(t *testing.T) {
	attachments := staticAttachmentsFromFiles(context.Background(), nil)
	assert.Empty(t, attachments)
}

func TestStaticAttachmentsFromFiles_SkipsEmptyContent(t *testing.T) {
	files := []*generated.File{
		{ProvidedFileName: "empty", FileContents: nil},
	}

	attachments := staticAttachmentsFromFiles(context.Background(), files)
	assert.Empty(t, attachments)
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

	require.Len(t, attachments, 1)
	assert.Equal(t, "report.pdf", attachments[0].Filename)
	assert.Equal(t, "application/pdf", attachments[0].ContentType)
	assert.Equal(t, []byte("pdf-content"), attachments[0].Content)
}

func TestStaticAttachmentsFromFiles_WithoutExtension(t *testing.T) {
	files := []*generated.File{
		{
			ProvidedFileName: "invoice",
			FileContents:     []byte("data"),
		},
	}

	attachments := staticAttachmentsFromFiles(context.Background(), files)

	require.Len(t, attachments, 1)
	assert.Equal(t, "invoice", attachments[0].Filename)
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

	require.Len(t, attachments, 1)
	assert.Equal(t, "doc.txt", attachments[0].Filename)
}
