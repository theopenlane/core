package workflows

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNormalizeStrings verifies normalization behavior for string slices
func TestNormalizeStrings(t *testing.T) {
	assert.Empty(t, NormalizeStrings([]string{"", ""}))

	result := NormalizeStrings([]string{"b", "", "a", "b"})
	assert.Equal(t, []string{"b", "a"}, result)
}
