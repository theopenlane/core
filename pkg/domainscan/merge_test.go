package domainscan

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestMergeReportsVendorsTechnologiesSystemsDedup(t *testing.T) {
	results := []Result{
		{Domain: "example.com", InternalScanID: "scan-1", Status: "completed"},
		{Domain: "example.org", InternalScanID: "scan-2", Status: "completed"},
	}

	reports := []ScanReportInput{
		{
			Domain: "example.com",
			Report: ScanReport{
				Vendors:      []Vendor{{Name: "Cloudflare", URL: "https://cloudflare.com"}},
				Technologies: []Technology{{Name: "Astro"}},
				Systems:      []SystemEntry{{SystemName: "Console"}},
			},
		},
		{
			// same vendor/technology/system by name (different case), should not duplicate
			Domain: "example.org",
			Report: ScanReport{
				Vendors:      []Vendor{{Name: "cloudflare", URL: "https://cloudflare.com"}, {Name: "Stripe"}},
				Technologies: []Technology{{Name: "astro"}, {Name: "React"}},
				Systems:      []SystemEntry{{SystemName: "Console"}, {SystemName: "API"}},
			},
		},
	}

	got := MergeReports(results, reports)

	assert.Assert(t, is.Len(got.Vendors, 2))
	assert.Equal(t, got.Vendors[0].Name, "Cloudflare")
	assert.Equal(t, got.Vendors[1].Name, "Stripe")

	assert.Assert(t, is.Len(got.Technologies, 2))
	assert.Equal(t, got.Technologies[0].Name, "Astro")
	assert.Equal(t, got.Technologies[1].Name, "React")

	assert.Assert(t, is.Len(got.Systems, 2))
	assert.Equal(t, got.Systems[0].SystemName, "Console")
	assert.Equal(t, got.Systems[1].SystemName, "API")
}

func TestMergeReportsAssetsUnion(t *testing.T) {
	results := []Result{{Domain: "example.com", InternalScanID: "scan-1", Status: "completed"}}

	reports := []ScanReportInput{
		{
			Domain: "example.com",
			Report: ScanReport{
				Assets: &Assets{
					DNSRecords: []DNSRecord{
						{Domain: "example.com", Type: "A"}, {Domain: "example.com", Type: "A"},
						{Domain: "b.example.com", Type: "internal"}, {Domain: "a.example.com", Type: "internal"},
					},
					IPAddresses: []IPAddress{{Address: "1.2.3.4"}, {Address: "1.2.3.4"}},
				},
			},
		},
		{
			Domain: "example.com",
			Report: ScanReport{
				Assets: &Assets{
					DNSRecords: []DNSRecord{
						{Domain: "example.com", Type: "NS"},
						{Domain: "a.example.com", Type: "internal"}, {Domain: "c.example.com", Type: "internal"},
					},
					IPAddresses: []IPAddress{{Address: "5.6.7.8"}},
				},
			},
		},
	}

	got := MergeReports(results, reports)

	assert.Assert(t, got.Assets != nil)
	assert.Assert(t, is.Len(got.Assets.DNSRecords, 5))
	assert.Assert(t, is.Len(got.Assets.IPAddresses, 2))
}

func TestMergeReportsFindingsUnion(t *testing.T) {
	results := []Result{{Domain: "example.com", InternalScanID: "scan-1", Status: "completed"}}

	reports := []ScanReportInput{
		{
			Domain: "example.com",
			Report: ScanReport{
				Findings: Findings{
					SecurityViolations: []string{"malware"},
					Risks:              []string{"phishing"},
					IsMalicious:        false,
				},
			},
		},
		{
			Domain: "example.org",
			Report: ScanReport{
				Findings: Findings{
					SecurityViolations:     []string{"malware", "spam"},
					IsMalicious:            true,
					MissingComplianceLinks: "- [ ] privacy_policy",
					AgentReadiness:         []AgentReadinessFinding{{Level: 1, Checklist: "- [ ] no MCP server card"}},
				},
			},
		},
	}

	got := MergeReports(results, reports)

	assert.DeepEqual(t, got.Findings.SecurityViolations, []string{"malware", "spam"})
	assert.DeepEqual(t, got.Findings.Risks, []string{"phishing"})
	assert.Equal(t, got.Findings.IsMalicious, true)
	assert.Equal(t, got.Findings.MissingComplianceLinks, "- [ ] privacy_policy")
	assert.DeepEqual(t, got.Findings.AgentReadiness, []AgentReadinessFinding{
		{Level: 1, Checklist: "- [ ] no MCP server card", Domain: "example.org"},
	})
}

func TestMergeReportsPlatformCompliancMetaFirstNonEmptyWins(t *testing.T) {
	results := []Result{
		{Domain: "example.com", InternalScanID: "scan-1", Status: "failed"},
		{Domain: "example.org", InternalScanID: "scan-2", Status: "completed"},
	}

	reports := []ScanReportInput{
		{
			Domain: "example.com",
			Report: ScanReport{
				Platform:   &Platform{Name: "Openlane"},
				Compliance: &Compliance{IsSOC2: true},
				Meta:       &Meta{Rank: 1},
			},
		},
		{
			Domain: "example.org",
			Report: ScanReport{
				Platform:   &Platform{Name: "should not win"},
				Compliance: &Compliance{IsSOC2: false},
				Meta:       &Meta{Rank: 2},
			},
		},
	}

	got := MergeReports(results, reports)

	assert.Equal(t, got.Platform.Name, "Openlane")
	assert.Equal(t, got.Compliance.IsSOC2, true)
	assert.Equal(t, got.Meta.Rank, 1)
}

func TestMergeReportsPreservesScansIncludingFailed(t *testing.T) {
	results := []Result{
		{Domain: "example.com", InternalScanID: "scan-1", Status: "completed"},
		{Domain: "bad-domain.example", InternalScanID: "scan-2", Status: "failed"},
	}

	got := MergeReports(results, []ScanReportInput{{Domain: "example.com"}})

	assert.DeepEqual(t, got.Scans, results)
}

func TestMergeReportsSingleDomainIsTrivialCase(t *testing.T) {
	results := []Result{{Domain: "example.com", InternalScanID: "scan-1", Status: "completed"}}

	reports := []ScanReportInput{
		{
			Domain: "example.com",
			Report: ScanReport{
				Vendors: []Vendor{{Name: "Vercel"}},
			},
		},
	}

	got := MergeReports(results, reports)

	assert.Assert(t, is.Len(got.Scans, 1))
	assert.Assert(t, is.Len(got.Vendors, 1))
	assert.Equal(t, got.Vendors[0].Name, "Vercel")
}
