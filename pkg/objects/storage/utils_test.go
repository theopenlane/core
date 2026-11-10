package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/pkg/pdftext"
	"github.com/theopenlane/core/v2/pkg/pdftext/pdftest"
)

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name           string
		data           []byte
		expectedMIME   string
		expectContains string
		expectError    bool
	}{
		{
			name:           "JSON file",
			data:           []byte(`{"key": "value"}`),
			expectedMIME:   "application/json",
			expectContains: "json",
		},
		{
			name:           "plain text",
			data:           []byte("hello world"),
			expectContains: "text",
		},
		{
			name:           "HTML content",
			data:           []byte("<html><body>test</body></html>"),
			expectContains: "html",
		},
		{
			name:           "XML content",
			data:           []byte(`<?xml version="1.0"?><root><item>test</item></root>`),
			expectContains: "xml",
		},
		{
			name:           "empty file",
			data:           []byte{},
			expectContains: "",
		},
		{
			name:           "binary data",
			data:           []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46},
			expectContains: "image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)
			mimeType, err := DetectContentType(reader)

			if tt.expectError {
				assert.Check(t, err != nil)
			} else {
				assert.NilError(t, err)
				if tt.expectedMIME != "" {
					assert.Check(t, is.Equal(tt.expectedMIME, mimeType))
				} else if tt.expectContains != "" {
					assert.Check(t, is.Contains(strings.ToLower(mimeType), tt.expectContains))
				}

				pos, err := reader.Seek(0, 1)
				assert.NilError(t, err)
				assert.Check(t, is.Equal(int64(0), pos), "reader should be seeked back to start")
			}
		})
	}
}

func TestDetectContentType_SeekError(t *testing.T) {
	t.Run("non-seekable reader", func(t *testing.T) {
		data := []byte("test data")
		reader := bytes.NewReader(data)

		_, err := reader.Seek(0, 0)
		assert.NilError(t, err)

		mimeType, err := DetectContentType(reader)
		assert.NilError(t, err)
		assert.Check(t, len(mimeType) > 0)
	})
}

func TestParseDocument(t *testing.T) {
	validPDF := pdftest.MinimalPDF(t)

	tests := []struct {
		name         string
		data         string
		mimeType     string
		expectType   string
		expectError  error
		expectFormat string
		validate     func(t *testing.T, result any)
	}{
		{
			name:       "JSON document",
			data:       `{"name": "test", "value": 123}`,
			mimeType:   "application/json",
			expectType: "map",
			validate: func(t *testing.T, result any) {
				m, ok := result.(map[string]any)
				assert.Assert(t, ok)
				assert.Check(t, is.Equal("test", m["name"]))
				assert.Check(t, is.Equal(float64(123), m["value"]))
			},
		},
		{
			name:       "JSON array",
			data:       `[1, 2, 3]`,
			mimeType:   "application/json",
			expectType: "slice",
			validate: func(t *testing.T, result any) {
				arr, ok := result.([]any)
				assert.Assert(t, ok)
				assert.Check(t, is.Len(arr, 3))
			},
		},
		{
			name:         "invalid JSON",
			data:         `{invalid json}`,
			mimeType:     "application/json",
			expectError:  ErrParseFailed,
			expectFormat: ParseFormatJSON,
		},
		{
			name:       "YAML document",
			data:       "name: test\nvalue: 123",
			mimeType:   "application/yaml",
			expectType: "map",
			validate: func(t *testing.T, result any) {
				m, ok := result.(map[string]any)
				assert.Assert(t, ok)
				assert.Check(t, is.Equal("test", m["name"]))
				assert.Check(t, is.Equal(123, m["value"]))
			},
		},
		{
			name:       "YAML with yml extension",
			data:       "key: value",
			mimeType:   "text/yml",
			expectType: "map",
			validate: func(t *testing.T, result any) {
				m, ok := result.(map[string]any)
				assert.Assert(t, ok)
				assert.Check(t, is.Equal("value", m["key"]))
			},
		},
		{
			name:         "invalid YAML",
			data:         "key: [unclosed",
			mimeType:     "application/yaml",
			expectError:  ErrParseFailed,
			expectFormat: ParseFormatYAML,
		},
		{
			name:       "plain text",
			data:       "hello world",
			mimeType:   "text/plain",
			expectType: "string",
			validate: func(t *testing.T, result any) {
				str, ok := result.(string)
				assert.Assert(t, ok, fmt.Sprintf("expected string, got %T", result))
				assert.Check(t, is.Equal("hello world", str))
			},
		},
		{
			name:       "binary data",
			data:       string([]byte{0xFF, 0xD8, 0xFF}),
			mimeType:   "image/jpeg",
			expectType: "[]uint8",
			validate: func(t *testing.T, result any) {
				data, ok := result.([]byte)
				assert.Assert(t, ok)
				assert.Check(t, is.DeepEqual([]byte{0xFF, 0xD8, 0xFF}, data))
			},
		},
		{
			name:       "PDF document",
			data:       string(validPDF),
			mimeType:   pdftext.ContentType,
			expectType: "string",
			validate: func(t *testing.T, result any) {
				str, ok := result.(string)
				assert.Assert(t, ok, fmt.Sprintf("expected string, got %T", result))
				assert.Check(t, is.Contains(str, "Original NDA"))
			},
		},
		{
			name:         "empty PDF",
			data:         "",
			mimeType:     pdftext.ContentType,
			expectError:  ErrParseFailed,
			expectFormat: ParseFormatPDF,
		},
		{
			name:         "not a PDF",
			data:         "definitely not a pdf",
			mimeType:     pdftext.ContentType,
			expectError:  ErrParseFailed,
			expectFormat: ParseFormatPDF,
		},
		{
			name:         "truncated PDF",
			data:         string(validPDF[:len(validPDF)/2]),
			mimeType:     pdftext.ContentType,
			expectError:  ErrParseFailed,
			expectFormat: ParseFormatPDF,
		},
		{
			name:         "PDF header only",
			data:         "%PDF-1.7\n%%EOF",
			mimeType:     pdftext.ContentType,
			expectError:  ErrParseFailed,
			expectFormat: ParseFormatPDF,
		},
		{
			name:       "unknown mime type returns raw data",
			data:       "some data",
			mimeType:   "application/octet-stream",
			expectType: "[]uint8",
			validate: func(t *testing.T, result any) {
				data, ok := result.([]byte)
				assert.Assert(t, ok)
				assert.Check(t, is.DeepEqual([]byte("some data"), data))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.data)
			result, err := ParseDocument(t.Context(), reader, tt.mimeType)

			if tt.expectError != nil {
				assert.Assert(t, err != nil)
				assert.Check(t, is.ErrorIs(err, tt.expectError))

				if tt.expectFormat != "" {
					var parseErr *ParseError
					assert.Assert(t, errors.As(err, &parseErr))
					assert.Check(t, is.Equal(tt.expectFormat, parseErr.Format))
				}
			} else {
				assert.NilError(t, err)
				if tt.validate != nil {
					tt.validate(t, result.Data)
				}
			}
		})
	}
}

