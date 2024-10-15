package objects

import (
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFiles(t *testing.T) {
	// Prepare test data
	fileHeader1 := &multipart.FileHeader{Filename: "file1.txt"}
	fileHeader2 := &multipart.FileHeader{Filename: "file2.txt"}
	files := map[string][]*multipart.FileHeader{
		"field1": {fileHeader1},
		"field2": {fileHeader2},
	}

	tests := []struct {
		name     string
		files    map[string][]*multipart.FileHeader
		names    []string
		expected []*multipart.FileHeader
	}{
		{
			name:     "Single name match",
			files:    files,
			names:    []string{"field1"},
			expected: []*multipart.FileHeader{fileHeader1},
		},
		{
			name:     "Multiple names, first match",
			files:    files,
			names:    []string{"field1", "field2"},
			expected: []*multipart.FileHeader{fileHeader1},
		},
		{
			name:     "Multiple names, second match",
			files:    files,
			names:    []string{"field3", "field2"},
			expected: []*multipart.FileHeader{fileHeader2},
		},
		{
			name:     "No match",
			files:    files,
			names:    []string{"field3"},
			expected: nil,
		},
		{
			name:     "Empty names",
			files:    files,
			names:    []string{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFiles(tt.files, tt.names...)
			assert.Equal(t, tt.expected, result)
		})
	}
}
