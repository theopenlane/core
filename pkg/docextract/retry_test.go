package docextract

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestWaitUsesFixedDelayWhenNotQuota(t *testing.T) {
	assert.Check(t, generationRetry{needed: true}.wait() == retryDelay)
}

func TestWaitHonorsRequestedDelay(t *testing.T) {
	retry := generationRetry{needed: true, after: 30 * time.Second}

	assert.Check(t, retry.wait() >= 30*time.Second)
	assert.Check(t, retry.wait() <= 30*time.Second+30*time.Second/4)
}
