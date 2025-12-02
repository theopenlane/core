package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	tests := []struct {
		name            string
		requestsLimit   int64
		windowSize      time.Duration
		incNumber       int
		wantLimitStatus *LimitStatus
	}{
		{
			name:          "test_RateLimiter_IsLimited_not_limited",
			requestsLimit: int64(5),
			windowSize:    10 * time.Second,
			incNumber:     4,
			wantLimitStatus: &LimitStatus{
				IsLimited:   false,
				CurrentRate: 4,
			},
		},
		{
			name:          "test_RateLimiter_IsLimited_limited",
			requestsLimit: int64(5),
			windowSize:    10 * time.Second,
			incNumber:     6,
			wantLimitStatus: &LimitStatus{
				IsLimited:   true,
				CurrentRate: 6,
			},
		},
	}
	for _, tt := range tests {
		store := NewMapLimitStore(1*time.Hour, 1*time.Hour)
		r := New(store, tt.requestsLimit, tt.windowSize)

		for i := 0; i < tt.incNumber; i++ {
			err := r.Inc("key")
			assert.NoError(t, err, tt.name)
		}

		limitStatus, err := r.Check("key")
		assert.NoError(t, err, tt.name)
		assert.Equal(t, tt.wantLimitStatus.IsLimited, limitStatus.IsLimited, tt.name)
		assert.Equal(t, tt.wantLimitStatus.CurrentRate, limitStatus.CurrentRate, tt.name)
	}
}

func TestCalcDuration(t *testing.T) {
	tests := []struct {
		name               string
		prevValue          int64
		currValue          int64
		timeFromCurrWindow time.Duration
		requestsLimit      int64
		windowSize         time.Duration
		want               time.Duration
	}{
		{
			name:               "TestRateLimiter_calcLimitDuration_prev_value_is_not_zero",
			prevValue:          5,
			currValue:          6,
			timeFromCurrWindow: 1 * time.Second,
			requestsLimit:      5,
			windowSize:         10 * time.Second,
			want:               time.Duration(11 * time.Second), // 10*(1.0-( (5-6)/5)) - 1
		},
		{
			name:               "TestRateLimiter_calcLimitDuration_prev_value_is_zero",
			prevValue:          0,
			currValue:          6,
			timeFromCurrWindow: 1 * time.Second,
			requestsLimit:      5,
			windowSize:         10 * time.Second,
			want:               time.Duration(10666666666 * time.Nanosecond), // 10*(1.0-(5/6)) + (10-1)
		},
	}
	for _, tt := range tests {
		store := NewMapLimitStore(1*time.Hour, 1*time.Hour)
		r := New(store, tt.requestsLimit, tt.windowSize)
		dur := r.calcLimitDuration(tt.prevValue, tt.currValue, tt.timeFromCurrWindow)
		assert.InDelta(t, tt.want, dur, 3)
	}
}

func TestCalcRate(t *testing.T) {
	tests := []struct {
		name               string
		requestsLimit      int64
		windowSize         time.Duration
		timeFromCurrWindow time.Duration
		prevValue          int64
		currentValue       int64
		want               float64
	}{
		{
			name:               "TestRateLimiter_calcRate_prev_not_zero",
			requestsLimit:      5,
			windowSize:         10 * time.Second,
			timeFromCurrWindow: 1 * time.Second,
			prevValue:          5,
			currentValue:       6,
			want:               (0.9 * 5) + 6.0,
		},
		{
			name:               "TestRateLimiter_calcRate_prev_zero",
			requestsLimit:      5,
			windowSize:         10 * time.Second,
			timeFromCurrWindow: 1 * time.Second,
			prevValue:          0,
			currentValue:       6,
			want:               6.0,
		},
		{
			name:               "TestRateLimiter_calcRate_timeFromCurrWindow_zero",
			requestsLimit:      5,
			windowSize:         10 * time.Second,
			timeFromCurrWindow: 0 * time.Second,
			prevValue:          5,
			currentValue:       0,
			want:               5.0,
		},
		{
			name:               "TestRateLimiter_calcRate_timeFromCurrWindow_max",
			requestsLimit:      5,
			windowSize:         10 * time.Second,
			timeFromCurrWindow: 10 * time.Second,
			prevValue:          5,
			currentValue:       6,
			want:               6.0,
		},
	}
	for _, tt := range tests {
		store := NewMapLimitStore(1*time.Hour, 1*time.Hour)
		r := New(store, tt.requestsLimit, tt.windowSize)
		rate := r.calcRate(tt.timeFromCurrWindow, tt.prevValue, tt.currentValue)
		assert.Equal(t, tt.want, rate)
	}
}
