package awssecurityhub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestParseDuration_Valid verifies valid durations are parsed.
func TestParseDuration_Valid(t *testing.T) {
	duration := parseDuration("1h30m")
	assert.Equal(t, 90*time.Minute, duration)
}

// TestParseDuration_Empty verifies empty durations return zero.
func TestParseDuration_Empty(t *testing.T) {
	duration := parseDuration("")
	assert.Equal(t, time.Duration(0), duration)
}

// TestParseDuration_Invalid verifies invalid durations return zero.
func TestParseDuration_Invalid(t *testing.T) {
	duration := parseDuration("notaduration")
	assert.Equal(t, time.Duration(0), duration)
}
