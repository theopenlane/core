package cloudflare

import (
	"strconv"

	"math/rand/v2"
	"time"

	"github.com/theopenlane/core/v2/pkg/gala"
)

const (
	// DomainScanPollMinInterval is the wait before the first retry, and the backoff's starting point
	DomainScanPollMinInterval = 10 * time.Second
	// DomainScanPollMaxInterval caps how long the backoff is allowed to grow to between poll cycles
	DomainScanPollMaxInterval = 60 * time.Second
	// DomainScanMaxAttempts bounds how many poll cycles are attempted before giving up on a scan
	DomainScanMaxAttempts = 30
)

// DomainScanPollBackoff returns the jittered wait before the next poll cycle for a scan
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

// DomainScanPollEnvelope carries one submitted scan through poll cycles until it's ready
type DomainScanPollEnvelope struct {
	// OrganizationID is the organization that owns the scan
	OrganizationID string `json:"organizationId"`
	// ScanResultID is the scan ID returned by Cloudflare's URL Scanner on submission
	ScanResultID string `json:"scanResultId"`
	// InternalScanID is the id of the Scan record created when the scan was submitted
	InternalScanID string `json:"internalScanId"`
	// Attempt is the number of poll cycles already attempted for this scan
	Attempt int `json:"attempt"`
	// SiblingScanIDs lists every internal Scan ID submitted together with this one
	SiblingScanIDs []string `json:"siblingScanIds"`
}

// domainScanTopics is the namespace for domain scan saga topics
var domainScanTopics = gala.IntegrationRun.At("domainscan.poll")

// domainScanPollTopic is the durable poll topic, keyed per-attempt to dedup crash-retry emits
var domainScanPollTopic = gala.NamespacedTopicFor(domainScanTopics, gala.WithUniqueKey(func(e DomainScanPollEnvelope) string {
	return domainScanTopics.Key(e.InternalScanID, strconv.Itoa(e.Attempt))
}))
