package docextract

import (
	"testing"
	"time"

	"google.golang.org/genai"
	"gotest.tools/v3/assert"
)

func TestRetryForQuotaStatus(t *testing.T) {
	retry, ok := retryFor(genai.APIError{Status: statusResourceExhausted})

	assert.Check(t, ok)
	assert.Check(t, retry.needed)
	assert.Check(t, retry.quota)
}

func TestRetryForTransientStatusIsNotQuota(t *testing.T) {
	retry, ok := retryFor(genai.APIError{Status: statusUnavailable})

	assert.Check(t, ok)
	assert.Check(t, retry.needed)
	assert.Check(t, !retry.quota)
}

func TestRetryForNonRetryableStatus(t *testing.T) {
	_, ok := retryFor(genai.APIError{Status: "INVALID_ARGUMENT"})

	assert.Check(t, !ok)
}

func TestRetryHonorsRetryInfo(t *testing.T) {
	retry, ok := retryFor(genai.APIError{
		Status: statusResourceExhausted,
		Details: []map[string]any{
			{"@type": "type.googleapis.com/google.rpc.Help"},
			{"@type": retryInfoType, "retryDelay": "42s"},
		},
	})

	assert.Check(t, ok)
	assert.Check(t, retry.after == 42*time.Second)
	assert.Check(t, retry.wait(0) >= 42*time.Second)
}

func TestRetryIgnoresMalformedRetryInfo(t *testing.T) {
	retry, ok := retryFor(genai.APIError{
		Status:  statusResourceExhausted,
		Details: []map[string]any{{"@type": retryInfoType, "retryDelay": "soon"}},
	})

	assert.Check(t, ok)
	assert.Check(t, retry.after == 0)
}

func TestWaitUsesFixedDelayWhenNotQuota(t *testing.T) {
	assert.Check(t, generationRetry{needed: true}.wait(5) == retryDelay)
}

func TestWaitBacksOffForQuota(t *testing.T) {
	retry := generationRetry{needed: true, quota: true}

	assert.Check(t, retry.wait(0) >= quotaBaseDelay)
	assert.Check(t, retry.wait(3) >= 8*quotaBaseDelay)
	assert.Check(t, retry.wait(0) < retry.wait(3))
}

func TestWaitCapsQuotaBackoff(t *testing.T) {
	retry := generationRetry{needed: true, quota: true}

	assert.Check(t, retry.wait(50) >= maxQuotaDelay)
	assert.Check(t, retry.wait(50) <= maxQuotaDelay+maxQuotaDelay/4)
}
