package graphapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	historygenerated "github.com/theopenlane/core/internal/ent/historygenerated"
)

func TestControlDiffValuesEqual(t *testing.T) {
	tests := []struct {
		name     string
		a, b     any
		expected bool
	}{
		{name: "both nil", a: nil, b: nil, expected: true},
		{name: "a nil", a: nil, b: "foo", expected: false},
		{name: "b nil", a: "foo", b: nil, expected: false},
		{name: "equal strings", a: "foo", b: "foo", expected: true},
		{name: "different strings", a: "foo", b: "bar", expected: false},
		{name: "equal slices", a: []string{"a", "b"}, b: []string{"a", "b"}, expected: true},
		{name: "different slices", a: []string{"a"}, b: []string{"a", "b"}, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, controlDiffValuesEqual(tt.a, tt.b))
		})
	}
}

func TestDiffControlHistoryRecords(t *testing.T) {
	oldRecord := &historygenerated.ControlHistory{
		Title:       "Old Title",
		Description: "Old Description",
		Category:    "Security",
		Subcategory: "Access Control",
	}

	newRecord := &historygenerated.ControlHistory{
		Title:       "New Title",
		Description: "Old Description",
		Category:    "Security",
		Subcategory: "Identity",
	}

	diffs, err := diffControlHistories(oldRecord, newRecord)
	require.NoError(t, err)

	fieldNames := make([]string, len(diffs))
	for i, d := range diffs {
		fieldNames[i] = d.Field
	}

	assert.Contains(t, fieldNames, "title")
	assert.Contains(t, fieldNames, "subcategory")
	assert.NotContains(t, fieldNames, "description")
	assert.NotContains(t, fieldNames, "category")

	for _, d := range diffs {
		if d.Field == "title" {
			assert.Equal(t, "Old Title", d.OldValue)
			assert.Equal(t, "New Title", d.NewValue)
		}
	}
}

func TestDiffControlHistoryRecordsNoDiffs(t *testing.T) {
	record := &historygenerated.ControlHistory{
		Title:       "Same",
		Description: "Same",
		Category:    "Same",
	}

	diffs, err := diffControlHistories(record, record)
	require.NoError(t, err)
	assert.Empty(t, diffs)
}
