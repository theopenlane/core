package domainscan

import (
	"slices"
	"strings"
)

// ScanReportInput pairs one domain with its typed scan report, for MergeReports. Domain is
// carried alongside the report (rather than embedded in it) since it identifies which scan
// produced the report, not something the report describes about itself
type ScanReportInput struct {
	// Domain is the domain that was scanned
	Domain string
	// Report is that domain's BuildScanReport output
	Report ScanReport
}

// MergeReports combines one BuildScanReport output per domain (already vendor-enriched, as
// stored on each domain's own Scan.Metadata) into a single report
func MergeReports(results []Result, reports []ScanReportInput) Report {
	merged := Report{Scans: results}

	vendors := newNameDedup(func(v Vendor) string { return v.Name })
	technologies := newNameDedup(func(t Technology) string { return t.Name })
	systems := newNameDedup(func(s SystemEntry) string { return s.SystemName })

	assets := newAssetMerge()
	findings := newFindingsMerge()

	for _, input := range reports {
		report := input.Report

		vendors.addAll(report.Vendors)
		technologies.addAll(report.Technologies)
		systems.addAll(report.Systems)

		assets.add(report.Assets)
		findings.add(input.Domain, report.Findings)

		firstNonNil(&merged.Platform, report.Platform)
		firstNonNil(&merged.Compliance, report.Compliance)
		firstNonNil(&merged.Meta, report.Meta)
		firstNonNil(&merged.Branding, report.Branding)
		firstNonNil(&merged.Registrar, report.Registrar)
	}

	merged.Vendors = vendors.entries()
	merged.Technologies = technologies.entries()
	merged.Systems = systems.entries()
	merged.Assets = assets.result()
	merged.Findings = findings.result()

	return merged
}

// firstNonNil sets *dst to src if *dst is still unset, so the first non-empty value found across
// every domain's report wins
func firstNonNil[T any](dst **T, src *T) {
	if *dst == nil && src != nil {
		*dst = src
	}
}

// nameDedup accumulates entries keyed by a lowercased, trimmed name keeping
// the first-seen entry for each name and preserving first-seen order
type nameDedup[T any] struct {
	seen   map[string]bool
	order  []T
	nameOf func(T) string
}

func newNameDedup[T any](nameOf func(T) string) *nameDedup[T] {
	return &nameDedup[T]{seen: map[string]bool{}, nameOf: nameOf}
}

func (d *nameDedup[T]) addAll(entries []T) {
	for _, entry := range entries {
		key := strings.ToLower(strings.TrimSpace(d.nameOf(entry)))
		if key == "" || d.seen[key] {
			continue
		}

		d.seen[key] = true
		d.order = append(d.order, entry)
	}
}

func (d *nameDedup[T]) entries() []T {
	return d.order
}

// assetMerge unions the dns_records and ip_addresses sections of an Assets section across every
// domain's report, deduping dns_records by domain+type and ip_addresses by address
type assetMerge struct {
	dnsRecordsSeen map[string]bool
	dnsRecords     []DNSRecord
	ipAddrsSeen    map[string]bool
	ipAddresses    []IPAddress
}

func newAssetMerge() *assetMerge {
	return &assetMerge{
		dnsRecordsSeen: map[string]bool{},
		ipAddrsSeen:    map[string]bool{},
	}
}

func (a *assetMerge) add(assets *Assets) {
	if assets == nil {
		return
	}

	for _, record := range assets.DNSRecords {
		key := strings.ToLower(record.Domain) + "|" + strings.ToLower(record.Type)
		if key == "|" || a.dnsRecordsSeen[key] {
			continue
		}

		a.dnsRecordsSeen[key] = true
		a.dnsRecords = append(a.dnsRecords, record)
	}

	for _, entry := range assets.IPAddresses {
		if entry.Address == "" || a.ipAddrsSeen[entry.Address] {
			continue
		}

		a.ipAddrsSeen[entry.Address] = true
		a.ipAddresses = append(a.ipAddresses, entry)
	}
}

func (a *assetMerge) result() *Assets {
	if len(a.dnsRecords) == 0 && len(a.ipAddresses) == 0 {
		return nil
	}

	return &Assets{
		DNSRecords:  a.dnsRecords,
		IPAddresses: a.ipAddresses,
	}
}

// findingsMerge unions security_violations and risks, ORs is_malicious, concatenates
// missing_compliance_links, and collects agent_readiness across every domain's findings section
type findingsMerge struct {
	securityViolations []string
	risks              []string
	isMalicious        bool
	missingLinks       []string
	agentReadiness     []AgentReadinessFinding
}

func newFindingsMerge() *findingsMerge {
	return &findingsMerge{}
}

func (f *findingsMerge) add(domain string, findings Findings) {
	f.securityViolations = unionStrings(f.securityViolations, findings.SecurityViolations)
	f.risks = unionStrings(f.risks, findings.Risks)

	if findings.IsMalicious {
		f.isMalicious = true
	}

	if strings.TrimSpace(findings.MissingComplianceLinks) != "" {
		f.missingLinks = append(f.missingLinks, findings.MissingComplianceLinks)
	}

	for _, finding := range findings.AgentReadiness {
		finding.Domain = domain
		f.agentReadiness = append(f.agentReadiness, finding)
	}
}

func (f *findingsMerge) result() Findings {
	findings := Findings{
		SecurityViolations: f.securityViolations,
		Risks:              f.risks,
		IsMalicious:        f.isMalicious,
		AgentReadiness:     f.agentReadiness,
	}

	if len(f.missingLinks) > 0 {
		findings.MissingComplianceLinks = strings.Join(f.missingLinks, "\n\n")
	}

	return findings
}

// unionStrings appends values from add not already present in existing, preserving order
func unionStrings(existing, add []string) []string {
	for _, v := range add {
		if !slices.Contains(existing, v) {
			existing = append(existing, v)
		}
	}

	return existing
}
