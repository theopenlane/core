package graphapi

import (
	"slices"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	historygenerated "github.com/theopenlane/core/v2/internal/ent/historygenerated"
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
			assert.Check(t, is.Equal(tt.expected, controlDiffValuesEqual(tt.a, tt.b)))
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
		assert.NilError(t, err)

		fieldNames := make([]string, len(diffs))
		for i, d := range diffs {
			fieldNames[i] = d.Field
		}

		assert.Check(t, is.Contains(fieldNames, "title"))
		assert.Check(t, is.Contains(fieldNames, "subcategory"))
		assert.Check(t, !slices.Contains(fieldNames, "description"))
		assert.Check(t, !slices.Contains(fieldNames, "category"))
	})

	t.Run("no diffs when records are identical", func(t *testing.T) {
		record := &historygenerated.ControlHistory{
			Title:       "Same",
			Description: "Same",
			Category:    "Same",
		}

		diffs, err := diffControlHistories(record, record)
		assert.NilError(t, err)
		assert.Check(t, is.Len(diffs, 0))
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
		assert.NilError(t, err)

		fieldNames := make([]string, len(diffs))
		for i, d := range diffs {
			fieldNames[i] = d.Field
		}

		assert.Check(t, is.Contains(fieldNames, "aliases"))
		assert.Check(t, is.Contains(fieldNames, "control_questions"))
		assert.Check(t, !slices.Contains(fieldNames, "mapped_categories"))
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
		assert.NilError(t, err)

		assert.Assert(t, is.Len(diffs, 1))
		assert.Check(t, is.Equal("title", diffs[0].Field))
	})

	t.Run("zero value produces no diff", func(t *testing.T) {
		diffs, err := diffControlHistories(
			&historygenerated.ControlHistory{},
			&historygenerated.ControlHistory{},
		)
		assert.NilError(t, err)
		assert.Check(t, is.Len(diffs, 0))
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
		assert.NilError(t, err)
		assert.Assert(t, is.Len(changes, 1))
		assert.Check(t, is.Equal("AC-1", changes[0].RefCode))
		assert.Check(t, is.Equal("New Title", changes[0].Title))

		assert.Check(t, is.Len(changes[0].Diffs, 1))
		assert.Check(t, is.Equal("title", changes[0].Diffs[0].Field))
	})

	t.Run("controls skipped in new revision", func(t *testing.T) {
		oldByRef := map[string]*historygenerated.ControlHistory{}

		newByRef := map[string]*historygenerated.ControlHistory{
			"AC-1": {RefCode: "AC-1", Title: "Brand New"},
		}

		changes, err := detectAndBuildControlChanges(oldByRef, newByRef)
		assert.NilError(t, err)
		assert.Check(t, is.Len(changes, 0))
	})

	t.Run("empty map returns empty slice", func(t *testing.T) {
		changes, err := detectAndBuildControlChanges(
			map[string]*historygenerated.ControlHistory{},
			map[string]*historygenerated.ControlHistory{},
		)
		assert.NilError(t, err)
		assert.Assert(t, changes != nil)
		assert.Check(t, is.Len(changes, 0))
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
		assert.NilError(t, err)
		assert.Check(t, is.Len(changes, 2))

		changesByRef := make(map[string]bool)
		for _, c := range changes {
			changesByRef[c.RefCode] = true
		}

		assert.Check(t, changesByRef["AC-1"])
		assert.Check(t, changesByRef["AC-2"])
		assert.Check(t, !changesByRef["AC-3"])
	})
}
