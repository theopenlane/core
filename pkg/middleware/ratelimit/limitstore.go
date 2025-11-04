package ratelimit

import (
	"time"
)

// inspired by Cloudflare's approach: https://blog.cloudflare.com/counting-things-a-lot-of-different-things/

// LimitStore is the interface that represents limiter internal data store
type LimitStore interface {
	// Inc increments current window limit counter for key
	Inc(key string, window time.Time) error
	// Get gets value of previous window counter and current window counter for key
	Get(key string, previousWindow, currentWindow time.Time) (prevValue int64, currValue int64, err error)
}

// RateLimiter is defining the structure of a rate limiter object
type RateLimiter struct {
	// dataStore is the internal limiter data store
	dataStore LimitStore
	// requestsLimit is the maximum number of requests allowed within a certain time window
	requestsLimit int64
	// windowSize is the duration of that time window
	windowSize time.Duration
}

// New creates new rate limiter
func New(dataStore LimitStore, requestsLimit int64, windowSize time.Duration) *RateLimiter {
	return &RateLimiter{
		dataStore:     dataStore,
		requestsLimit: requestsLimit,
		windowSize:    windowSize,
	}
}

// RequestsLimit returns the maximum number of requests allowed within the window.
func (r *RateLimiter) RequestsLimit() int64 {
	return r.requestsLimit
}

// WindowSize returns the duration of the configured window.
func (r *RateLimiter) WindowSize() time.Duration {
	return r.windowSize
}

// Inc increments the limit counter for a given key
func (r *RateLimiter) Inc(key string) error {
	currentWindow := time.Now().UTC().Truncate(r.windowSize)

	return r.dataStore.Inc(key, currentWindow)
}

// LimitStatus represents current status of limitation for a given key
type LimitStatus struct {
	// IsLimited is true when a given key should be rate-limited
	IsLimited bool
	// LimitDuration is the time for which a given key should be blocked
	LimitDuration *time.Duration
	// CurrentRate is approximated current requests rate per window size
	CurrentRate float64
}

// Check checks status of the rate limit key
func (r *RateLimiter) Check(key string) (limitStatus *LimitStatus, err error) {
	currentWindow := time.Now().UTC().Truncate(r.windowSize)
	previousWindow := currentWindow.Add(-r.windowSize)

	prevValue, currentValue, err := r.dataStore.Get(key, previousWindow, currentWindow)
	if err != nil {
		return nil, err
	}

	timeFromCurrWindow := time.Now().UTC().Sub(currentWindow)

	rate := float64((float64(r.windowSize)-float64(timeFromCurrWindow))/float64(r.windowSize))*float64(prevValue) + float64(currentValue)
	limitStatus = &LimitStatus{}

	if rate >= float64(r.requestsLimit) {
		limitStatus.IsLimited = true
		limitDuration := r.calcLimitDuration(prevValue, currentValue, timeFromCurrWindow)
		limitStatus.LimitDuration = &limitDuration
	}

	limitStatus.CurrentRate = rate

	return limitStatus, nil
}

// calcRate calculates current rate based on previous and current values
func (r *RateLimiter) calcRate(timeFromCurrWindow time.Duration, prevValue int64, currentValue int64) float64 { // nolint: unused
	return float64((float64(r.windowSize)-float64(timeFromCurrWindow))/float64(r.windowSize))*float64(prevValue) + float64(currentValue)
}

func (r *RateLimiter) calcLimitDuration(prevValue, currValue int64, timeFromCurrWindow time.Duration) time.Duration {
	var limitDuration time.Duration

	if prevValue == 0 {
		// unblock in the next window where prevValue is currValue and currValue is zero (assuming that since limit start all requests are blocked)
		if currValue != 0 {
			nextWindowUnblockPoint := float64(r.windowSize) * (1.0 - (float64(r.requestsLimit) / float64(currValue)))
			timeToNextWindow := r.windowSize - timeFromCurrWindow
			limitDuration = timeToNextWindow + time.Duration(int64(nextWindowUnblockPoint)+1)
		} else {
			// when requestsLimit is 0 we want to block all requests - set limitDuration to -1
			limitDuration = -1
		}
	} else {
		currWindowUnblockPoint := float64(r.windowSize) * (1.0 - (float64(r.requestsLimit-currValue) / float64(prevValue)))
		limitDuration = time.Duration(int64(currWindowUnblockPoint+1)) - timeFromCurrWindow
	}

	return limitDuration
}
