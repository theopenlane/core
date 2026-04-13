package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitFullName_FirstAndLast(t *testing.T) {
	first, last := splitFullName("Ada Lovelace")
	assert.Equal(t, "Ada", first)
	assert.Equal(t, "Lovelace", last)
}

func TestSplitFullName_FirstOnly(t *testing.T) {
	first, last := splitFullName("Ada")
	assert.Equal(t, "Ada", first)
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

func TestSplitFullName_MultipleNames(t *testing.T) {
	first, last := splitFullName("Ada Byron Lovelace")
	assert.Equal(t, "Ada", first)
	assert.Equal(t, "Byron Lovelace", last)
}
