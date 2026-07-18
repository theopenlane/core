package operations

import (
	"math/rand/v2"
	"time"

	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	// DomainScanInitialWait is how long the poll cycle waits after submission before checking the scan result for the first time
	DomainScanInitialWait = 20 * time.Second
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

// DomainScanCreateEnvelope triggers submission of an organization's domains to Cloudflare's URL Scanner
type DomainScanCreateEnvelope struct {
	// OrganizationID is the organization that owns the domains being scanned
	OrganizationID string `json:"organizationId"`
	// Domains is the list of domains to submit for scanning
	Domains []string `json:"domains"`
	// ForceRefresh bypasses Cloudflare's Browser Rendering cache, forcing a fresh render
	// instead of reusing one from a previous scan of the same domain
	ForceRefresh bool `json:"forceRefresh,omitempty"`
}

// domainScanCreateSchemaName is the type name derived from the JSON schema reflector
var domainScanCreateSchemaName = jsonx.SchemaID(jsonx.SchemaFrom[DomainScanCreateEnvelope]())

var (
	// DomainScanCreateTopic is the Gala topic name for domain scan submission
	DomainScanCreateTopic = gala.TopicName("domainscan.create." + domainScanCreateSchemaName)
	// DomainScanCreateListenerName is the Gala listener name for the domain scan create handler
	DomainScanCreateListenerName = "domainscan.create." + domainScanCreateSchemaName + ".handler"
)

// DomainScanPollEnvelope carries one submitted scan through poll cycles until it's ready or the attempt budget is exhausted
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
}

// domainScanPollSchemaName is the type name derived from the JSON schema reflector
var domainScanPollSchemaName = jsonx.SchemaID(jsonx.SchemaFrom[DomainScanPollEnvelope]())

var (
	// DomainScanPollTopic is the Gala topic name for domain scan polling
	DomainScanPollTopic = gala.TopicName("domainscan.poll." + domainScanPollSchemaName)
	// DomainScanPollListenerName is the Gala listener name for the domain scan poll handler
	DomainScanPollListenerName = "domainscan.poll." + domainScanPollSchemaName + ".handler"
)
