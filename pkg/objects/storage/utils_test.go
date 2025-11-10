package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.expectedMIME != "" {
					assert.Equal(t, tt.expectedMIME, mimeType)
				} else if tt.expectContains != "" {
					assert.Contains(t, strings.ToLower(mimeType), tt.expectContains)
				}

				pos, err := reader.Seek(0, 1)
				require.NoError(t, err)
				assert.Equal(t, int64(0), pos, "reader should be seeked back to start")
			}
		})
	}
}

func TestDetectContentType_SeekError(t *testing.T) {
	t.Run("non-seekable reader", func(t *testing.T) {
		data := []byte("test data")
		reader := bytes.NewReader(data)

		_, err := reader.Seek(0, 0)
		require.NoError(t, err)

		mimeType, err := DetectContentType(reader)
		require.NoError(t, err)
		assert.NotEmpty(t, mimeType)
	})
}

func TestParseDocument(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		mimeType    string
		expectType  string
		expectError error
		validate    func(t *testing.T, result any)
	}{
		{
			name:       "JSON document",
			data:       `{"name": "test", "value": 123}`,
			mimeType:   "application/json",
			expectType: "map",
			validate: func(t *testing.T, result any) {
				m, ok := result.(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "test", m["name"])
				assert.Equal(t, float64(123), m["value"])
			},
		},
		{
			name:       "JSON array",
			data:       `[1, 2, 3]`,
			mimeType:   "application/json",
			expectType: "slice",
			validate: func(t *testing.T, result any) {
				arr, ok := result.([]any)
				require.True(t, ok)
				assert.Len(t, arr, 3)
			},
		},
		{
			name:        "invalid JSON",
			data:        `{invalid json}`,
			mimeType:    "application/json",
			expectError: ErrJSONParseFailed,
		},
		{
			name:       "YAML document",
			data:       "name: test\nvalue: 123",
			mimeType:   "application/yaml",
			expectType: "map",
			validate: func(t *testing.T, result any) {
				m, ok := result.(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "test", m["name"])
				assert.Equal(t, 123, m["value"])
			},
		},
		{
			name:       "YAML with yml extension",
			data:       "key: value",
			mimeType:   "text/yml",
			expectType: "map",
			validate: func(t *testing.T, result any) {
				m, ok := result.(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "value", m["key"])
			},
		},
		{
			name:        "invalid YAML",
			data:        "key: [unclosed",
			mimeType:    "application/yaml",
			expectError: ErrYAMLParseFailed,
		},
		{
			name:       "plain text",
			data:       "hello world",
			mimeType:   "text/plain",
			expectType: "string",
			validate: func(t *testing.T, result any) {
				str, ok := result.(string)
				require.True(t, ok, fmt.Sprintf("expected string, got %T", result))
				assert.Equal(t, "hello world", str)
			},
		},
		{
			name:       "binary data",
			data:       string([]byte{0xFF, 0xD8, 0xFF}),
			mimeType:   "image/jpeg",
			expectType: "[]uint8",
			validate: func(t *testing.T, result any) {
				data, ok := result.([]byte)
				require.True(t, ok)
				assert.Equal(t, []byte{0xFF, 0xD8, 0xFF}, data)
			},
		},
		{
			name:       "unknown mime type returns raw data",
			data:       "some data",
			mimeType:   "application/octet-stream",
			expectType: "[]uint8",
			validate: func(t *testing.T, result any) {
				data, ok := result.([]byte)
				require.True(t, ok)
				assert.Equal(t, []byte("some data"), data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.data)
			result, err := ParseDocument(reader, tt.mimeType)

			if tt.expectError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectError)
			} else {
				require.NoError(t, err)
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

		result, err := ParseDocument(strings.NewReader(data), "application/json")
		require.NoError(t, err)

		m, ok := result.Data.(map[string]any)
		require.True(t, ok)

		user, ok := m["user"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "John", user["name"])
		assert.Equal(t, float64(30), user["age"])

		addresses, ok := user["addresses"].([]any)
		require.True(t, ok)
		assert.Len(t, addresses, 2)
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
		result, err := ParseDocument(strings.NewReader(data), "application/yaml")
		require.NoError(t, err)

		m, ok := result.Data.(map[string]any)
		require.True(t, ok)

		user, ok := m["user"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "Jane", user["name"])
		assert.Equal(t, 25, user["age"])

		hobbies, ok := user["hobbies"].([]any)
		require.True(t, ok)
		assert.Len(t, hobbies, 2)
	})
}

