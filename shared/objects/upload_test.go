package objects

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/shared/objects/mocks"
	"github.com/theopenlane/shared/objects/storage"
)

func TestExtractUploads(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int
	}{
		{
			name: "single upload",
			input: graphql.Upload{
				Filename: "test.txt",
				Size:     100,
			},
			expected: 1,
		},
		{
			name: "slice of uploads",
			input: []graphql.Upload{
				{Filename: "file1.txt", Size: 100},
				{Filename: "file2.txt", Size: 200},
			},
			expected: 2,
		},
		{
			name: "slice of any with uploads",
			input: []any{
				graphql.Upload{Filename: "file1.txt", Size: 100},
				graphql.Upload{Filename: "file2.txt", Size: 200},
				"not an upload",
			},
			expected: 2,
		},
		{
			name: "map with uploads",
			input: map[string]any{
				"file1": graphql.Upload{Filename: "file1.txt", Size: 100},
				"file2": graphql.Upload{Filename: "file2.txt", Size: 200},
			},
			expected: 2,
		},
		{
			name:     "empty slice",
			input:    []graphql.Upload{},
			expected: 0,
		},
		{
			name:     "nil value",
			input:    nil,
			expected: 0,
		},
		{
			name:     "unsupported type",
			input:    "string",
			expected: 0,
		},
		{
			name:     "nested map",
			input:    map[string]any{"nested": map[string]any{"file": graphql.Upload{Filename: "nested.txt", Size: 100}}},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractUploads(tt.input)
			assert.Len(t, result, tt.expected)
		})
	}
}

func TestParseVariablesMap(t *testing.T) {
	tests := []struct {
		name      string
		variables map[string]any
		keys      []string
		wantFiles int
		wantKeys  []string
	}{
		{
			name: "single file upload",
			variables: map[string]any{
				"file": graphql.Upload{
					Filename:    "test.txt",
					Size:        100,
					ContentType: "text/plain",
					File:        strings.NewReader("test content"),
				},
			},
			keys:      nil,
			wantFiles: 1,
			wantKeys:  []string{"file"},
		},
		{
			name: "multiple file uploads",
			variables: map[string]any{
				"file1": graphql.Upload{
					Filename:    "test1.txt",
					Size:        100,
					ContentType: "text/plain",
					File:        strings.NewReader("test1"),
				},
				"file2": graphql.Upload{
					Filename:    "test2.txt",
					Size:        200,
					ContentType: "text/plain",
					File:        strings.NewReader("test2"),
				},
			},
			keys:      nil,
			wantFiles: 2,
			wantKeys:  []string{"file1", "file2"},
		},
		{
			name: "filtered keys",
			variables: map[string]any{
				"file1": graphql.Upload{
					Filename:    "test1.txt",
					Size:        100,
					ContentType: "text/plain",
					File:        strings.NewReader("test1"),
				},
				"file2": graphql.Upload{
					Filename:    "test2.txt",
					Size:        200,
					ContentType: "text/plain",
					File:        strings.NewReader("test2"),
				},
			},
			keys:      []string{"file1"},
			wantFiles: 1,
			wantKeys:  []string{"file1"},
		},
		{
			name: "array of uploads",
			variables: map[string]any{
				"files": []graphql.Upload{
					{
						Filename:    "test1.txt",
						Size:        100,
						ContentType: "text/plain",
						File:        strings.NewReader("test1"),
					},
					{
						Filename:    "test2.txt",
						Size:        200,
						ContentType: "text/plain",
						File:        strings.NewReader("test2"),
					},
				},
			},
			keys:      nil,
			wantFiles: 1,
			wantKeys:  []string{"files"},
		},
		{
			name:      "empty variables",
			variables: map[string]any{},
			keys:      nil,
			wantFiles: 0,
			wantKeys:  []string{},
		},
		{
			name: "no matching keys",
			variables: map[string]any{
				"file": graphql.Upload{
					Filename: "test.txt",
					Size:     100,
					File:     strings.NewReader("test"),
				},
			},
			keys:      []string{"otherKey"},
			wantFiles: 0,
			wantKeys:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseVariablesMap(tt.variables, tt.keys...)
			require.NoError(t, err)
			assert.Len(t, result, tt.wantFiles)

			for _, key := range tt.wantKeys {
				assert.Contains(t, result, key)
			}
		})
	}
}

