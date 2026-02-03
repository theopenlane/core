package helpers

import (
	"strings"
	"time"
)

// ParseDuration returns a parsed duration or zero when invalid.
func ParseDuration(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0
	}
	return duration
}
