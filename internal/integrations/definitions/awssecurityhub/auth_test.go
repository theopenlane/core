package awssecurityhub

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// TestParseDuration_Valid verifies valid durations are parsed.
func TestParseDuration_Valid(t *testing.T) {
	duration := parseDuration("1h30m")
	assert.Check(t, is.Equal(90*time.Minute, duration))
}

// TestParseDuration_Empty verifies empty durations return zero.
func TestParseDuration_Empty(t *testing.T) {
	duration := parseDuration("")
	assert.Check(t, is.Equal(time.Duration(0), duration))
}

// TestParseDuration_Invalid verifies invalid durations return zero.
func TestParseDuration_Invalid(t *testing.T) {
	duration := parseDuration("notaduration")
	assert.Check(t, is.Equal(time.Duration(0), duration))
}
