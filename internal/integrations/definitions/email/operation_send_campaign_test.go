package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitFullName_FirstAndLast(t *testing.T) {
	first, last := splitFullName("Alice Smith")

	assert.Equal(t, "Alice", first)
	assert.Equal(t, "Smith", last)
}

func TestSplitFullName_FirstOnly(t *testing.T) {
	first, last := splitFullName("Alice")

	assert.Equal(t, "Alice", first)
	assert.Equal(t, "", last)
}

func TestSplitFullName_Empty(t *testing.T) {
	first, last := splitFullName("")

	assert.Equal(t, "", first)
	assert.Equal(t, "", last)
}

func TestSplitFullName_Whitespace(t *testing.T) {
	first, last := splitFullName("   ")

	assert.Equal(t, "", first)
	assert.Equal(t, "", last)
}

func TestSplitFullName_LeadingTrailingSpaces(t *testing.T) {
	first, last := splitFullName("  Alice Smith  ")

	assert.Equal(t, "Alice", first)
	assert.Equal(t, "Smith", last)
}

func TestSplitFullName_MultipleNames(t *testing.T) {
	// Only splits on the first space; "Mary Smith" stays in last
	first, last := splitFullName("Jane Mary Smith")

	assert.Equal(t, "Jane", first)
	assert.Equal(t, "Mary Smith", last)
}

func TestSplitFullName_Unicode(t *testing.T) {
	first, last := splitFullName("José García")

	assert.Equal(t, "José", first)
	assert.Equal(t, "García", last)
}
