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
func TestUpdatePlaceholderText(t *testing.T) {
	tests := []struct {
		name     string
		details  string
		orgName  string
		expected string
	}{
		{
			name:     "replace with organization name",
			details:  "Welcome to {{company_name}}, we're glad to have you!",
			orgName:  "Openlane",
			expected: "Welcome to Openlane, we're glad to have you!",
		},
		{
			name:     "multiple replacements",
			details:  "{{company_name}} is great! Visit {{company_name}} today.",
			orgName:  "Openlane",
			expected: "Openlane is great! Visit Openlane today.",
		},
		{
			name:     "empty organization name",
			details:  "Welcome to {{company_name}}!",
			orgName:  "",
			expected: "Welcome to [Company Name]!",
		},
		{
			name:     "no placeholders",
			details:  "Just a regular string with no placeholders.",
			orgName:  "Openlane",
			expected: "Just a regular string with no placeholders.",
		},
		{
			name:     "empty details",
			details:  "",
			orgName:  "Openlane",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updatePlaceholderText(tt.details, tt.orgName)
			assert.Check(t, is.Equal(result, tt.expected))
		})
	}
}
