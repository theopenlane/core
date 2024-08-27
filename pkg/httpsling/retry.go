package httpsling

import (
	"math"
	"net/http"
	"time"
)

// RetryConfig defines the configuration for retrying requests
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// Strategy is the backoff strategy function
	Strategy BackoffStrategy
	// RetryIf is the custom retry condition function
	RetryIf RetryIfFunc
}

// RetryIfFunc defines the function signature for retry conditions
type RetryIfFunc func(req *http.Request, resp *http.Response, err error) bool

// BackoffStrategy defines a function that returns the delay before the next retry
type BackoffStrategy func(attempt int) time.Duration

// DefaultBackoffStrategy provides a simple constant delay between retries
func DefaultBackoffStrategy(delay time.Duration) func(int) time.Duration {
	return func(attempt int) time.Duration {
		return delay
	}
}

// LinearBackoffStrategy increases the delay linearly with each retry attempt
func LinearBackoffStrategy(initialInterval time.Duration) func(int) time.Duration {
	return func(attempt int) time.Duration {
		return initialInterval * time.Duration(attempt+1)
	}
}

// ExponentialBackoffStrategy increases the delay exponentially with each retry attempt
func ExponentialBackoffStrategy(initialInterval time.Duration, multiplier float64, maxBackoffTime time.Duration) func(int) time.Duration {
	return func(attempt int) time.Duration {
		delay := initialInterval * time.Duration(math.Pow(multiplier, float64(attempt)))

		if delay > maxBackoffTime {
			return maxBackoffTime
		}

		return delay
	}
}

// DefaultRetryIf is a simple retry condition that retries on 5xx status codes
func DefaultRetryIf(req *http.Request, resp *http.Response, err error) bool {
	return resp.StatusCode >= http.StatusInternalServerError || err != nil
}
