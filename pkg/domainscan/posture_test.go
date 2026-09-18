package domainscan

import (
	"testing"

	"gotest.tools/v3/assert"
)

func postureChecks(findings []PostureFinding) []string {
	checks := make([]string, 0, len(findings))
	for _, f := range findings {
		checks = append(checks, f.Check)
	}

	return checks
}

func TestEmailAuthFindings(t *testing.T) {
	tests := []struct {
		name string
		auth EmailAuth
		want []string
	}{
		{name: "nothing published", auth: EmailAuth{MXHosts: []string{"mx.example.com"}}, want: []string{"dmarc", "spf", "dkim"}},
		{name: "dmarc none", auth: EmailAuth{DMARCRecord: "v=DMARC1; p=none", DMARCPolicy: DMARCPolicyNone, DMARCPercentage: 100, SPFPolicy: SPFPolicySoftFail}, want: []string{"dmarc"}},
		{name: "dmarc sampled", auth: EmailAuth{DMARCRecord: "v=DMARC1; p=reject; pct=10", DMARCPolicy: DMARCPolicyReject, DMARCPercentage: 10, SPFPolicy: SPFPolicyHardFail}, want: []string{"dmarc"}},
		{name: "spf permissive", auth: EmailAuth{DMARCRecord: "v=DMARC1; p=reject", DMARCPolicy: DMARCPolicyReject, DMARCPercentage: 100, SPFPolicy: SPFPolicyPassAll}, want: []string{"spf"}},
		{name: "dkim skipped without mx", auth: EmailAuth{DMARCRecord: "v=DMARC1; p=reject", DMARCPolicy: DMARCPolicyReject, DMARCPercentage: 100, SPFPolicy: SPFPolicyHardFail}, want: []string{}},
		{name: "all good", auth: EmailAuth{DMARCRecord: "v=DMARC1; p=quarantine", DMARCPolicy: DMARCPolicyQuarantine, DMARCPercentage: 100, SPFPolicy: SPFPolicySoftFail, DKIMSelectors: []string{"google"}, MXHosts: []string{"mx"}}, want: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.DeepEqual(t, postureChecks(emailAuthFindings(tt.auth)), tt.want)
		})
	}
}

func TestWellKnownFindings(t *testing.T) {
	tests := []struct {
		name string
		wk   WellKnown
		want []string
	}{
		{name: "nothing probed", wk: WellKnown{}, want: []string{}},
		{name: "everything missing", wk: WellKnown{
			SecurityTxt: &SecurityTxt{},
			Transport:   &TransportSecurity{HTTPSReachable: true},
			LLMs:        &LLMsTxt{},
			Robots:      &RobotsTxt{},
		}, want: []string{"security_txt", "https", "llms_txt", "ai_crawlers"}},
		{name: "redirect without hsts", wk: WellKnown{Transport: &TransportSecurity{HTTPSReachable: true, RedirectsToHTTPS: true}}, want: []string{"https"}},
		{name: "https unreachable is skipped", wk: WellKnown{Transport: &TransportSecurity{}}, want: []string{}},
		{name: "expired security txt", wk: WellKnown{SecurityTxt: &SecurityTxt{Present: true, Expired: true}}, want: []string{"security_txt"}},
		{name: "robots with ai position", wk: WellKnown{Robots: &RobotsTxt{Present: true, NamedAICrawlers: []string{"GPTBot"}}}, want: []string{}},
		{name: "all good", wk: WellKnown{
			SecurityTxt: &SecurityTxt{Present: true},
			Transport:   &TransportSecurity{HTTPSReachable: true, RedirectsToHTTPS: true, HSTS: true},
			LLMs:        &LLMsTxt{Present: true},
			Robots:      &RobotsTxt{Present: true, ContentSignals: map[string]string{"ai-train": "no"}},
		}, want: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.DeepEqual(t, postureChecks(wellKnownFindings(tt.wk)), tt.want)
		})
	}
}

func TestBuildPostureFindingsSkipsUnprobed(t *testing.T) {
	assert.Assert(t, buildPostureFindings(Enrichment{}) == nil)
}

func TestMergePostureFindingsSetsDomain(t *testing.T) {
	m := newFindingsMerge()
	m.add("a.example.com", Findings{Posture: []PostureFinding{{Check: "dmarc"}}})
	m.add("b.example.com", Findings{Posture: []PostureFinding{{Check: "spf"}}})

	got := m.result().Posture

	assert.Equal(t, len(got), 2)
	assert.Equal(t, got[0].Domain, "a.example.com")
	assert.Equal(t, got[1].Domain, "b.example.com")
}
