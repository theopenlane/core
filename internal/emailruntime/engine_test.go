package emailruntime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/generated"
)

func TestValidateTemplateData_EmptySchema(t *testing.T) {
	err := validateTemplateData(nil, map[string]any{"foo": "bar"})
	require.NoError(t, err)
}

func TestValidateTemplateData_EmptyPayload(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
	}

	err := validateTemplateData(schema, map[string]any{})
	require.NoError(t, err)
}

func TestValidateTemplateData_ValidData(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
		"required": []any{"name"},
	}

	err := validateTemplateData(schema, map[string]any{"name": "Alice"})
	require.NoError(t, err)
}

func TestValidateTemplateData_MissingRequired(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
		"required": []any{"name"},
	}

	err := validateTemplateData(schema, map[string]any{})
	require.ErrorIs(t, err, ErrTemplateDataInvalid)
}

func TestValidateTemplateData_WrongType(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"count": map[string]any{"type": "integer"},
		},
		"required": []any{"count"},
	}

	err := validateTemplateData(schema, map[string]any{"count": "not-an-int"})
	require.ErrorIs(t, err, ErrTemplateDataInvalid)
}

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
