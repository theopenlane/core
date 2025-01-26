package objects

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/utils/contextx"
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
			fCtxKey := FileContextKey{Files: tt.initialFiles}
			ctx := contextx.With(context.Background(), fCtxKey)

			ctx = WriteFilesToContext(ctx, tt.newFiles)

			result, ok := contextx.From[FileContextKey](ctx)
			require.True(t, ok)
			require.NotNil(t, result)

			assert.Equal(t, tt.expectedResult, result.Files)
		})
	}
}
func TestFilesFromContext(t *testing.T) {
	tests := []struct {
		name           string
		initialFiles   Files
		expectedResult Files
		errExpected    bool
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
			errExpected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fCtxKey := FileContextKey{Files: tt.initialFiles}
			ctx := contextx.With(context.Background(), fCtxKey)

			result, err := FilesFromContext(ctx)
			if tt.errExpected {
				require.Error(t, err)
				assert.Equal(t, ErrNoFilesUploaded, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestFilesFromContextWithKey(t *testing.T) {
	tests := []struct {
		name           string
		ctxKey         FileContextKey
		key            string
		expectedResult []File
		expectedError  error
	}{
		{
			name: "Files exist for the given key",
			ctxKey: FileContextKey{
				Files{
					"field1": {
						{ID: "1", FieldName: "field1"},
						{ID: "2", FieldName: "field1"},
					},
				}},
			key: "field1",
			expectedResult: []File{
				{ID: "1", FieldName: "field1"},
				{ID: "2", FieldName: "field1"},
			},
			expectedError: nil,
		},
		{
			name: "No files for the given key",
			ctxKey: FileContextKey{
				Files{
					"field1": {
						{ID: "1", FieldName: "field1"},
					},
				}},
			key:            "field2",
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:           "No files in context",
			ctxKey:         FileContextKey{Files: nil},
			key:            "field1",
			expectedResult: nil,
			expectedError:  ErrNoFilesUploaded,
		},
		{
			name:           "No files in context, invalid key",
			key:            "field1",
			expectedResult: nil,
			expectedError:  ErrNoFilesUploaded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := contextx.With(context.Background(), tt.ctxKey)

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
			fCtxKey := FileContextKey{Files: tt.initialFiles}
			ctx := contextx.With(context.Background(), fCtxKey)

			result := GetFileIDsFromContext(ctx)

			require.Len(t, result, len(tt.expectedResult))

			if len(tt.expectedResult) > 0 {
				assert.ElementsMatch(t, tt.expectedResult, result)
			}
		})
	}
}
