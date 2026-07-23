package domainscan

import (
	"strings"
)

// mergeTrustCenterIntoCompliancePage folds a TrustCenterPage's findings into a
// CompliancePage fetched from the main domain, treating the trust center as
// the source of truth for frameworks, controls, subprocessors, and SOC 2
// status
func mergeTrustCenterIntoCompliancePage(comp *CompliancePage, tc *TrustCenterPage, trustURL string) *CompliancePage {
	merged := *comp

	subprocessors := tc.Subprocessors
	if hostedBy := tc.HostedBy; hostedBy != "" && !strings.EqualFold(hostedBy, "self-hosted") {
		// the platform hosting the trust center is itself a vendor the company relies on
		subprocessors = append(subprocessors, hostedBy)
	}

	merged.Frameworks = mergeStrings(tc.Frameworks, comp.Frameworks)
	merged.SOC2Certified = comp.SOC2Certified || tc.SOC2Certified
	merged.Subprocessors = mergeStrings(subprocessors, comp.Subprocessors)
	merged.Controls = mergeStrings(tc.Controls, comp.Controls)
	merged.TrustCenterHostedBy = tc.HostedBy
	merged.Documents = tc.Documents
	merged.ComplianceLinks = mergeComplianceLinks(comp.ComplianceLinks, []ComplianceLink{{URL: trustURL, Type: "trust_center"}})

	return &merged
}

// mergeTrustCenterPages combines the results of probing multiple trust center
// URLs (e.g. the root and known subpaths) into a single TrustCenterPage,
// unioning list fields and preferring the first non-empty HostedBy value seen
func mergeTrustCenterPages(pages ...*TrustCenterPage) *TrustCenterPage {
	merged := &TrustCenterPage{}

	for _, p := range pages {
		if p == nil {
			continue
		}

		if merged.HostedBy == "" || strings.EqualFold(merged.HostedBy, "self-hosted") {
			merged.HostedBy = p.HostedBy
		}

		merged.Frameworks = mergeStrings(merged.Frameworks, p.Frameworks)
		merged.SOC2Certified = merged.SOC2Certified || p.SOC2Certified
		merged.Controls = mergeStrings(merged.Controls, p.Controls)
		merged.Documents = mergeTrustDocuments(merged.Documents, p.Documents)
		merged.Subprocessors = mergeStrings(merged.Subprocessors, p.Subprocessors)
	}

	return merged
}

// mergeTrustDocuments unions two TrustDocument slices, deduplicating by name
// while preserving first-seen order
func mergeTrustDocuments(a, b []TrustDocument) []TrustDocument {
	seen := make(map[string]struct{}, len(a)+len(b))
	merged := make([]TrustDocument, 0, len(a)+len(b))

	for _, doc := range append(append([]TrustDocument{}, a...), b...) {
		if doc.Name == "" {
			continue
		}

		if _, ok := seen[doc.Name]; ok {
			continue
		}

		seen[doc.Name] = struct{}{}
		merged = append(merged, doc)
	}

	return merged
}

// mergeComplianceLinks unions two ComplianceLink slices, deduplicating by URL
// while preserving first-seen order
func mergeComplianceLinks(a, b []ComplianceLink) []ComplianceLink {
	seen := make(map[string]struct{}, len(a)+len(b))
	merged := make([]ComplianceLink, 0, len(a)+len(b))

	for _, link := range append(append([]ComplianceLink{}, a...), b...) {
		if link.URL == "" {
			continue
		}

		if _, ok := seen[link.URL]; ok {
			continue
		}

		seen[link.URL] = struct{}{}
		merged = append(merged, link)
	}

	return merged
}
