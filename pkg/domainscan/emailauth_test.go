package domainscan

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestSPFAllQualifier(t *testing.T) {
	tests := []struct {
		name   string
		record string
		want   string
	}{
		{name: "empty record", record: "", want: ""},
		{name: "hard fail", record: "v=spf1 include:_spf.google.com -all", want: SPFPolicyHardFail},
		{name: "soft fail", record: "v=spf1 include:_spf.google.com ~all", want: SPFPolicySoftFail},
		{name: "neutral", record: "v=spf1 ?all", want: SPFPolicyNeutral},
		{name: "explicit pass all", record: "v=spf1 +all", want: SPFPolicyPassAll},
		{name: "bare all is pass all", record: "v=spf1 all", want: SPFPolicyPassAll},
		{name: "uppercase qualifier", record: "v=spf1 -ALL", want: SPFPolicyHardFail},
		{name: "no all mechanism at all", record: "v=spf1 include:_spf.google.com", want: ""},
		{name: "last all mechanism wins", record: "v=spf1 ~all -all", want: SPFPolicyHardFail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, spfAllQualifier(tt.record), tt.want)
		})
	}
}

func TestDMARCPercentage(t *testing.T) {
	tests := []struct {
		name   string
		record string
		want   int
	}{
		{name: "absent pct defaults to 100", record: "v=DMARC1; p=reject", want: 100},
		{name: "explicit pct", record: "v=DMARC1; p=reject; pct=10", want: 10},
		{name: "pct of zero is honored", record: "v=DMARC1; p=reject; pct=0", want: 0},
		{name: "unparsable pct falls back to 100", record: "v=DMARC1; p=reject; pct=most", want: 100},
		{name: "out of range pct clamps to 100", record: "v=DMARC1; p=reject; pct=250", want: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, dmarcPercentage(tt.record), tt.want)
		})
	}
}

func TestNormalizeDMARCPolicy(t *testing.T) {
	assert.Equal(t, normalizeDMARCPolicy("reject"), DMARCPolicyReject)
	assert.Equal(t, normalizeDMARCPolicy("  QUARANTINE "), DMARCPolicyQuarantine)
	assert.Equal(t, normalizeDMARCPolicy("none"), DMARCPolicyNone)
	assert.Equal(t, normalizeDMARCPolicy("blocklist"), "")
	assert.Equal(t, normalizeDMARCPolicy(""), "")
}

func TestDMARCReportingURIs(t *testing.T) {
	got := dmarcReportingURIs("v=DMARC1; p=none; rua=mailto:a@example.com, mailto:b@example.com")

	assert.DeepEqual(t, got, []string{"mailto:a@example.com", "mailto:b@example.com"})
	assert.Assert(t, dmarcReportingURIs("v=DMARC1; p=none") == nil)
}

func TestEmailAuthEnforces(t *testing.T) {
	tests := []struct {
		name string
		auth EmailAuth
		want bool
	}{
		{name: "reject at full percentage enforces", auth: EmailAuth{DMARCPolicy: DMARCPolicyReject, DMARCPercentage: 100}, want: true},
		{name: "quarantine at full percentage enforces", auth: EmailAuth{DMARCPolicy: DMARCPolicyQuarantine, DMARCPercentage: 100}, want: true},
		{name: "reject at partial percentage does not", auth: EmailAuth{DMARCPolicy: DMARCPolicyReject, DMARCPercentage: 10}, want: false},
		{name: "policy none never enforces", auth: EmailAuth{DMARCPolicy: DMARCPolicyNone, DMARCPercentage: 100}, want: false},
		{name: "no policy never enforces", auth: EmailAuth{}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.auth.Enforces(), tt.want)
		})
	}
}

func TestBuildEmailAuth(t *testing.T) {
	t.Run("nil dns enrichment returns nil", func(t *testing.T) {
		assert.Assert(t, buildEmailAuth(Enrichment{}) == nil)
	})

	t.Run("empty dns info returns nil", func(t *testing.T) {
		assert.Assert(t, buildEmailAuth(Enrichment{DNS: &DNSVendorInfo{}}) == nil)
	})

	t.Run("full record set is derived", func(t *testing.T) {
		got := buildEmailAuth(Enrichment{DNS: &DNSVendorInfo{
			SPFRecord:     "v=spf1 include:_spf.google.com ~all",
			DMARCRecord:   "v=DMARC1; p=quarantine; sp=reject; pct=50; rua=mailto:dmarc@example.com",
			DMARCPolicy:   "quarantine",
			DKIMSelectors: []string{"google"},
			MXHosts:       []string{"aspmx.l.google.com"},
		}})

		assert.Assert(t, got != nil)
		assert.Equal(t, got.SPFPolicy, SPFPolicySoftFail)
		assert.Equal(t, got.DMARCPolicy, DMARCPolicyQuarantine)
		assert.Equal(t, got.DMARCSubdomainPolicy, DMARCPolicyReject)
		assert.Equal(t, got.DMARCPercentage, 50)
		assert.DeepEqual(t, got.DMARCReportingURIs, []string{"mailto:dmarc@example.com"})
		assert.Equal(t, got.Enforces(), false)
	})

	t.Run("mx only still produces a section", func(t *testing.T) {
		got := buildEmailAuth(Enrichment{DNS: &DNSVendorInfo{MXHosts: []string{"mx.example.com"}}})

		assert.Assert(t, got != nil)
		assert.Equal(t, got.DMARCPolicy, "")
	})
}
