package enums

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestToDocumentStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected *DocumentStatus
	}{
		{"PUBLISHED", &DocumentPublished},
		{"DRAFT", &DocumentDraft},
		{"NEEDS_APPROVAL", &DocumentNeedsApproval},
		{"APPROVED", &DocumentApproved},
		{"ARCHIVED", &DocumentArchived},
		{"invalid", nil},
		{"", nil},
	}

	for _, test := range tests {
		result := ToDocumentStatus(test.input)
		assert.Check(t, is.DeepEqual(test.expected, result), "ToDocumentStatus(%s)", test.input)
	}
}
