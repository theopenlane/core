package objects_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
)

func TestWriteFilesToContext(t *testing.T) {
	tests := []struct {
		name           string
		existingFiles  storage.Files
		newFiles       storage.Files
		expectedResult storage.Files
	}{
		{
			name:          "add files to empty context",
			existingFiles: nil,
			newFiles: storage.Files{
				"upload1": []storage.File{
					{ID: "file1", Name: "test1.txt", FieldName: "upload1"},
				},
			},
			expectedResult: storage.Files{
				"upload1": []storage.File{
					{ID: "file1", Name: "test1.txt", FieldName: "upload1"},
				},
			},
		},
		{
			name: "add files to existing context",
			existingFiles: storage.Files{
				"upload1": []storage.File{
					{ID: "file1", Name: "test1.txt", FieldName: "upload1"},
				},
			},
			newFiles: storage.Files{
				"upload2": []storage.File{
					{ID: "file2", Name: "test2.txt", FieldName: "upload2"},
				},
			},
			expectedResult: storage.Files{
				"upload1": []storage.File{
					{ID: "file1", Name: "test1.txt", FieldName: "upload1"},
				},
				"upload2": []storage.File{
					{ID: "file2", Name: "test2.txt", FieldName: "upload2"},
				},
			},
		},
		{
			name: "append to existing field",
			existingFiles: storage.Files{
				"upload1": []storage.File{
					{ID: "file1", Name: "test1.txt", FieldName: "upload1"},
				},
			},
			newFiles: storage.Files{
				"upload1": []storage.File{
					{ID: "file2", Name: "test2.txt", FieldName: "upload1"},
				},
			},
			expectedResult: storage.Files{
				"upload1": []storage.File{
					{ID: "file1", Name: "test1.txt", FieldName: "upload1"},
					{ID: "file2", Name: "test2.txt", FieldName: "upload1"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Add existing files to context if any
			if tt.existingFiles != nil {
				ctx = objects.WriteFilesToContext(ctx, tt.existingFiles)
			}

			// Add new files
			result := objects.WriteFilesToContext(ctx, tt.newFiles)

			// Verify result
			files, err := objects.FilesFromContext(result)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, files)
		})
	}
}

func TestUpdateFileInContextByKey(t *testing.T) {
	tests := []struct {
		name           string
		initialFiles   storage.Files
		key            string
		updatedFile    storage.File
		expectedResult storage.Files
	}{
		{
			name: "update existing file",
			initialFiles: storage.Files{
				"upload": []storage.File{
					{ID: "file1", Name: "original.txt", FieldName: "upload"},
					{ID: "file2", Name: "other.txt", FieldName: "upload"},
				},
			},
			key:         "upload",
			updatedFile: storage.File{ID: "file1", Name: "updated.txt", FieldName: "upload"},
			expectedResult: storage.Files{
				"upload": []storage.File{
					{ID: "file1", Name: "updated.txt", FieldName: "upload"},
					{ID: "file2", Name: "other.txt", FieldName: "upload"},
				},
			},
		},
		{
			name: "update non-existent file",
			initialFiles: storage.Files{
				"upload": []storage.File{
					{ID: "file1", Name: "original.txt", FieldName: "upload"},
				},
			},
			key:         "upload",
			updatedFile: storage.File{ID: "file999", Name: "nonexistent.txt", FieldName: "upload"},
			expectedResult: storage.Files{
				"upload": []storage.File{
					{ID: "file1", Name: "original.txt", FieldName: "upload"},
				},
			},
		},
		{
			name:           "update file in empty context",
			initialFiles:   nil,
			key:            "upload",
			updatedFile:    storage.File{ID: "file1", Name: "test.txt", FieldName: "upload"},
			expectedResult: storage.Files{}, // Results in empty Files{} since nothing was there to update
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Add initial files to context if any
			if tt.initialFiles != nil {
				ctx = objects.WriteFilesToContext(ctx, tt.initialFiles)
			}

			// Update file
			result := objects.UpdateFileInContextByKey(ctx, tt.key, tt.updatedFile)

			// Verify result
			files, err := objects.FilesFromContext(result)
			if tt.expectedResult == nil {
				assert.Error(t, err)
				assert.Equal(t, storage.ErrNoFilesUploaded, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, files)
			}
		})
	}
}

func TestRemoveFileFromContext(t *testing.T) {
	tests := []struct {
		name           string
		initialFiles   storage.Files
		fileToRemove   storage.File
		expectedResult storage.Files
	}{
		{
			name: "remove existing file",
			initialFiles: storage.Files{
				"upload": []storage.File{
					{ID: "file1", Name: "test1.txt", FieldName: "upload"},
					{ID: "file2", Name: "test2.txt", FieldName: "upload"},
				},
			},
			fileToRemove: storage.File{ID: "file1"},
			expectedResult: storage.Files{
				"upload": []storage.File{
					{ID: "file2", Name: "test2.txt", FieldName: "upload"},
				},
			},
		},
		{
			name: "remove last file in key",
			initialFiles: storage.Files{
				"upload": []storage.File{
					{ID: "file1", Name: "test1.txt", FieldName: "upload"},
				},
			},
			fileToRemove:   storage.File{ID: "file1"},
			expectedResult: storage.Files{},
		},
		{
			name: "remove non-existent file",
			initialFiles: storage.Files{
				"upload": []storage.File{
					{ID: "file1", Name: "test1.txt", FieldName: "upload"},
				},
			},
			fileToRemove: storage.File{ID: "file999"},
			expectedResult: storage.Files{
				"upload": []storage.File{
					{ID: "file1", Name: "test1.txt", FieldName: "upload"},
				},
			},
		},
		{
			name:           "remove file from empty context",
			initialFiles:   nil,
			fileToRemove:   storage.File{ID: "file1"},
			expectedResult: storage.Files{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Add initial files to context if any
			if tt.initialFiles != nil {
				ctx = objects.WriteFilesToContext(ctx, tt.initialFiles)
			}

			// Remove file
			result := objects.RemoveFileFromContext(ctx, tt.fileToRemove)

			// Verify result
			files, err := objects.FilesFromContext(result)
			if tt.expectedResult == nil {
				assert.Error(t, err)
				assert.Equal(t, storage.ErrNoFilesUploaded, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, files)
			}
		})
	}
}