func createMultipartForm(files map[string][]string) (*multipart.Form, error) {
	form := &multipart.Form{
		File: make(map[string][]*multipart.FileHeader),
	}

	for key, fileNames := range files {
		for _, fileName := range fileNames {
			content := "test content for " + fileName

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", `form-data; name="`+key+`"; filename="`+fileName+`"`)
			h.Set("Content-Type", "text/plain")

			part, err := writer.CreatePart(h)
			if err != nil {
				return nil, err
			}

			if _, err := part.Write([]byte(content)); err != nil {
				return nil, err
			}

			writer.Close()

			reader := multipart.NewReader(body, writer.Boundary())
			multipartForm, err := reader.ReadForm(1024 * 1024)
			if err != nil {
				return nil, err
			}

			form.File[key] = append(form.File[key], multipartForm.File[key]...)
		}
	}

	return form, nil
}

func TestParseMultipartForm(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string][]string
		keys      []string
		wantFiles int
	}{
		{
			name: "single file",
			files: map[string][]string{
				"file": {"test.txt"},
			},
			keys:      nil,
			wantFiles: 1,
		},
		{
			name: "multiple files same key",
			files: map[string][]string{
				"files": {"test1.txt", "test2.txt"},
			},
			keys:      nil,
			wantFiles: 1,
		},
		{
			name: "multiple keys",
			files: map[string][]string{
				"file1": {"test1.txt"},
				"file2": {"test2.txt"},
			},
			keys:      nil,
			wantFiles: 2,
		},
		{
			name: "filtered keys",
			files: map[string][]string{
				"file1": {"test1.txt"},
				"file2": {"test2.txt"},
			},
			keys:      []string{"file1"},
			wantFiles: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form, err := createMultipartForm(tt.files)
			require.NoError(t, err)

			result, err := parseMultipartForm(form, tt.keys...)
			require.NoError(t, err)
			assert.Len(t, result, tt.wantFiles)
		})
	}
}

