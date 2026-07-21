package cloudflare

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	// DomainScanPollIntervalMin is the wait before the first retry, and the starting point the backoff doubles from
	DomainScanPollIntervalMin = 10 * time.Second
	// DomainScanPollIntervalMax caps how long the backoff is allowed to grow to between poll cycles
	DomainScanPollIntervalMax = 60 * time.Second
	// DomainScanMaxAttempts bounds how many poll cycles are attempted before giving up on a scan
	DomainScanMaxAttempts = 30
)

// DomainScanPollBackoff returns the wait before the next poll cycle for a scan that's still
// processing. The interval doubles from DomainScanPollIntervalMin up to DomainScanPollIntervalMax
// as attempt grows, so slow scans are checked less often instead of exhausting the attempt budget
// at a flat cadence. Jitter is added on top to desynchronize scans that were submitted together
// and would otherwise poll Cloudflare in lockstep
func DomainScanPollBackoff(attempt int) time.Duration {
	interval := DomainScanPollIntervalMin
	for i := 0; i < attempt && interval < DomainScanPollIntervalMax; i++ {
		interval *= 2
	}

	if interval > DomainScanPollIntervalMax {
		interval = DomainScanPollIntervalMax
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

// domainScanPollSchemaName is the type name derived from the JSON schema reflector
var domainScanPollSchemaName = jsonx.SchemaID(jsonx.SchemaFrom[DomainScanPollEnvelope]())

var (
	// DomainScanPollTopic is the Gala topic name for domain scan polling
	DomainScanPollTopic = gala.TopicName("domainscan.poll." + domainScanPollSchemaName)
	// DomainScanPollListenerName is the Gala listener name for the domain scan poll handler
	DomainScanPollListenerName = "domainscan.poll." + domainScanPollSchemaName + ".handler"
)

// DomainScanPollHandler processes one poll cycle for a submitted scan. It returns done=true once the scan has been fully processed
// (succeeded or given up), and done=false when the cycle re-emitted itself for another attempt
type DomainScanPollHandler func(context.Context, DomainScanPollEnvelope) (done bool, err error)