func TestParseDocument_ComplexStructures(t *testing.T) {
	t.Run("nested JSON", func(t *testing.T) {
		data := `{
			"user": {
				"name": "John",
				"age": 30,
				"addresses": [
					{"city": "New York"},
					{"city": "Los Angeles"}
				]
			}
		}`

		result, err := ParseDocument(t.Context(), strings.NewReader(data), "application/json")
		assert.NilError(t, err)

		m, ok := result.Data.(map[string]any)
		assert.Assert(t, ok)

		user, ok := m["user"].(map[string]any)
		assert.Assert(t, ok)
		assert.Check(t, is.Equal("John", user["name"]))
		assert.Check(t, is.Equal(float64(30), user["age"]))

		addresses, ok := user["addresses"].([]any)
		assert.Assert(t, ok)
		assert.Check(t, is.Len(addresses, 2))
	})

	t.Run("nested YAML", func(t *testing.T) {
		data := `
user:
  name: Jane
  age: 25
  hobbies:
    - reading
    - coding
`
		result, err := ParseDocument(t.Context(), strings.NewReader(data), "application/yaml")
		assert.NilError(t, err)

		m, ok := result.Data.(map[string]any)
		assert.Assert(t, ok)

		user, ok := m["user"].(map[string]any)
		assert.Assert(t, ok)
		assert.Check(t, is.Equal("Jane", user["name"]))
		assert.Check(t, is.Equal(25, user["age"]))

		hobbies, ok := user["hobbies"].([]any)
		assert.Assert(t, ok)
		assert.Check(t, is.Len(hobbies, 2))
	})
}

