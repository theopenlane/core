package hooks

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
			assert.Contains(t, result, tt.expected)

			suffix := strings.Replace(result, tt.expected, "", 1)
			assert.Len(t, suffix, 6)
		})
	}
}