func TestNewUploadFile(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test-*.txt")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())

		content := "test content"
		_, err = tempFile.WriteString(content)
		require.NoError(t, err)
		tempFile.Close()

		file, err := NewUploadFile(tempFile.Name())
		require.NoError(t, err)
		require.NotNil(t, file)

		assert.Equal(t, filepath.Base(tempFile.Name()), file.OriginalName)
		assert.Equal(t, int64(len(content)), file.Size)
		assert.NotEmpty(t, file.ContentType)
		assert.Equal(t, "file_upload", file.Key)
		assert.NotNil(t, file.RawFile)

		file.RawFile.(*os.File).Close()
	})

	t.Run("JSON file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test-*.json")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())

		content := map[string]string{"key": "value"}
		data, err := json.Marshal(content)
		require.NoError(t, err)

		_, err = tempFile.Write(data)
		require.NoError(t, err)
		tempFile.Close()

		file, err := NewUploadFile(tempFile.Name())
		require.NoError(t, err)
		require.NotNil(t, file)

		assert.Contains(t, file.ContentType, "json")

		file.RawFile.(*os.File).Close()
	})

	t.Run("YAML file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())

		content := map[string]string{"key": "value"}
		data, err := yaml.Marshal(content)
		require.NoError(t, err)

		_, err = tempFile.Write(data)
		require.NoError(t, err)
		tempFile.Close()

		file, err := NewUploadFile(tempFile.Name())
		require.NoError(t, err)
		require.NotNil(t, file)

		assert.NotEmpty(t, file.ContentType)

		file.RawFile.(*os.File).Close()
	})

	t.Run("binary file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test-*.bin")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())

		binaryData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
		_, err = tempFile.Write(binaryData)
		require.NoError(t, err)
		tempFile.Close()

		file, err := NewUploadFile(tempFile.Name())
		require.NoError(t, err)
		require.NotNil(t, file)

		assert.Contains(t, file.ContentType, "image")

		file.RawFile.(*os.File).Close()
	})

	t.Run("empty file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test-*.txt")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		file, err := NewUploadFile(tempFile.Name())
		require.NoError(t, err)
		require.NotNil(t, file)

		assert.Equal(t, int64(0), file.Size)
		assert.NotEmpty(t, file.ContentType)

		file.RawFile.(*os.File).Close()
	})

	t.Run("non-existent file", func(t *testing.T) {
		file, err := NewUploadFile("/nonexistent/path/file.txt")
		assert.Error(t, err)
		assert.Nil(t, file)
	})

	t.Run("directory instead of file", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "test-dir-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		file, err := NewUploadFile(tempDir)
		assert.NoError(t, err)
		assert.NotNil(t, file)

		file.RawFile.(*os.File).Close()
	})

	t.Run("file with special characters in name", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test-file-with-spaces-*.txt")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())

		_, err = tempFile.WriteString("test content")
		require.NoError(t, err)
		tempFile.Close()

		file, err := NewUploadFile(tempFile.Name())
		require.NoError(t, err)
		require.NotNil(t, file)

		assert.Contains(t, file.OriginalName, "test-file-with-spaces")

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
		require.NoError(t, err)
		require.NotNil(t, file)

		assert.Equal(t, "image.jpg", file.OriginalName)
		assert.Greater(t, file.Size, int64(0))
		assert.Contains(t, strings.ToLower(file.ContentType), "image")

		file.RawFile.(*os.File).Close()
	})
}

func TestDetectContentType_ReaderPosition(t *testing.T) {
	t.Run("reader position is restored", func(t *testing.T) {
		data := []byte("test data for position check")
		reader := bytes.NewReader(data)

		_, err := reader.Seek(5, 0)
		require.NoError(t, err)

		_, err = DetectContentType(reader)
		require.NoError(t, err)

		pos, err := reader.Seek(0, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(0), pos)
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
			result, err := ParseDocument(reader, tt.mimeType)

			if strings.Contains(tt.mimeType, "json") {
				require.Error(t, err)
			} else if strings.Contains(tt.mimeType, "yaml") {
				require.NoError(t, err)
				assert.Nil(t, result.Data)
			} else if strings.Contains(tt.mimeType, "text") {
				require.NoError(t, err)
				assert.Equal(t, "", result.Data)
			} else {
				require.NoError(t, err)
				assert.Equal(t, []byte{}, result.Data)
			}
		})
	}
}
