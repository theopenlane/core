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
		{name: "empty strings", a: "", b: "", expected: true},
		{name: "empty vs non-empty string", a: "", b: "something", expected: false},
		{name: "equal slices", a: []string{"a", "b"}, b: []string{"a", "b"}, expected: true},
		{name: "different slices", a: []string{"a"}, b: []string{"a", "b"}, expected: false},
		{name: "both empty slices", a: []string{}, b: []string{}, expected: true},
		{name: "equal integers", a: 42, b: 42, expected: true},
		{name: "different integers", a: 42, b: 99, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, controlDiffValuesEqual(tt.a, tt.b))
		})
	}
}

func TestDiffControlHistories(t *testing.T) {
	t.Run("detects title and subcategory changes", func(t *testing.T) {
		oldRecord := &historygenerated.ControlHistory{
			Title:       "Old Title",
			Description: "Same Description",
			Category:    "Security",
			Subcategory: "Access Control",
		}

		newRecord := &historygenerated.ControlHistory{
			Title:       "New Title",
			Description: "Same Description",
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
	})

	t.Run("no diffs when records are identical", func(t *testing.T) {
		record := &historygenerated.ControlHistory{
			Title:       "Same",
			Description: "Same",
			Category:    "Same",
		}

		diffs, err := diffControlHistories(record, record)
		require.NoError(t, err)
		assert.Empty(t, diffs)
	})

	t.Run("slice field changes detected", func(t *testing.T) {
		oldRecord := &historygenerated.ControlHistory{
			Aliases:          []string{"AC-1", "AC-2"},
			MappedCategories: []string{"NIST"},
			ControlQuestions: []string{"Q1"},
		}

		newRecord := &historygenerated.ControlHistory{
			Aliases:          []string{"AC-1", "AC-2", "AC-3"},
			MappedCategories: []string{"NIST"},
			ControlQuestions: []string{"Q1", "Q2"},
		}

		diffs, err := diffControlHistories(oldRecord, newRecord)
		require.NoError(t, err)

		fieldNames := make([]string, len(diffs))
		for i, d := range diffs {
			fieldNames[i] = d.Field
		}

		assert.Contains(t, fieldNames, "aliases")
		assert.Contains(t, fieldNames, "control_questions")
		assert.NotContains(t, fieldNames, "mapped_categories")
	})

	t.Run("detects change from empty to non-empty", func(t *testing.T) {
		oldRecord := &historygenerated.ControlHistory{
			Title:       "",
			Description: "",
		}

		newRecord := &historygenerated.ControlHistory{
			Title:       "There is now a title",
			Description: "",
		}

		diffs, err := diffControlHistories(oldRecord, newRecord)
		require.NoError(t, err)

		require.Len(t, diffs, 1)
		assert.Equal(t, "title", diffs[0].Field)
	})

	t.Run("zero value produces no diff", func(t *testing.T) {
		diffs, err := diffControlHistories(
			&historygenerated.ControlHistory{},
			&historygenerated.ControlHistory{},
		)
		require.NoError(t, err)
		assert.Empty(t, diffs)
	})
}

func TestDetectAndBuildControlChanges(t *testing.T) {
	t.Run("detected between old and new", func(t *testing.T) {
		oldByRef := map[string]*historygenerated.ControlHistory{
			"AC-1": {RefCode: "AC-1", Title: "Old Title", Description: "Same"},
			"AC-2": {RefCode: "AC-2", Title: "Unchanged", Description: "Same"},
		}

		newByRef := map[string]*historygenerated.ControlHistory{
			"AC-1": {RefCode: "AC-1", Title: "New Title", Description: "Same"},
			"AC-2": {RefCode: "AC-2", Title: "Unchanged", Description: "Same"},
		}

		changes, err := detectAndBuildControlChanges(oldByRef, newByRef)
		require.NoError(t, err)
		require.Len(t, changes, 1)
		assert.Equal(t, "AC-1", changes[0].RefCode)
		assert.Equal(t, "New Title", changes[0].Title)

		assert.Len(t, changes[0].Diffs, 1)
		assert.Equal(t, "title", changes[0].Diffs[0].Field)
	})

	t.Run("controls skipped in new revision", func(t *testing.T) {
		oldByRef := map[string]*historygenerated.ControlHistory{}

		newByRef := map[string]*historygenerated.ControlHistory{
			"AC-1": {RefCode: "AC-1", Title: "Brand New"},
		}

		changes, err := detectAndBuildControlChanges(oldByRef, newByRef)
		require.NoError(t, err)
		assert.Empty(t, changes)
	})

	t.Run("empty map returns empty slice", func(t *testing.T) {
		changes, err := detectAndBuildControlChanges(
			map[string]*historygenerated.ControlHistory{},
			map[string]*historygenerated.ControlHistory{},
		)
		require.NoError(t, err)
		require.NotNil(t, changes)
		assert.Empty(t, changes)
	})

	t.Run("multiple controls contain changes", func(t *testing.T) {
		oldRecordsByRef := map[string]*historygenerated.ControlHistory{
			"AC-1": {RefCode: "AC-1", Title: "Old1", Category: "Cat1"},
			"AC-2": {RefCode: "AC-2", Title: "Old2", Category: "Cat2"},
			"AC-3": {RefCode: "AC-3", Title: "Same", Category: "Same"},
		}

		newRecordsByRef := map[string]*historygenerated.ControlHistory{
			"AC-1": {RefCode: "AC-1", Title: "New1", Category: "Cat1"},
			"AC-2": {RefCode: "AC-2", Title: "Old2", Category: "NewCat2"},
			"AC-3": {RefCode: "AC-3", Title: "Same", Category: "Same"},
		}

		changes, err := detectAndBuildControlChanges(oldRecordsByRef, newRecordsByRef)
		require.NoError(t, err)
		assert.Len(t, changes, 2)

		changesByRef := make(map[string]bool)
		for _, c := range changes {
			changesByRef[c.RefCode] = true
		}

		assert.True(t, changesByRef["AC-1"])
		assert.True(t, changesByRef["AC-2"])
		assert.False(t, changesByRef["AC-3"])
	})
}
