package utils_test

import (
	"testing"

	"github.com/theopenlane/ent/privacy/utils"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestSliceToMap(t *testing.T) {
	type testCase struct {
		name     string
		input    []string
		expected map[string]any
	}
	tests := []testCase{
		{
			name:     "empty slice",
			input:    []string{},
			expected: map[string]any{},
		},
		{
			name:     "single element",
			input:    []string{"foo"},
			expected: map[string]any{"foo": struct{}{}},
		},
		{
			name:     "multiple elements",
			input:    []string{"foo", "bar", "baz"},
			expected: map[string]any{"foo": struct{}{}, "bar": struct{}{}, "baz": struct{}{}},
		},
		{
			name:     "duplicate elements",
			input:    []string{"foo", "bar", "foo"},
			expected: map[string]any{"foo": struct{}{}, "bar": struct{}{}},
		},
		{
			name:     "empty string element",
			input:    []string{""},
			expected: map[string]any{"": struct{}{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := utils.SliceToMap(tc.input)
			assert.Check(t, is.Len(got, len(tc.expected)))
			assert.Check(t, is.DeepEqual(got, tc.expected))
		})
	}
}
func TestCheckContains(t *testing.T) {
	type testCase struct {
		name     string
		s        []string
		e        []string
		expected bool
	}
	tests := []testCase{
		{
			name:     "empty s and e",
			s:        []string{},
			e:        []string{},
			expected: false,
		},
		{
			name:     "empty s, non-empty e",
			s:        []string{},
			e:        []string{"foo"},
			expected: false,
		},
		{
			name:     "non-empty s, empty e",
			s:        []string{"foo"},
			e:        []string{},
			expected: false,
		},
		{
			name:     "no intersection",
			s:        []string{"foo", "bar"},
			e:        []string{"baz", "qux"},
			expected: false,
		},
		{
			name:     "one element matches",
			s:        []string{"foo", "bar"},
			e:        []string{"baz", "foo"},
			expected: true,
		},
		{
			name:     "all elements match",
			s:        []string{"foo", "bar"},
			e:        []string{"foo", "bar"},
			expected: true,
		},
		{
			name:     "duplicate elements in e",
			s:        []string{"foo"},
			e:        []string{"foo", "foo"},
			expected: true,
		},
		{
			name:     "duplicate elements in s",
			s:        []string{"foo", "foo"},
			e:        []string{"foo"},
			expected: true,
		},
		{
			name:     "empty string in s and e",
			s:        []string{""},
			e:        []string{""},
			expected: true,
		},
		{
			name:     "empty string in e only",
			s:        []string{"foo"},
			e:        []string{""},
			expected: false,
		},
		{
			name:     "empty string in s only",
			s:        []string{""},
			e:        []string{"foo"},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := utils.CheckContains(tc.s, tc.e)
			assert.Check(t, is.Equal(got, tc.expected))
		})
	}
}
func TestGetIntersection(t *testing.T) {
	type testCase struct {
		name     string
		s1       []string
		s2       []string
		expected []string
	}
	tests := []testCase{
		{
			name:     "both empty",
			s1:       []string{},
			s2:       []string{},
			expected: []string{},
		},
		{
			name:     "s1 empty",
			s1:       []string{},
			s2:       []string{"foo", "bar"},
			expected: []string{},
		},
		{
			name:     "s2 empty",
			s1:       []string{"foo", "bar"},
			s2:       []string{},
			expected: []string{},
		},
		{
			name:     "no intersection",
			s1:       []string{"foo", "bar"},
			s2:       []string{"baz", "qux"},
			expected: []string{},
		},
		{
			name:     "one intersection",
			s1:       []string{"foo", "bar"},
			s2:       []string{"baz", "foo"},
			expected: []string{"foo"},
		},
		{
			name:     "multiple intersections",
			s1:       []string{"foo", "bar", "baz"},
			s2:       []string{"baz", "foo", "qux"},
			expected: []string{"baz", "foo"},
		},
		{
			name:     "duplicates in s2",
			s1:       []string{"foo", "bar"},
			s2:       []string{"foo", "foo", "bar"},
			expected: []string{"foo", "bar"},
		},
		{
			name:     "duplicates in s1",
			s1:       []string{"foo", "foo", "bar"},
			s2:       []string{"foo", "bar"},
			expected: []string{"foo", "bar"},
		},
		{
			name:     "empty string intersection",
			s1:       []string{""},
			s2:       []string{""},
			expected: []string{""},
		},
		{
			name:     "empty string in s2 only",
			s1:       []string{"foo"},
			s2:       []string{""},
			expected: []string{},
		},
		{
			name:     "empty string in s1 only",
			s1:       []string{""},
			s2:       []string{"foo"},
			expected: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := utils.GetIntersection(tc.s1, tc.s2)
			assert.Check(t, is.DeepEqual(got, tc.expected))
		})
	}
}
