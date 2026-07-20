package domainscan

import (
	"slices"
	"sort"
	"strings"
)

// MergeReports combines one BuildScanReport output per domain (already vendor-enriched, as
// stored on each domain's own Scan.Metadata) into a single DomainScanReport: vendors,
// technologies, and systems are deduped by name across every domain; assets and findings are
// unioned; platform/compliance/meta/trust_center_settings take the first non-empty value found,
// since they describe the company rather than any one domain
func MergeReports(results []DomainScanResult, reports []map[string]any) DomainScanReport {
	merged := DomainScanReport{Scans: results}

	vendors := newNameDedup()
	technologies := newNameDedup()
	systems := newNameDedup()

	assets := newAssetMerge()
	findings := newFindingsMerge()

	for _, report := range reports {
		vendors.addAll(mapSlice(report["vendors"]), "name")
		technologies.addAll(mapSlice(report["technologies"]), "name")
		systems.addAll(mapSlice(report["systems"]), "system_name")

		if section, ok := asMap(report["assets"]); ok {
			assets.add(section)
		}

		if section, ok := asMap(report["findings"]); ok {
			findings.add(section)
		}

		if merged.Platform == nil {
			if section, ok := asMap(report["platform"]); ok && len(section) > 0 {
				merged.Platform = section
			}
		}

		if merged.Compliance == nil {
			if section, ok := asMap(report["compliance"]); ok && len(section) > 0 {
				merged.Compliance = section
			}
		}

		if merged.Meta == nil {
			if section, ok := asMap(report["meta"]); ok && len(section) > 0 {
				merged.Meta = section
			}
		}

		if merged.TrustCenterSettings == nil {
			if section, ok := asMap(report["trust_center_settings"]); ok && len(section) > 0 {
				merged.TrustCenterSettings = section
			}
		}

		if merged.Registrar == nil {
			if section, ok := asMap(report["registrar"]); ok && len(section) > 0 {
				merged.Registrar = section
			}
		}
	}

	if entries := vendors.entries(); len(entries) > 0 {
		merged.Vendors = entries
	}

	if entries := technologies.entries(); len(entries) > 0 {
		merged.Technologies = entries
	}

	if entries := systems.entries(); len(entries) > 0 {
		merged.Systems = entries
	}

	if section := assets.result(); len(section) > 0 {
		merged.Assets = section
	}

	if section := findings.result(); len(section) > 0 {
		merged.Findings = section
	}

	return merged
}

// asMap type-asserts raw as map[string]any, the shape a JSON object section decodes to
// whether or not it round-tripped through JSON storage first
func asMap(raw any) (map[string]any, bool) {
	m, ok := raw.(map[string]any)
	return m, ok
}

// mapSlice normalizes raw to []map[string]any, whether it's still []map[string]any (built
// in-process) or has round-tripped through JSON (e.g. read back from Scan.Metadata) and decoded
// as []any of map[string]any
func mapSlice(raw any) []map[string]any {
	switch v := raw.(type) {
	case []map[string]any:
		return v
	case []any:
		maps := make([]map[string]any, 0, len(v))

		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				maps = append(maps, m)
			}
		}

		return maps
	default:
		return nil
	}
}

// stringSlice normalizes raw to []string, whether it's still []string or has round-tripped
// through JSON and decoded as []any of string
func stringSlice(raw any) []string {
	switch v := raw.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))

		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}

		return out
	default:
		return nil
	}
}

// nameDedup accumulates map[string]any entries keyed by a lowercased name field (the field name
// varies by section, e.g. "name" for vendors/technologies, "system_name" for systems), keeping
// the first-seen entry for each name and preserving first-seen order
type nameDedup struct {
	seen  map[string]bool
	order []map[string]any
}

func newNameDedup() *nameDedup {
	return &nameDedup{seen: map[string]bool{}}
}

func (d *nameDedup) addAll(entries []map[string]any, nameField string) {
	for _, entry := range entries {
		name, _ := entry[nameField].(string)

		key := strings.ToLower(strings.TrimSpace(name))
		if key == "" || d.seen[key] {
			continue
		}

		d.seen[key] = true
		d.order = append(d.order, entry)
	}
}