func TestParseFilesFromSource(t *testing.T) {
	t.Run("map source", func(t *testing.T) {
		source := map[string]any{
			"file": graphql.Upload{
				Filename:    "test.txt",
				Size:        100,
				ContentType: "text/plain",
				File:        strings.NewReader("test"),
			},
		}

		result, err := ParseFilesFromSource(source)
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("http request with nil multipart form", func(t *testing.T) {
		req := &http.Request{}

		result, err := ParseFilesFromSource(req)
		require.NoError(t, err)
		assert.Len(t, result, 0)
	})

	t.Run("http request with multipart form", func(t *testing.T) {
		form, err := createMultipartForm(map[string][]string{
			"file": {"test.txt"},
		})
		require.NoError(t, err)

		req := &http.Request{
			MultipartForm: form,
		}

		result, err := ParseFilesFromSource(req)
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("multipart form source", func(t *testing.T) {
		form, err := createMultipartForm(map[string][]string{
			"file": {"test.txt"},
		})
		require.NoError(t, err)

		result, err := ParseFilesFromSource(form)
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})
}

func TestWriteFilesToContext(t *testing.T) {
	t.Run("new context", func(t *testing.T) {
		ctx := context.Background()
		files := Files{
			"file1": []File{
				{ID: "1", FieldName: "file1"},
			},
		}

		newCtx := WriteFilesToContext(ctx, files)

		result, err := FilesFromContext(newCtx)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Len(t, result["file1"], 1)
	})

	t.Run("append to existing context", func(t *testing.T) {
		ctx := context.Background()
		files1 := Files{
			"file1": []File{
				{ID: "1", FieldName: "file1"},
			},
		}

		ctx = WriteFilesToContext(ctx, files1)

		files2 := Files{
			"file1": []File{
				{ID: "2", FieldName: "file1"},
			},
		}

		ctx = WriteFilesToContext(ctx, files2)

		result, err := FilesFromContext(ctx)
		require.NoError(t, err)
		assert.Len(t, result["file1"], 2)
	})

	t.Run("multiple keys", func(t *testing.T) {
		ctx := context.Background()
		files := Files{
			"file1": []File{
				{ID: "1", FieldName: "file1"},
			},
			"file2": []File{
				{ID: "2", FieldName: "file2"},
			},
		}

		ctx = WriteFilesToContext(ctx, files)

		result, err := FilesFromContext(ctx)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})
}

func TestUpdateFileInContextByKey(t *testing.T) {
	t.Run("update existing file", func(t *testing.T) {
		ctx := context.Background()
		files := Files{
			"file1": []File{
				{ID: "1", FieldName: "file1", OriginalName: "old.txt"},
			},
		}

		ctx = WriteFilesToContext(ctx, files)

		updatedFile := File{
			ID:           "1",
			FieldName:    "file1",
			OriginalName: "new.txt",
		}

		ctx = UpdateFileInContextByKey(ctx, "file1", updatedFile)

		result, err := FilesFromContextWithKey(ctx, "file1")
		require.NoError(t, err)
		assert.Equal(t, "new.txt", result[0].OriginalName)
	})

	t.Run("file not found", func(t *testing.T) {
		ctx := context.Background()
		files := Files{
			"file1": []File{
				{ID: "1", FieldName: "file1"},
			},
		}

		ctx = WriteFilesToContext(ctx, files)

		updatedFile := File{
			ID:        "999",
			FieldName: "file1",
		}

		ctx = UpdateFileInContextByKey(ctx, "file1", updatedFile)

		result, err := FilesFromContextWithKey(ctx, "file1")
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "1", result[0].ID)
	})

	t.Run("empty context", func(t *testing.T) {
		ctx := context.Background()

		updatedFile := File{
			ID:        "1",
			FieldName: "file1",
		}

		ctx = UpdateFileInContextByKey(ctx, "file1", updatedFile)

		result, err := FilesFromContextWithKey(ctx, "file1")
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestRemoveFileFromContext(t *testing.T) {
	t.Run("remove single file", func(t *testing.T) {
		ctx := context.Background()
		files := Files{
			"file1": []File{
				{ID: "1", FieldName: "file1"},
			},
		}

		ctx = WriteFilesToContext(ctx, files)

		fileToRemove := File{ID: "1"}
		ctx = RemoveFileFromContext(ctx, fileToRemove)

		result, err := FilesFromContext(ctx)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("remove one of multiple files", func(t *testing.T) {
		ctx := context.Background()
		files := Files{
			"file1": []File{
				{ID: "1", FieldName: "file1"},
				{ID: "2", FieldName: "file1"},
			},
		}

		ctx = WriteFilesToContext(ctx, files)

		fileToRemove := File{ID: "1"}
		ctx = RemoveFileFromContext(ctx, fileToRemove)

		result, err := FilesFromContextWithKey(ctx, "file1")
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "2", result[0].ID)
	})

	t.Run("remove non-existent file", func(t *testing.T) {
		ctx := context.Background()
		files := Files{
			"file1": []File{
				{ID: "1", FieldName: "file1"},
			},
		}

		ctx = WriteFilesToContext(ctx, files)

		fileToRemove := File{ID: "999"}
		ctx = RemoveFileFromContext(ctx, fileToRemove)

		result, err := FilesFromContext(ctx)
		require.NoError(t, err)
		assert.Len(t, result["file1"], 1)
	})

	t.Run("empty context", func(t *testing.T) {
		ctx := context.Background()

		fileToRemove := File{ID: "1"}
		ctx = RemoveFileFromContext(ctx, fileToRemove)

		result, err := FilesFromContext(ctx)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestFilesFromContext(t *testing.T) {
	t.Run("retrieve files", func(t *testing.T) {
		ctx := context.Background()
		files := Files{
			"file1": []File{
				{ID: "1", FieldName: "file1"},
			},
		}

		ctx = WriteFilesToContext(ctx, files)

		result, err := FilesFromContext(ctx)
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("empty context", func(t *testing.T) {
		ctx := context.Background()

		_, err := FilesFromContext(ctx)
		assert.ErrorIs(t, err, storage.ErrNoFilesUploaded)
	})
}

func TestFilesFromContextWithKey(t *testing.T) {
	t.Run("retrieve files by key", func(t *testing.T) {
		ctx := context.Background()
		files := Files{
			"file1": []File{
				{ID: "1", FieldName: "file1"},
			},
			"file2": []File{
				{ID: "2", FieldName: "file2"},
			},
		}

		ctx = WriteFilesToContext(ctx, files)

		result, err := FilesFromContextWithKey(ctx, "file1")
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "1", result[0].ID)
	})

	t.Run("key not found", func(t *testing.T) {
		ctx := context.Background()
		files := Files{
			"file1": []File{
				{ID: "1", FieldName: "file1"},
			},
		}

		ctx = WriteFilesToContext(ctx, files)

		result, err := FilesFromContextWithKey(ctx, "nonexistent")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("empty context", func(t *testing.T) {
		ctx := context.Background()

		_, err := FilesFromContextWithKey(ctx, "file1")
		assert.ErrorIs(t, err, storage.ErrNoFilesUploaded)
	})
}

func TestGetFileIDsFromContext(t *testing.T) {
	t.Run("retrieve file IDs", func(t *testing.T) {
		ctx := context.Background()
		files := Files{
			"file1": []File{
				{ID: "1", FieldName: "file1"},
				{ID: "2", FieldName: "file1"},
			},
			"file2": []File{
				{ID: "3", FieldName: "file2"},
			},
		}

		ctx = WriteFilesToContext(ctx, files)

		result := GetFileIDsFromContext(ctx)
		assert.Len(t, result, 3)
		assert.Contains(t, result, "1")
		assert.Contains(t, result, "2")
		assert.Contains(t, result, "3")
	})

	t.Run("empty context", func(t *testing.T) {
		ctx := context.Background()

		result := GetFileIDsFromContext(ctx)
		assert.Empty(t, result)
	})
}

func TestProcessFilesForMutation(t *testing.T) {
	t.Run("process files with mutation", func(t *testing.T) {
		ctx := context.Background()
		files := Files{
			"file1": []File{
				{ID: "1", FieldName: "file1"},
			},
		}

		ctx = WriteFilesToContext(ctx, files)

		mockMutation := mocks.NewMockMutation(t)
		mockMutation.EXPECT().Type().Return("TestType")
		mockMutation.EXPECT().ID().Return("mutation-id-123", nil)

		newCtx, err := ProcessFilesForMutation(ctx, mockMutation, "file1")
		require.NoError(t, err)

		result, err := FilesFromContextWithKey(newCtx, "file1")
		require.NoError(t, err)
		assert.Equal(t, "mutation-id-123", result[0].Parent.ID)
		assert.Equal(t, "TestType", result[0].Parent.Type)
	})

	t.Run("process files with parent type override", func(t *testing.T) {
		ctx := context.Background()
		files := Files{
			"file1": []File{
				{ID: "1", FieldName: "file1"},
			},
		}

		ctx = WriteFilesToContext(ctx, files)

		mockMutation := mocks.NewMockMutation(t)
		mockMutation.EXPECT().Type().Return("TestType")
		mockMutation.EXPECT().ID().Return("mutation-id-456", nil)

		newCtx, err := ProcessFilesForMutation(ctx, mockMutation, "file1", "OverrideType")
		require.NoError(t, err)

		result, err := FilesFromContextWithKey(newCtx, "file1")
		require.NoError(t, err)
		assert.Equal(t, "mutation-id-456", result[0].Parent.ID)
		assert.Equal(t, "OverrideType", result[0].Parent.Type)
	})

	t.Run("no files in context", func(t *testing.T) {
		ctx := context.Background()

		mockMutation := mocks.NewMockMutation(t)

		newCtx, err := ProcessFilesForMutation(ctx, mockMutation, "file1")
		require.NoError(t, err)
		assert.Equal(t, ctx, newCtx)
	})

	t.Run("different field name", func(t *testing.T) {
		ctx := context.Background()
		files := Files{
			"file1": []File{
				{ID: "1", FieldName: "file1"},
			},
		}

		ctx = WriteFilesToContext(ctx, files)

		mockMutation := mocks.NewMockMutation(t)

		newCtx, err := ProcessFilesForMutation(ctx, mockMutation, "file2")
		require.NoError(t, err)

		result, err := FilesFromContextWithKey(newCtx, "file1")
		require.NoError(t, err)
		assert.Empty(t, result[0].Parent.ID)
	})
}

func TestReaderToSeeker(t *testing.T) {
	t.Run("nil reader", func(t *testing.T) {
		result, err := ReaderToSeeker(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("already a ReadSeeker", func(t *testing.T) {
		data := []byte("test data")
		reader := bytes.NewReader(data)

		result, err := ReaderToSeeker(reader)
		require.NoError(t, err)
		assert.Equal(t, reader, result)
	})

	t.Run("small reader uses buffering", func(t *testing.T) {
		data := []byte("small test data")
		reader := io.NopCloser(bytes.NewReader(data))

		result, err := ReaderToSeeker(reader)
		require.NoError(t, err)
		assert.NotNil(t, result)

		buf := make([]byte, len(data))
		n, err := result.Read(buf)
		require.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.Equal(t, data, buf)
	})

	t.Run("buffered reader is returned as is", func(t *testing.T) {
		data := []byte("test data")
		br := NewBufferedReader(data)

		result, err := ReaderToSeeker(br)
		require.NoError(t, err)
		assert.Equal(t, br, result)
	})
}
