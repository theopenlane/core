package docextract

import (
	"math/rand/v2"
	"time"
)

// generationRetry describes a generation that should be re-requested
type generationRetry struct {
	// needed is true when another attempt should be made
	needed bool
	// quota is true when the request was rejected for exceeding a quota; it is returned to the caller
	// as a QuotaError rather than waited out in place
	quota bool
	// after is the delay the api asked for, zero when it supplied none
	after time.Duration
}

// wait reports how long to pause before re-requesting an interrupted generation, honoring the delay
// the api asked for when it supplied one
func (r generationRetry) wait() time.Duration {
	if r.after > 0 {
		return jitter(r.after)
	}

	return retryDelay
}

// jitter extends a delay by up to a quarter so concurrent batches do not come back at the same
// moment; it only adds, so a delay the api asked for is always honored in full
func jitter(delay time.Duration) time.Duration {
	return delay + time.Duration(rand.Int64N(int64(delay/4)+1)) //nolint:gosec,mnd
}