func (d *nameDedup) entries() []map[string]any {
	return d.order
}

// assetMerge unions the dns_records, ip_addresses, and internal_domains sections of an assets
// map across every domain's report, deduping dns_records by domain+type, ip_addresses by
// address, and internal_domains by hostname
type assetMerge struct {
	dnsRecordsSeen      map[string]bool
	dnsRecords          []map[string]any
	ipAddrsSeen         map[string]bool
	ipAddresses         []map[string]any
	internalDomainsSeen map[string]bool
	internalDomains     []string
}

func newAssetMerge() *assetMerge {
	return &assetMerge{
		dnsRecordsSeen:      map[string]bool{},
		ipAddrsSeen:         map[string]bool{},
		internalDomainsSeen: map[string]bool{},
	}
}

func (a *assetMerge) add(assets map[string]any) {
	for _, record := range mapSlice(assets["dns_records"]) {
		domain, _ := record["domain"].(string)
		recordType, _ := record["type"].(string)

		key := strings.ToLower(domain) + "|" + strings.ToLower(recordType)
		if key == "|" || a.dnsRecordsSeen[key] {
			continue
		}

		a.dnsRecordsSeen[key] = true
		a.dnsRecords = append(a.dnsRecords, record)
	}

	for _, entry := range mapSlice(assets["ip_addresses"]) {
		address, _ := entry["address"].(string)
		if address == "" || a.ipAddrsSeen[address] {
			continue
		}

		a.ipAddrsSeen[address] = true
		a.ipAddresses = append(a.ipAddresses, entry)
	}

	for _, domain := range stringSlice(assets["internal_domains"]) {
		domain = strings.ToLower(strings.TrimSpace(domain))
		if domain == "" || a.internalDomainsSeen[domain] {
			continue
		}

		a.internalDomainsSeen[domain] = true
		a.internalDomains = append(a.internalDomains, domain)
	}
}

func (a *assetMerge) result() map[string]any {
	assets := map[string]any{}

	if len(a.dnsRecords) > 0 {
		assets["dns_records"] = a.dnsRecords
	}

	if len(a.ipAddresses) > 0 {
		assets["ip_addresses"] = a.ipAddresses
	}

	if len(a.internalDomains) > 0 {
		sort.Strings(a.internalDomains)
		assets["internal_domains"] = a.internalDomains
	}

	return assets
}

// findingsMerge unions security_violations and risks, ORs is_malicious, concatenates
// missing_compliance_links, and collects agent_readiness across every domain's findings section
type findingsMerge struct {
	securityViolations []string
	risks              []string
	isMalicious        bool
	missingLinks       []string
	agentReadiness     []map[string]any
}

func newFindingsMerge() *findingsMerge {
	return &findingsMerge{}
}

func (f *findingsMerge) add(findings map[string]any) {
	f.securityViolations = unionStrings(f.securityViolations, stringSlice(findings["security_violations"]))
	f.risks = unionStrings(f.risks, stringSlice(findings["risks"]))

	if malicious, ok := findings["is_malicious"].(bool); ok && malicious {
		f.isMalicious = true
	}

	if links, ok := findings["missing_compliance_links"].(string); ok && strings.TrimSpace(links) != "" {
		f.missingLinks = append(f.missingLinks, links)
	}

	if readiness, ok := findings["agent_readiness"].(map[string]any); ok && len(readiness) > 0 {
		f.agentReadiness = append(f.agentReadiness, readiness)
	}
}

func (f *findingsMerge) result() map[string]any {
	findings := map[string]any{}

	if len(f.securityViolations) > 0 {
		findings["security_violations"] = f.securityViolations
	}

	if len(f.risks) > 0 {
		findings["risks"] = f.risks
	}

	if f.isMalicious {
		findings["is_malicious"] = true
	}

	if len(f.missingLinks) > 0 {
		findings["missing_compliance_links"] = strings.Join(f.missingLinks, "\n\n")
	}

	if len(f.agentReadiness) > 0 {
		findings["agent_readiness"] = f.agentReadiness
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
