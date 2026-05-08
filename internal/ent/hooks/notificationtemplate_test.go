package hooks

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestExtractTemplateVarNames(t *testing.T) {
	tests := []struct {
		name      string
		templates []string
		expected  map[string]string
	}{
		{
			name:      "simple variable",
			templates: []string{"Hello {{ .Name }}"},
			expected:  map[string]string{"Name": "string"},
		},
		{
			name:      "dotted path yields object",
			templates: []string{"{{ .User.Name }}"},
			expected:  map[string]string{"User": "object"},
		},
		{
			name:      "mixed simple and dotted",
			templates: []string{"{{ .Title }} by {{ .Author.FirstName }}"},
			expected:  map[string]string{"Title": "string", "Author": "object"},
		},
		{
			name:      "dotted path promotes existing string to object",
			templates: []string{"{{ .User }}", "{{ .User.Email }}"},
			expected:  map[string]string{"User": "object"},
		},
		{
			name:      "multiple templates merged",
			templates: []string{"{{ .Foo }}", "{{ .Bar }}"},
			expected:  map[string]string{"Foo": "string", "Bar": "string"},
		},
		{
			name:      "whitespace variations",
			templates: []string{"{{.Tight}}", "{{  .Loose  }}"},
			expected:  map[string]string{"Tight": "string", "Loose": "string"},
		},
		{
			name:      "deeply nested path uses top-level key",
			templates: []string{"{{ .Config.Database.Host }}"},
			expected:  map[string]string{"Config": "object"},
		},
		{
			name:      "duplicate variables deduplicated",
			templates: []string{"{{ .Name }} and {{ .Name }}"},
			expected:  map[string]string{"Name": "string"},
		},
		{
			name:      "no variables",
			templates: []string{"plain text with no templates"},
			expected:  map[string]string{},
		},
		{
			name:      "empty input",
			templates: nil,
			expected:  map[string]string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := extractTemplateVarNames(tc.templates...)
			assert.Assert(t, is.DeepEqual(tc.expected, result))
		})
	}
}
