package enums

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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
			assert.Check(t, is.Nil(result), "ToStandardStatus(%q) should be nil", test.input)
		} else {
			assert.Check(t, result != nil, "ToStandardStatus(%q) should not be nil", test.input)
			assert.Check(t, is.Equal(*test.expected, *result), "ToStandardStatus(%q) should be %q", test.input, *test.expected)
		}
	}
}
