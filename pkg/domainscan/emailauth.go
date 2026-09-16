package domainscan

import (
	"strconv"
	"strings"
)

// SPFPolicy is the qualifier on an SPF record's "all" mechanism, which is what actually
// decides how a receiver treats mail from a host the record doesn't authorize
type SPFPolicy string

const (
	// SPFPolicyHardFail is "-all": unauthorized senders are explicitly not permitted
	SPFPolicyHardFail SPFPolicy = "hard_fail"
	// SPFPolicySoftFail is "~all": unauthorized senders should be accepted but marked
	SPFPolicySoftFail SPFPolicy = "soft_fail"
	// SPFPolicyNeutral is "?all": the record expresses no opinion, which is equivalent to no SPF
	SPFPolicyNeutral SPFPolicy = "neutral"
	// SPFPolicyPassAll is "+all": every host on the internet is authorized, effectively always a misconfiguration
	SPFPolicyPassAll SPFPolicy = "pass_all"
)

// Valid reports whether the policy is one of the recognized qualifiers
func (p SPFPolicy) Valid() bool {
	switch p {
	case SPFPolicyHardFail, SPFPolicySoftFail, SPFPolicyNeutral, SPFPolicyPassAll:
		return true
	}

	return false
}

// DMARCPolicy is the enforcement policy from a DMARC record's p= or sp= tag
type DMARCPolicy string

const (
	// DMARCPolicyNone requests reporting only, with no instruction to block
	DMARCPolicyNone DMARCPolicy = "none"
	// DMARCPolicyQuarantine asks receivers to treat failing mail as suspicious
	DMARCPolicyQuarantine DMARCPolicy = "quarantine"
	// DMARCPolicyReject asks receivers to refuse failing mail outright
	DMARCPolicyReject DMARCPolicy = "reject"
)

// Valid reports whether the policy is one a receiver would honor
func (p DMARCPolicy) Valid() bool {
	switch p {
	case DMARCPolicyNone, DMARCPolicyQuarantine, DMARCPolicyReject:
		return true
	}

	return false
}

// ParseDMARCPolicy folds a p= or sp= tag value onto a DMARCPolicy, returning ok false when
// the value isn't a recognized policy so a malformed record doesn't score as enforcing
func ParseDMARCPolicy(value string) (DMARCPolicy, bool) {
	policy := DMARCPolicy(strings.ToLower(strings.TrimSpace(value)))

	return policy, policy.Valid()
}

// defaultDMARCPercentage is the pct= value assumed when the tag is absent, per RFC 7489
const defaultDMARCPercentage = 100

// EmailAuth is a domain's email authentication posture: the records that decide whether
// someone else can send mail that appears to come from it. Derived from the DNS enrichment,
// which is otherwise consumed for vendor attribution and not retained in the report
type EmailAuth struct {
	// SPFRecord is the raw SPF TXT record found at the apex, if any
	SPFRecord string `json:"spf_record,omitempty"`
	// SPFPolicy is the qualifier on the record's all mechanism, empty when the record has none
	SPFPolicy SPFPolicy `json:"spf_policy,omitempty"`
	// DMARCRecord is the raw DMARC TXT record found at _dmarc.<domain>, if any
	DMARCRecord string `json:"dmarc_record,omitempty"`
	// DMARCPolicy is the enforcement policy from the p= tag, empty when absent or unrecognized
	DMARCPolicy DMARCPolicy `json:"dmarc_policy,omitempty"`
	// DMARCSubdomainPolicy is the sp= tag, which overrides DMARCPolicy for subdomains when present
	DMARCSubdomainPolicy DMARCPolicy `json:"dmarc_subdomain_policy,omitempty"`
	// DMARCPercentage is the pct= tag: the share of failing mail the policy is applied to.
	// A p=reject at pct=10 enforces on a tenth of messages, so policy alone overstates protection
	DMARCPercentage int `json:"dmarc_percentage,omitempty"`
	// DMARCReportingURIs are the rua= aggregate report destinations, absent when nobody is reading the reports
	DMARCReportingURIs []string `json:"dmarc_reporting_uris,omitempty"`
	// DKIMSelectors are the DKIM selector labels found at conventional names. Absence is
	// suggestive rather than conclusive, since only common selectors are probed
	DKIMSelectors []string `json:"dkim_selectors,omitempty"`
	// MXHosts are the apex mail exchangers, in preference order
	MXHosts []string `json:"mx_hosts,omitempty"`
}

// Enforces reports whether the DMARC policy actually instructs receivers to act on
// failing mail across all of it, rather than reporting only or sampling a fraction
func (e EmailAuth) Enforces() bool {
	switch e.DMARCPolicy {
	case DMARCPolicyQuarantine, DMARCPolicyReject:
		return e.DMARCPercentage >= defaultDMARCPercentage
	default:
		return false
	}
}

// buildEmailAuth derives the email authentication section from the DNS enrichment,
// returning nil when no DNS lookup succeeded
func buildEmailAuth(enrichment Enrichment) *EmailAuth {
	dns := enrichment.DNS
	if dns == nil {
		return nil
	}

	auth := &EmailAuth{
		SPFRecord:     dns.SPFRecord,
		SPFPolicy:     spfAllQualifier(dns.SPFRecord),
		DMARCRecord:   dns.DMARCRecord,
		DKIMSelectors: dns.DKIMSelectors,
		MXHosts:       dns.MXHosts,
	}

	if policy, ok := ParseDMARCPolicy(dns.DMARCPolicy); ok {
		auth.DMARCPolicy = policy
	}

	if dns.DMARCRecord != "" {
		if policy, ok := ParseDMARCPolicy(dmarcTagValue(dns.DMARCRecord, "sp")); ok {
			auth.DMARCSubdomainPolicy = policy
		}

		auth.DMARCPercentage = dmarcPercentage(dns.DMARCRecord)
		auth.DMARCReportingURIs = dmarcReportingURIs(dns.DMARCRecord)
	}

	if auth.SPFRecord == "" && auth.DMARCRecord == "" && len(auth.DKIMSelectors) == 0 && len(auth.MXHosts) == 0 {
		return nil
	}

	return auth
}

// dmarcPercentage reads the pct= tag, defaulting to 100 when absent, unparsable or outside 0-100
func dmarcPercentage(record string) int {
	pct, err := strconv.Atoi(dmarcTagValue(record, "pct"))
	if err != nil || pct < 0 || pct > defaultDMARCPercentage {
		return defaultDMARCPercentage
	}

	return pct
}

// dmarcReportingURIs splits the rua= tag into its individual destinations
func dmarcReportingURIs(record string) []string {
	raw := dmarcTagValue(record, "rua")
	if raw == "" {
		return nil
	}

	var uris []string

	for _, uri := range strings.Split(raw, ",") {
		if uri = strings.TrimSpace(uri); uri != "" {
			uris = append(uris, uri)
		}
	}

	return uris
}

// spfAllQualifier finds the qualifier on the record's all mechanism. The first one wins,
// since evaluation stops at the first match and anything after it is unreachable
func spfAllQualifier(record string) SPFPolicy {
	for _, field := range strings.Fields(record) {
		switch strings.ToLower(field) {
		case "-all":
			return SPFPolicyHardFail
		case "~all":
			return SPFPolicySoftFail
		case "?all":
			return SPFPolicyNeutral
		case "all", "+all":
			return SPFPolicyPassAll
		}
	}

	return ""
}
