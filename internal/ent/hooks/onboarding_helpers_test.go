package hooks

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestUniqueOrganizationName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Trim whitespace and replace spaces",
			input:    "  Acme Corp  ",
			expected: "acme-corp-",
		},
		{
			name:     "Remove non-alphanumeric characters",
			input:    "Acme Corp, Inc.",
			expected: "acme-corp-inc-",
		},
		{
			name:     "Lowercase conversion",
			input:    "ACME CORP",
			expected: "acme-corp-",
		},
		{
			name:     "No changes needed",
			input:    "acme-corp",
			expected: "acme-corp-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uniqueOrganizationName(tt.input)
			assert.Check(t, is.Contains(result, tt.expected))

			suffix := strings.Replace(result, tt.expected, "", 1)
			assert.Check(t, is.Len(suffix, 6))
		})
	}
}
