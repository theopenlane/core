package runtime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsPastPaymentInterval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		createdAt    time.Time
		intervalDays uint8
		expected     bool
	}{
		{
			name:         "created well past interval",
			createdAt:    time.Now().Add(-40 * 24 * time.Hour),
			intervalDays: 30,
			expected:     true,
		},
		{
			name:         "created exactly at interval boundary",
			createdAt:    time.Now().Add(-30 * 24 * time.Hour),
			intervalDays: 30,
			expected:     true,
		},
		{
			name:         "created before interval",
			createdAt:    time.Now().Add(-5 * 24 * time.Hour),
			intervalDays: 30,
			expected:     false,
		},
		{
			name:         "zero interval always past",
			createdAt:    time.Now(),
			intervalDays: 0,
			expected:     true,
		},
		{
			name:         "created one day ago with one day interval",
			createdAt:    time.Now().Add(-25 * time.Hour),
			intervalDays: 1,
			expected:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, isPastPaymentInterval(tc.createdAt, tc.intervalDays))
		})
	}
}
