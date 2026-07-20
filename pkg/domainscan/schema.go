package domainscan

import (
	"github.com/theopenlane/core/pkg/jsonx"
)

// DomainScanReport is the JSON-schema-described payload attached to a domain scan completion
// Notification. Scans holds one entry per domain in the originating submission (a single entry
// for a one-off scan, one entry per domain for an org-settings batch); every other section is
// the union of that data across every successfully completed domain — the same sections
// (vendors/technologies/assets/findings/platform/systems/compliance) BuildScanReport already
// produces per domain today, just merged instead of duplicated per notification
type DomainScanReport struct {
	// Scans is one entry per domain in the originating submission, carrying its own identity and outcome
	Scans []DomainScanResult `json:"scans"`
	// Vendors is the union of vendors detected across every completed domain, deduped by name
	Vendors []map[string]any `json:"vendors,omitempty"`
	// Technologies is the union of technologies detected across every completed domain, deduped by name
	Technologies []map[string]any `json:"technologies,omitempty"`
	// Assets is the union of DNS records, IP addresses, and internal domains across every completed domain
	Assets map[string]any `json:"assets,omitempty"`
	// TrustCenterSettings is the first non-empty trust center settings section found across completed domains
	TrustCenterSettings map[string]any `json:"trust_center_settings,omitempty"`
	// Findings is the union of findings across every completed domain
	Findings map[string]any `json:"findings,omitempty"`
	// Meta is the first non-empty meta section found across completed domains
	Meta map[string]any `json:"meta,omitempty"`
	// Platform is the first non-empty platform section found across completed domains
	Platform map[string]any `json:"platform,omitempty"`
	// Systems is the union of systems detected across every completed domain, deduped by name
	Systems []map[string]any `json:"systems,omitempty"`
	// Compliance is the first non-empty compliance section found across completed domains
	Compliance map[string]any `json:"compliance,omitempty"`
	// Registrar is the first non-empty WHOIS registration section found across completed domains
	Registrar map[string]any `json:"registrar,omitempty"`
}

// DomainScanResult is one domain's outcome within a DomainScanReport
type DomainScanResult struct {
	// Domain is the scanned domain
	Domain string `json:"domain"`
	// InternalScanID is the Scan record id this result belongs to
	InternalScanID string `json:"internal_scan_id"`
	// ExternalScanID is the Cloudflare URL Scanner task id, present when the scan completed
	ExternalScanID string `json:"external_scan_id,omitempty"`
	// URL is the scanned URL, present when the scan completed
	URL string `json:"url,omitempty"`
	// Status is "completed" or "failed"
	Status string `json:"status"`
}

// DomainScanReportSchema is the reflected JSON schema for DomainScanReport, the shape of a
// domain scan completion Notification's data
var DomainScanReportSchema = jsonx.SchemaFrom[DomainScanReport]()
