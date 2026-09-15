package domainscan

import "strings"

// complianceTypeOther is the extraction's escape hatch for a page that is legal or
// compliance-adjacent but not one of the recognized document types
const complianceTypeOther = "other"

// complianceDocumentTypes is the closed set of document types a compliance link may carry.
// The browser rendering schema constrains the extraction to these, and link discovery maps
// onto them, so every consumer compares the same vocabulary
var complianceDocumentTypes = []string{
	"privacy_policy",
	"terms_of_service",
	"trust_center",
	"dpa",
	"soc2_report",
	"security",
	"subprocessors",
	"gdpr",
	"cookie_policy",
	complianceTypeOther,
}

// complianceTypeAliases maps the other names these documents go by onto the canonical type.
// It exists for URLs: a site's path is what link discovery reads, and no site spells one
// "/privacy_policy". It is "/legal/privacy", "/tos", "/trust", "/cookies". Stored reports
// scanned before the extraction schema constrained its output benefit too.
//
// Keys are in compacted form, with separators removed, because separators are handled
// mechanically below. So this table holds only genuinely different words: "privacy" for
// "privacy_policy" earns its place, while "privacy-policy" and "privacypolicy" do not
var complianceTypeAliases = map[string]string{
	"privacy":                 "privacy_policy",
	"privacynotice":           "privacy_policy",
	"terms":                   "terms_of_service",
	"termsofuse":              "terms_of_service",
	"tos":                     "terms_of_service",
	"dataprocessing":          "dpa",
	"dataprocessingagreement": "dpa",
	"subprocessor":            "subprocessors",
	"subprocessorlist":        "subprocessors",
	"trust":                   "trust_center",
	"cookies":                 "cookie_policy",
	"securitypolicy":          "security",
}

// canonicalComplianceTypes maps each recognized document type's compacted form back to the
// type, so "privacy-policy", "Privacy Policy" and "privacypolicy" all resolve without an
// alias entry each. "other" is excluded: it names a page that is legal-adjacent rather than
// one of the documents
var canonicalComplianceTypes = func() map[string]string {
	set := make(map[string]string, len(complianceDocumentTypes))

	for _, t := range complianceDocumentTypes {
		if t != complianceTypeOther {
			set[compactComplianceType(t)] = t
		}
	}

	return set
}()

// NormalizeComplianceType folds a document type onto its canonical form, tolerating case,
// spacing, hyphenation and the alternative names above. A value it does not recognize is
// returned in normalized form rather than dropped, so a caller can still compare it and a
// type added to the vocabulary later needs no alias to work.
//
// Every consumer of a compliance document type goes through this. Comparing raw strings works
// right up until a footer links "/legal/privacy", at which point a document the site plainly
// publishes is reported as missing
func NormalizeComplianceType(value string) string {
	normalized := normalizeComplianceSeparators(value)
	if normalized == "" {
		return ""
	}

	compact := compactComplianceType(normalized)

	if canonical, ok := canonicalComplianceTypes[compact]; ok {
		return canonical
	}

	if canonical, ok := complianceTypeAliases[compact]; ok {
		return canonical
	}

	return normalized
}

// IsCanonicalComplianceType reports whether value names one of the recognized document types
func IsCanonicalComplianceType(value string) bool {
	_, ok := canonicalComplianceTypes[compactComplianceType(NormalizeComplianceType(value))]

	return ok
}

// MatchesComplianceType reports whether an extracted type denotes the given canonical one
func MatchesComplianceType(extracted, canonical string) bool {
	normalized := NormalizeComplianceType(extracted)

	return normalized != "" && normalized == canonical
}

// normalizeComplianceSeparators lowercases a type and folds spaces and hyphens to single
// underscores, so "Privacy Policy" and "privacy-policy" both reduce to "privacy_policy"
func normalizeComplianceSeparators(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")

	for strings.Contains(value, "__") {
		value = strings.ReplaceAll(value, "__", "_")
	}

	return strings.Trim(value, "_")
}

// compactComplianceType strips separators entirely, so a lookup is insensitive to whether a
// name was written as one word or several
func compactComplianceType(value string) string {
	return strings.ReplaceAll(value, "_", "")
}