func TestFilesFromContext(t *testing.T) {
	tests := []struct {
		name        string
		files       storage.Files
		expectError bool
	}{
		{
			name: "get files from context",
			files: storage.Files{
				"upload": []storage.File{
					{ID: "file1", Name: "test1.txt", FieldName: "upload"},
				},
			},
			expectError: false,
		},
		{
			name:        "get files from empty context",
			files:       nil,
			expectError: true,
		},
		{
			name:        "get files from context with empty files",
			files:       storage.Files{},
			expectError: false, // Empty Files{} is valid, just contains no files
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			if tt.files != nil {
				ctx = objects.WriteFilesToContext(ctx, tt.files)
			}

			files, err := objects.FilesFromContext(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, storage.ErrNoFilesUploaded, err)
				assert.Nil(t, files)
			} else {
				assert.NoError(t, err)
				if tt.files == nil {
					assert.Equal(t, storage.Files{}, files) // Empty files map
				} else {
					assert.Equal(t, tt.files, files)
				}
			}
		})
	}
}

func TestFilesFromContextWithKey(t *testing.T) {
	tests := []struct {
		name          string
		files         storage.Files
		key           string
		expectedFiles []storage.File
		expectError   bool
	}{
		{
			name: "get files with existing key",
			files: storage.Files{
				"upload1": []storage.File{
					{ID: "file1", Name: "test1.txt", FieldName: "upload1"},
				},
				"upload2": []storage.File{
					{ID: "file2", Name: "test2.txt", FieldName: "upload2"},
				},
			},
			key: "upload1",
			expectedFiles: []storage.File{
				{ID: "file1", Name: "test1.txt", FieldName: "upload1"},
			},
			expectError: false,
		},
		{
			name: "get files with non-existent key",
			files: storage.Files{
				"upload1": []storage.File{
					{ID: "file1", Name: "test1.txt", FieldName: "upload1"},
				},
			},
			key:           "upload2",
			expectedFiles: nil,
			expectError:   false,
		},
		{
			name:          "get files from empty context",
			files:         nil,
			key:           "upload",
			expectedFiles: nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			if tt.files != nil {
				ctx = objects.WriteFilesToContext(ctx, tt.files)
			}

			files, err := objects.FilesFromContextWithKey(ctx, tt.key)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, storage.ErrNoFilesUploaded, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFiles, files)
			}
		})
	}
}

func TestGetFileIDsFromContext(t *testing.T) {
	tests := []struct {
		name        string
		files       storage.Files
		expectedIDs []string
	}{
		{
			name: "get file IDs from context with multiple files",
			files: storage.Files{
				"upload1": []storage.File{
					{ID: "file1", Name: "test1.txt"},
					{ID: "file2", Name: "test2.txt"},
				},
				"upload2": []storage.File{
					{ID: "file3", Name: "test3.txt"},
				},
			},
			expectedIDs: []string{"file1", "file2", "file3"},
		},
		{
			name: "get file IDs from context with single file",
			files: storage.Files{
				"upload": []storage.File{
					{ID: "file1", Name: "test1.txt"},
				},
			},
			expectedIDs: []string{"file1"},
		},
		{
			name:        "get file IDs from empty context",
			files:       nil,
			expectedIDs: []string{},
		},
		{
			name:        "get file IDs from context with empty files",
			files:       storage.Files{},
			expectedIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			if tt.files != nil {
				ctx = objects.WriteFilesToContext(ctx, tt.files)
			}

			fileIDs := objects.GetFileIDsFromContext(ctx)

			// Since the order might vary due to map iteration, we'll check if all expected IDs are present
			assert.ElementsMatch(t, tt.expectedIDs, fileIDs)
		})
	}
}

func TestFileContextIntegration(t *testing.T) {
	ctx := context.Background()

	// Test the full workflow: write files -> update -> remove -> verify
	initialFiles := storage.Files{
		"upload": []storage.File{
			{ID: "file1", Name: "test1.txt", FieldName: "upload"},
			{ID: "file2", Name: "test2.txt", FieldName: "upload"},
		},
	}

	// Write initial files
	ctx = objects.WriteFilesToContext(ctx, initialFiles)

	// Verify initial state
	files, err := objects.FilesFromContext(ctx)
	require.NoError(t, err)
	assert.Len(t, files["upload"], 2)

	// Update one file
	updatedFile := storage.File{ID: "file1", Name: "updated.txt", FieldName: "upload"}
	ctx = objects.UpdateFileInContextByKey(ctx, "upload", updatedFile)

	// Verify update
	files, err = objects.FilesFromContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, "updated.txt", files["upload"][0].Name)

	// Remove one file
	ctx = objects.RemoveFileFromContext(ctx, storage.File{ID: "file2"})

	// Verify removal
	files, err = objects.FilesFromContext(ctx)
	require.NoError(t, err)
	assert.Len(t, files["upload"], 1)
	assert.Equal(t, "file1", files["upload"][0].ID)

	// Get file IDs
	fileIDs := objects.GetFileIDsFromContext(ctx)
	assert.Equal(t, []string{"file1"}, fileIDs)
}
