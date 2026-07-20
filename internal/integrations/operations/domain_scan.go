package operations

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
	// DomainScanPollIntervalMax caps how long the backoff is allowed togrow to between poll cycles
	DomainScanPollIntervalMax = 60 * time.Second
	// DomainScanMaxAttempts bounds how many poll cycles are attempted before giving up on a scan
	DomainScanMaxAttempts = 30
	// DomainScanPerformedBy marks a Scan record as one the system should actually submit to the cloudflare domain scan job
	DomainScanPerformedBy = "openlane_domain_scan"
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

// DomainScanSubmitEnvelope submits an already-created scan object record to the domain scan
type DomainScanSubmitEnvelope struct {
	// OrganizationID is the organization that owns the scan
	OrganizationID string `json:"organizationId"`
	// ScanID is the id of the already-created Scan record to submit
	ScanID string `json:"scanId"`
	// Domain is the target domain of the scan
	Domain string `json:"domain"`
	// ForceRefresh bypasses Cloudflare's Browser Rendering cache, forcing a fresh render
	// instead of reusing one from a previous scan of the same domain
	ForceRefresh bool `json:"forceRefresh,omitempty"`
}

// domainScanSubmitExistingSchemaName is the type name derived from the JSON schema reflector
var domainScanSubmitExistingSchemaName = jsonx.SchemaID(jsonx.SchemaFrom[DomainScanSubmitEnvelope]())

var (
	// DomainScanSubmitTopic is the Gala topic name for submitting an already-created scan object
	DomainScanSubmitTopic = gala.TopicName("domainscan.submit." + domainScanSubmitExistingSchemaName)
	// DomainScanSubmitListenerName is the Gala listener name for the domain scan submit-existing handler
	DomainScanSubmitListenerName = "domainscan.submit." + domainScanSubmitExistingSchemaName + ".handler"
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
	// SiblingScanIDs lists every internal Scan ID submitted together with this one (a
	// single-element slice for a one-off scan), so the last one to finish can gather and
	// combine every sibling's report into a single notification
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

// ImportDomainScanReviewVendor is one vendor the reviewer accepted, keyed by a client-assigned
// Ref so ImportDomainScanReviewPlatform/ImportDomainScanReviewSystem can reference it before it
// has a real Entity ID
type ImportDomainScanReviewVendor struct {
	// Ref is a client-assigned identifier for this vendor, referenced by EntityRefs elsewhere in the envelope
	Ref string `json:"ref"`
	// Name is the vendor's name
	Name string `json:"name"`
	// LegalName is the vendor's raw legal entity name, if known and different from Name
	LegalName string `json:"legalName,omitempty"`
	// Domain is the vendor's domain, if known
	Domain string `json:"domain,omitempty"`
	// Categories are the vendor's detected categories
	Categories []string `json:"categories,omitempty"`
}

// ImportDomainScanReviewAsset is one asset the reviewer accepted, keyed by a client-assigned Ref
// so ImportDomainScanReviewPlatform/ImportDomainScanReviewSystem can reference it before it has a
// real Asset ID
type ImportDomainScanReviewAsset struct {
	// Ref is a client-assigned identifier for this asset, referenced by AssetRefs elsewhere in the envelope
	Ref string `json:"ref"`
	// Name is the asset's display name
	Name string `json:"name"`
	// Identifier is the asset's domain, IP, or other unique identifier
	Identifier string `json:"identifier,omitempty"`
	// Website is the asset's URL, if known
	Website string `json:"website,omitempty"`
	// Categories are the asset's detected categories
	Categories []string `json:"categories,omitempty"`
}

// ImportDomainScanReviewPlatform is one accepted platform, linked to a subset of the accepted
// vendors/assets, and keyed by a client-assigned Ref so ImportDomainScanReviewSystem can
// reference it before it has a real Platform ID
type ImportDomainScanReviewPlatform struct {
	// Ref is a client-assigned identifier for this platform, referenced by PlatformRefs elsewhere in the envelope
	Ref string `json:"ref"`
	// Name is the platform's name
	Name string `json:"name"`
	// Description is the platform's description
	Description string `json:"description,omitempty"`
	// EntityRefs are the Refs of accepted vendors linked to this platform
	EntityRefs []string `json:"entityRefs,omitempty"`
	// AssetRefs are the Refs of accepted assets linked to this platform
	AssetRefs []string `json:"assetRefs,omitempty"`
}

// ImportDomainScanReviewSystem is one accepted system detail, linked to its own subset of the
// accepted vendors/assets/platforms
type ImportDomainScanReviewSystem struct {
	// Name is the system's name
	Name string `json:"name"`
	// Description is the system's description
	Description string `json:"description,omitempty"`
	// EntityRefs are the Refs of accepted vendors linked to this system
	EntityRefs []string `json:"entityRefs,omitempty"`
	// AssetRefs are the Refs of accepted assets linked to this system
	AssetRefs []string `json:"assetRefs,omitempty"`
	// PlatformRefs are the Refs of accepted platforms this system belongs to
	PlatformRefs []string `json:"platformRefs,omitempty"`
}

// ImportDomainScanReviewFinding is one accepted finding
type ImportDomainScanReviewFinding struct {
	// Category is the finding's category
	Category string `json:"category,omitempty"`
	// Description is the finding's description
	Description string `json:"description,omitempty"`
	// Severity is the finding's severity
	Severity string `json:"severity,omitempty"`
	// Domain is the domain this finding was raised against, if known
	Domain string `json:"domain,omitempty"`
}

// ImportDomainScanReviewEnvelope carries what a reviewer accepted from a domain scan report so
// it can be turned into real Platform/SystemDetail/Entity/Asset/Finding records asynchronously,
// off the GraphQL request that submitted it
type ImportDomainScanReviewEnvelope struct {
	// OrganizationID is the organization the created records belong to
	OrganizationID string `json:"organizationId"`
	// ScanIDs are the Scan records the created records should link back to
	ScanIDs []string `json:"scanIds"`
	// Platforms are the accepted platforms, if any
	Platforms []ImportDomainScanReviewPlatform `json:"platforms,omitempty"`
	// Systems are the accepted system details
	Systems []ImportDomainScanReviewSystem `json:"systems,omitempty"`
	// Vendors are the accepted vendors
	Vendors []ImportDomainScanReviewVendor `json:"vendors"`
	// Assets are the accepted assets
	Assets []ImportDomainScanReviewAsset `json:"assets"`
	// Findings are the accepted findings
	Findings []ImportDomainScanReviewFinding `json:"findings,omitempty"`
}

// domainScanImportSchemaName is the type name derived from the JSON schema reflector
var domainScanImportSchemaName = jsonx.SchemaID(jsonx.SchemaFrom[ImportDomainScanReviewEnvelope]())

var (
	// DomainScanImportTopic is the Gala topic name for importing an accepted domain scan review
	DomainScanImportTopic = gala.TopicName("domainscan.import." + domainScanImportSchemaName)
	// DomainScanImportListenerName is the Gala listener name for the domain scan import handler
	DomainScanImportListenerName = "domainscan.import." + domainScanImportSchemaName + ".handler"
)