func TestNewUploadFile(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test-*.txt")
		assert.NilError(t, err)
		defer os.Remove(tempFile.Name())

		content := "test content"
		_, err = tempFile.WriteString(content)
		assert.NilError(t, err)
		tempFile.Close()

		file, err := NewUploadFile(tempFile.Name())
		assert.NilError(t, err)
		assert.Assert(t, file != nil)

		assert.Check(t, is.Equal(filepath.Base(tempFile.Name()), file.OriginalName))
		assert.Check(t, is.Equal(int64(len(content)), file.Size))
		assert.Check(t, len(file.ContentType) > 0)
		assert.Check(t, is.Equal("file_upload", file.Key))
		assert.Check(t, file.RawFile != nil)

		file.RawFile.(*os.File).Close()
	})

	t.Run("JSON file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test-*.json")
		assert.NilError(t, err)
		defer os.Remove(tempFile.Name())

		content := map[string]string{"key": "value"}
		data, err := json.Marshal(content)
		assert.NilError(t, err)

		_, err = tempFile.Write(data)
		assert.NilError(t, err)
		tempFile.Close()

		file, err := NewUploadFile(tempFile.Name())
		assert.NilError(t, err)
		assert.Assert(t, file != nil)

		assert.Check(t, is.Contains(file.ContentType, "json"))

		file.RawFile.(*os.File).Close()
	})

	t.Run("YAML file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test-*.yaml")
		assert.NilError(t, err)
		defer os.Remove(tempFile.Name())

		content := map[string]string{"key": "value"}
		data, err := yaml.Marshal(content)
		assert.NilError(t, err)

		_, err = tempFile.Write(data)
		assert.NilError(t, err)
		tempFile.Close()

		file, err := NewUploadFile(tempFile.Name())
		assert.NilError(t, err)
		assert.Assert(t, file != nil)

		assert.Check(t, len(file.ContentType) > 0)

		file.RawFile.(*os.File).Close()
	})

	t.Run("binary file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test-*.bin")
		assert.NilError(t, err)
		defer os.Remove(tempFile.Name())

		binaryData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
		_, err = tempFile.Write(binaryData)
		assert.NilError(t, err)
		tempFile.Close()

		file, err := NewUploadFile(tempFile.Name())
		assert.NilError(t, err)
		assert.Assert(t, file != nil)

		assert.Check(t, is.Contains(file.ContentType, "image"))

		file.RawFile.(*os.File).Close()
	})

	t.Run("empty file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test-*.txt")
		assert.NilError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		file, err := NewUploadFile(tempFile.Name())
		assert.NilError(t, err)
		assert.Assert(t, file != nil)

		assert.Check(t, is.Equal(int64(0), file.Size))
		assert.Check(t, len(file.ContentType) > 0)

		file.RawFile.(*os.File).Close()
	})

	t.Run("non-existent file", func(t *testing.T) {
		file, err := NewUploadFile("/nonexistent/path/file.txt")
		assert.Check(t, err != nil)
		assert.Check(t, is.Nil(file))
	})

	t.Run("directory instead of file", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "test-dir-*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		file, err := NewUploadFile(tempDir)
		assert.Check(t, err)
		assert.Check(t, file != nil)

		file.RawFile.(*os.File).Close()
	})

	t.Run("file with special characters in name", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test-file-with-spaces-*.txt")
		assert.NilError(t, err)
		defer os.Remove(tempFile.Name())

		_, err = tempFile.WriteString("test content")
		assert.NilError(t, err)
		tempFile.Close()

		file, err := NewUploadFile(tempFile.Name())
		assert.NilError(t, err)
		assert.Assert(t, file != nil)

		assert.Check(t, is.Contains(file.OriginalName, "test-file-with-spaces"))

		file.RawFile.(*os.File).Close()
	})
}

func TestNewUploadFile_WithExistingTestData(t *testing.T) {
	testDataPath := filepath.Join("..", "testdata", "image.jpg")

	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Skip("test data file does not exist")
	}

	t.Run("image.jpg from testdata", func(t *testing.T) {
		file, err := NewUploadFile(testDataPath)
		assert.NilError(t, err)
		assert.Assert(t, file != nil)

		assert.Check(t, is.Equal("image.jpg", file.OriginalName))
		assert.Check(t, file.Size > int64(0))
		assert.Check(t, is.Contains(strings.ToLower(file.ContentType), "image"))

		file.RawFile.(*os.File).Close()
	})
}

func TestDetectContentType_ReaderPosition(t *testing.T) {
	t.Run("reader position is restored", func(t *testing.T) {
		data := []byte("test data for position check")
		reader := bytes.NewReader(data)

		_, err := reader.Seek(5, 0)
		assert.NilError(t, err)

		_, err = DetectContentType(reader)
		assert.NilError(t, err)

		pos, err := reader.Seek(0, 1)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(int64(0), pos))
	})
}

func TestParseDocument_EmptyContent(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		mimeType string
	}{
		{"empty JSON", "", "application/json"},
		{"empty YAML", "", "application/yaml"},
		{"empty text", "", "text/plain"},
		{"empty binary", "", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.data)
			result, err := ParseDocument(t.Context(), reader, tt.mimeType)

			if strings.Contains(tt.mimeType, "json") {
				assert.Assert(t, err != nil)
			} else if strings.Contains(tt.mimeType, "yaml") {
				assert.NilError(t, err)
				assert.Check(t, is.Nil(result.Data))
			} else if strings.Contains(tt.mimeType, "text") {
				assert.NilError(t, err)
				assert.Check(t, is.Equal("", result.Data))
			} else {
				assert.NilError(t, err)
				assert.Check(t, is.DeepEqual([]byte{}, result.Data))
			}
		})
	}
}
