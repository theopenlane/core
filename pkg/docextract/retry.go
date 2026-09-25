package docextract

import (
	"math/rand/v2"
	"strings"
	"time"

	"google.golang.org/genai"
)

const (
	// quotaBaseDelay is the wait after the first quota rejection, doubled on each further one
	quotaBaseDelay = 20 * time.Second
	// maxQuotaDelay caps the quota backoff
	maxQuotaDelay = 5 * time.Minute
	// retryInfoType is the error detail carrying the delay the api asks callers to wait
	retryInfoType = "type.googleapis.com/google.rpc.RetryInfo"
)

// generationRetry describes a generation that should be re-requested
type generationRetry struct {
	// needed is true when another attempt should be made
	needed bool
	// quota is true when the request was rejected for exceeding a quota; the batches running
	// alongside it were rejected too, so they must not all come back at the same moment
	quota bool
	// after is the delay the api asked for, zero when it supplied none
	after time.Duration
}

// retryFor builds the advice for an api error, reporting false when the error is not retryable
func retryFor(apiErr genai.APIError) (generationRetry, bool) {
	if !isRetryableStatus(apiErr.Status) {
		return generationRetry{}, false
	}

	return generationRetry{
		needed: true,
		quota:  strings.EqualFold(apiErr.Status, statusResourceExhausted),
		after:  retryInfoDelay(apiErr.Details),
	}, true
}

// retryInfoDelay reads the delay google returns alongside a quota rejection
func retryInfoDelay(details []map[string]any) time.Duration {
	for _, detail := range details {
		if kind, _ := detail["@type"].(string); kind != retryInfoType {
			continue
		}

		delay, ok := detail["retryDelay"].(string)
		if !ok {
			continue
		}

		parsed, err := time.ParseDuration(delay)
		if err != nil {
			continue
		}

		return parsed
	}

	return 0
}

// wait reports how long to pause before the next attempt; a quota rejection backs off
// exponentially and is spread so concurrent batches do not retry in lockstep
func (r generationRetry) wait(attempt int) time.Duration {
	if r.after > 0 {
		return jitter(r.after)
	}

	if !r.quota {
		return retryDelay
	}

	backoff := quotaBaseDelay

	for range attempt {
		backoff *= 2

		if backoff >= maxQuotaDelay {
			return jitter(maxQuotaDelay)
		}
	}

	return jitter(backoff)
}

// jitter extends a delay by up to a quarter so concurrent batches do not come back at the same
// moment; it only adds, so a delay the api asked for is always honored in full
func jitter(delay time.Duration) time.Duration {
	return delay + time.Duration(rand.Int64N(int64(delay/4)+1))
}
