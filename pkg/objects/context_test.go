package objects

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteFilesToContext(t *testing.T) {
	tests := []struct {
		name           string
		initialFiles   Files
		newFiles       Files
		expectedResult Files
	}{
		{
			name:         "Add new files to empty context",
			initialFiles: Files{},
			newFiles: Files{
				"field1": {
					{ID: "1", FieldName: "field1"},
				},
			},
			expectedResult: Files{
				"field1": {
					{ID: "1", FieldName: "field1"},
				},
			},
		},
		{
			name: "Add new files to existing context",
			initialFiles: Files{
				"field1": {
					{ID: "1", FieldName: "field1"},
				},
			},
			newFiles: Files{
				"field1": {
					{ID: "2", FieldName: "field1"},
				},
			},
			expectedResult: Files{
				"field1": {
					{ID: "1", FieldName: "field1"},
					{ID: "2", FieldName: "field1"},
				},
			},
		},
		{
			name: "Add files to different fields",
			initialFiles: Files{
				"field1": {
					{ID: "1", FieldName: "field1"},
				},
			},
			newFiles: Files{
				"field2": {
					{ID: "2", FieldName: "field2"},
				},
			},
			expectedResult: Files{
				"field1": {
					{ID: "1", FieldName: "field1"},
				},
				"field2": {
					{ID: "2", FieldName: "field2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), FileContextKey, tt.initialFiles)
			ctx = WriteFilesToContext(ctx, tt.newFiles)

			result, ok := ctx.Value(FileContextKey).(Files)
			require.True(t, ok)

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
func TestFilesFromContext(t *testing.T) {
	tests := []struct {
		name           string
		initialFiles   Files
		expectedResult Files
	}{
		{
			name:           "Files exist in context",
			initialFiles:   Files{"field1": {{ID: "1", FieldName: "field1"}}},
			expectedResult: Files{"field1": {{ID: "1", FieldName: "field1"}}},
		},
		{
			name:           "No files in context",
			initialFiles:   nil,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), FileContextKey, tt.initialFiles)

			result, err := FilesFromContext(ctx)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestFilesFromContextWithKey(t *testing.T) {
	tests := []struct {
		name           string
		initialFiles   Files
		ctxKey         *ContextKey
		key            string
		expectedResult []File
		expectedError  error
	}{
		{
			name: "Files exist for the given key",
			initialFiles: Files{
				"field1": {
					{ID: "1", FieldName: "field1"},
					{ID: "2", FieldName: "field1"},
				},
			},
			ctxKey: FileContextKey,
			key:    "field1",
			expectedResult: []File{
				{ID: "1", FieldName: "field1"},
				{ID: "2", FieldName: "field1"},
			},
			expectedError: nil,
		},
		{
			name: "No files for the given key",
			initialFiles: Files{
				"field1": {
					{ID: "1", FieldName: "field1"},
				},
			},
			ctxKey:         FileContextKey,
			key:            "field2",
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:           "No files in context",
			initialFiles:   nil,
			ctxKey:         FileContextKey,
			key:            "field1",
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:           "No files in context, invalid key",
			initialFiles:   nil,
			ctxKey:         nil,
			key:            "field1",
			expectedResult: nil,
			expectedError:  ErrNoFilesUploaded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), tt.ctxKey, tt.initialFiles)

			result, err := FilesFromContextWithKey(ctx, tt.key)
			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
func TestGetFileIDsFromContext(t *testing.T) {
	tests := []struct {
		name           string
		initialFiles   Files
		expectedResult []string
	}{
		{
			name: "Files exist in context",
			initialFiles: Files{
				"field1": {
					{ID: "1", FieldName: "field1"},
					{ID: "2", FieldName: "field1"},
				},
				"field2": {
					{ID: "3", FieldName: "field2"},
				},
			},
			expectedResult: []string{"1", "2", "3"},
		},
		{
			name:           "No files in context",
			initialFiles:   nil,
			expectedResult: []string{},
		},
		{
			name: "Empty files in context",
			initialFiles: Files{
				"field1": {},
			},
			expectedResult: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), FileContextKey, tt.initialFiles)

			result := GetFileIDsFromContext(ctx)

			require.Len(t, result, len(tt.expectedResult))

			if len(tt.expectedResult) > 0 {
				assert.EqualValues(t, tt.expectedResult, result)
			}
		})
	}
}
