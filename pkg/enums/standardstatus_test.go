package enums

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToStandardStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected *StandardStatus
	}{
		{"ACTIVE", &StandardActive},
		{"DRAFT", &StandardDraft},
		{"ARCHIVED", &StandardArchived},
		{"active", &StandardActive},
		{"draft", &StandardDraft},
		{"archived", &StandardArchived},
		{"invalid", nil},
		{"", nil},
	}

	for _, test := range tests {
		result := ToStandardStatus(test.input)
		if test.expected == nil {
			assert.Nil(t, result, "ToStandardStatus(%q) should be nil", test.input)
		} else {
			assert.NotNil(t, result, "ToStandardStatus(%q) should not be nil", test.input)
			assert.Equal(t, *test.expected, *result, "ToStandardStatus(%q) should be %q", test.input, *test.expected)
		}
	}
}
