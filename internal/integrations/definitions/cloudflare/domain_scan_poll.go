package cloudflare

import (
	"strconv"

	"math/rand/v2"
	"time"

	"github.com/theopenlane/core/v2/pkg/gala"
)

const (
	// DomainScanPollMinInterval is the wait before the first retry, and the starting point the backoff doubles from
	DomainScanPollMinInterval = 10 * time.Second
	// DomainScanPollMaxInterval caps how long the backoff is allowed to grow to between poll cycles
	DomainScanPollMaxInterval = 60 * time.Second
	// DomainScanMaxAttempts bounds how many poll cycles are attempted before giving up on a scan
	DomainScanMaxAttempts = 30
)

// DomainScanPollBackoff returns the wait before the next poll cycle for a scan that's still
// processing. The interval doubles from DomainScanPollMinInterval up to DomainScanPollMaxInterval
// as attempt grows, so slow scans are checked less often instead of exhausting the attempt budget
// at a flat cadence. Jitter is added on top to desynchronize scans that were submitted together
// and would otherwise poll Cloudflare in lockstep
func DomainScanPollBackoff(attempt int) time.Duration {
	interval := DomainScanPollMinInterval
	for i := 0; i < attempt && interval < DomainScanPollMaxInterval; i++ {
		interval *= 2
	}

	if interval > DomainScanPollMaxInterval {
		interval = DomainScanPollMaxInterval
	}

	jitter := time.Duration(rand.Int64N(int64(interval) / 4)) //nolint:gosec,mnd

	return interval + jitter
}

// DomainScanPollEnvelope carries one submitted scan through poll cycles until it's ready or the attempt budget is exhausted.
// This stays on raw Gala pub/sub rather than a Dispatch-able Operation because the poll cycle
// self-reschedules via Gala's ScheduledAt header, which Dispatch has no equivalent for
type DomainScanPollEnvelope struct {
	// OrganizationID is the organization that owns the scan
	OrganizationID string `json:"organizationId"`
	// ScanResultID is the scan ID returned by Cloudflare's URL Scanner on submission
	ScanResultID string `json:"scanResultId"`
	// InternalScanID is the id of the Scan record created when the scan was submitted;
	// the poll cycle updates this same record rather than creating a new one on completion
	InternalScanID string `json:"internalScanId"`
	// Attempt is the number of poll cycles already attempted for this scan
	Attempt int `json:"attempt"`
	// SiblingScanIDs lists every internal Scan ID submitted together with this one
	// so the last one to finish can gather and combine every sibling's report into a single notification
	SiblingScanIDs []string `json:"siblingScanIds"`
}

// domainScanTopics is the namespace for domain scan saga topics
var domainScanTopics = gala.IntegrationRun.At("domainscan.poll")

// domainScanPollTopic is the durable poll topic: the name derives from the envelope type
// under the domain scan namespace; per-attempt keys dedup crash-retry re-emissions of the poll chain
var domainScanPollTopic = gala.NamespacedTopicFor(domainScanTopics, gala.WithUniqueKey(func(e DomainScanPollEnvelope) string {
	return domainScanTopics.Key(e.InternalScanID, strconv.Itoa(e.Attempt))
}))
