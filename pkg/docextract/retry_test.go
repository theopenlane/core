package docextract

import (
	"testing"

	"gotest.tools/v3/assert"
)

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
