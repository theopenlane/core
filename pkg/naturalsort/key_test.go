package naturalsort

import (
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyOrder(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name:     "nist families and enhancements",
			expected: []string{"AC-1", "AC-2", "AC-2(1)", "AC-2(2)", "AC-2(10)", "AC-3", "AC-9", "AC-10", "AC-11", "AC-20", "AU-1"},
		},
		{
			name:     "soc 2 criteria",
			expected: []string{"CC1.1", "CC1.2", "CC2.1", "CC9.9", "CC10.1"},
		},
		{
			name:     "case insensitive letters",
			expected: []string{"A1", "a2", "A10"},
		},
		{
			name:     "growing digit runs",
			expected: []string{"A9", "A99", "A100", "A999", "A1000"},
		},
		{
			name:     "ascii before non-ascii",
			expected: []string{"A1", "Z1", "Ä1"},
		},
		{
			name:     "prefix before extensions",
			expected: []string{"", "A", "A-1", "A-1.1", "A-1.1.1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			shuffled := slices.Clone(tc.expected)
			rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

			slices.SortStableFunc(shuffled, func(a, b string) int { return strings.Compare(Key(a), Key(b)) })
			assert.Equal(t, tc.expected, shuffled)
		})
	}
}

func TestKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "", expected: ""},
		{input: "AC-2(10)", expected: "ac-012(0210)"},
		{input: "AC-2", expected: "ac-012"},
		{input: "AC-02", expected: "ac-012"},
		{input: "a01", expected: "a011"},
		{input: "A1", expected: "a011"},
		{input: "AC-0", expected: "ac-010"},
		{input: "Ä1", expected: "Ä011"},
		{input: "A" + strings.Repeat("1", 120), expected: "a99" + strings.Repeat("1", 120)},
	}

	for _, tc := range tests {
		t.Run(tc.input[:min(len(tc.input), 16)], func(t *testing.T) {
			assert.Equal(t, tc.expected, Key(tc.input))
		})
	}
}
