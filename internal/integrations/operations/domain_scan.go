package operations

import (
	"context"
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

	jitter := time.Duration(rand.Int64N(int64(interval) / 4))

	return interval + jitter
}

// DomainScanCreateEnvelope triggers submission of an organization's domains to Cloudflare's URL Scanner
type DomainScanCreateEnvelope struct {
	// OrganizationID is the organization that owns the domains being scanned
	OrganizationID string `json:"organizationId"`
	// Domains is the list of domains to submit for scanning
	Domains []string `json:"domains"`
}

// domainScanCreateSchemaName is the type name derived from the JSON schema reflector
var domainScanCreateSchemaName = jsonx.SchemaID(jsonx.SchemaFrom[DomainScanCreateEnvelope]())

var (
	// DomainScanCreateTopic is the Gala topic name for domain scan submission
	DomainScanCreateTopic = gala.TopicName("domainscan.create." + domainScanCreateSchemaName)
	// domainScanCreateListenerName is the Gala listener name for the domain scan create handler
	domainScanCreateListenerName = "domainscan.create." + domainScanCreateSchemaName + ".handler"
)

// DomainScanCreateHandler submits an organization's domains for scanning
type DomainScanCreateHandler func(context.Context, DomainScanCreateEnvelope) error

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
	// domainScanPollListenerName is the Gala listener name for the domain scan poll handler
	domainScanPollListenerName = "domainscan.poll." + domainScanPollSchemaName + ".handler"
)

// DomainScanPollHandler processes one poll cycle for a submitted scan. It returns done=true once the scan has been fully processed
// (succeeded or given up), and done=false when the cycle re-emitted itself for another attempt
type DomainScanPollHandler func(context.Context, DomainScanPollEnvelope) (done bool, err error)

// RegisterDomainScanListeners registers the Gala listeners for domain scan submission and polling
func RegisterDomainScanListeners(runtime *gala.Gala, handleCreate DomainScanCreateHandler, handlePoll DomainScanPollHandler) error {
	if _, err := gala.RegisterListeners(runtime.Registry(), gala.Definition[DomainScanCreateEnvelope]{
		Topic: gala.Topic[DomainScanCreateEnvelope]{Name: DomainScanCreateTopic},
		Name:  domainScanCreateListenerName,
		Handle: func(hc gala.HandlerContext, envelope DomainScanCreateEnvelope) error {
			return handleCreate(hc.Context, envelope)
		},
	}); err != nil {
		return err
	}

	_, err := gala.RegisterListeners(runtime.Registry(), gala.Definition[DomainScanPollEnvelope]{
		Topic: gala.Topic[DomainScanPollEnvelope]{Name: DomainScanPollTopic},
		Name:  domainScanPollListenerName,
		Handle: func(hc gala.HandlerContext, envelope DomainScanPollEnvelope) error {
			_, err := handlePoll(hc.Context, envelope)
			return err
		},
	})

	return err
}
