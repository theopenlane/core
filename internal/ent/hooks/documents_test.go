package hooks

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestFilenameToTitle(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "with extension",
			filename: "test_file.txt",
			want:     "Test File",
		},
		{
			name:     "with hyphens",
			filename: "hello-world.pdf",
			want:     "Hello World",
		},
		{
			name:     "with underscores and hyphens",
			filename: "complex_file-name.docx",
			want:     "Complex File Name",
		},
		{
			name:     "no extension",
			filename: "simple_name",
			want:     "Simple Name",
		},
		{
			name:     "with extra spaces",
			filename: "  spaced_file.pdf  ",
			want:     "Spaced File",
		},
		{
			name:     "empty string",
			filename: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filenameToTitle(tt.filename)
			assert.Check(t, is.Equal(got, tt.want))
		})
	}
}
