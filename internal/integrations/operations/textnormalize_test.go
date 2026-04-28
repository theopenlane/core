package operations

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestNormalizeDescription(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "already has newlines, returned unchanged",
			input: "### Impact\n\nSQL Injection can occur when:\n\n1. The non-default simple protocol is used.",
			want:  "### Impact\n\nSQL Injection can occur when:\n\n1. The non-default simple protocol is used.",
		},
		{
			name:  "space-separated paragraphs normalized",
			input: "A Git repository can be crafted in a dangerous way.     Update Instructions:     Run `sudo pro fix CVE-2025-27614` to fix the vulnerability.",
			want:  "A Git repository can be crafted in a dangerous way.\n\nUpdate Instructions:\n\nRun `sudo pro fix CVE-2025-27614` to fix the vulnerability.",
		},
		{
			name:  "leading and trailing spaces trimmed",
			input: "   Some description.     More details.   ",
			want:  "Some description.\n\nMore details.",
		},
		{
			name:  "exactly two spaces left alone",
			input: "Sentence one.  Sentence two.",
			want:  "Sentence one.  Sentence two.",
		},
		{
			name:  "exactly three spaces normalized",
			input: "Section one.   Section two.",
			want:  "Section one.\n\nSection two.",
		},
		{
			name:  "empty string returned unchanged",
			input: "",
			want:  "",
		},
		{
			name:  "no spaces or newlines returned unchanged",
			input: "Plain description with no formatting.",
			want:  "Plain description with no formatting.",
		},
		{
			name:  "newline mid-string prevents normalization",
			input: "First line.\nSecond line.     Not normalized.",
			want:  "First line.\nSecond line.     Not normalized.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeDescription(tt.input)
			assert.Equal(t, got, tt.want)
		})
	}
}
