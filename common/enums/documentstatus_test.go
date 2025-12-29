package enums

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, test.expected, result, "ToDocumentStatus(%s)", test.input)
	}
}
