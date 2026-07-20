package domainscan

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestMergeReportsVendorsTechnologiesSystemsDedup(t *testing.T) {
	results := []DomainScanResult{
		{Domain: "example.com", InternalScanID: "scan-1", Status: "completed"},
		{Domain: "example.org", InternalScanID: "scan-2", Status: "completed"},
	}

	reports := []map[string]any{
		{
			"vendors":      []map[string]any{{"name": "Cloudflare", "url": "https://cloudflare.com"}},
			"technologies": []map[string]any{{"name": "Astro"}},
			"systems":      []map[string]any{{"system_name": "Console"}},
		},
		{
			// same vendor/technology/system by name (different case), should not duplicate
			"vendors":      []map[string]any{{"name": "cloudflare", "url": "https://cloudflare.com"}, {"name": "Stripe"}},
			"technologies": []map[string]any{{"name": "astro"}, {"name": "React"}},
			"systems":      []map[string]any{{"system_name": "Console"}, {"system_name": "API"}},
		},
	}

	got := MergeReports(results, reports)

	assert.Assert(t, is.Len(got.Vendors, 2))
	assert.Equal(t, got.Vendors[0]["name"], "Cloudflare")
	assert.Equal(t, got.Vendors[1]["name"], "Stripe")

	assert.Assert(t, is.Len(got.Technologies, 2))
	assert.Equal(t, got.Technologies[0]["name"], "Astro")
	assert.Equal(t, got.Technologies[1]["name"], "React")

	assert.Assert(t, is.Len(got.Systems, 2))
	assert.Equal(t, got.Systems[0]["system_name"], "Console")
	assert.Equal(t, got.Systems[1]["system_name"], "API")
}

func TestMergeReportsAssetsUnion(t *testing.T) {
	results := []DomainScanResult{{Domain: "example.com", InternalScanID: "scan-1", Status: "completed"}}

	reports := []map[string]any{
		{
			"assets": map[string]any{
				"dns_records":      []map[string]any{{"domain": "example.com", "type": "A"}, {"domain": "example.com", "type": "A"}},
				"ip_addresses":     []map[string]any{{"address": "1.2.3.4"}, {"address": "1.2.3.4"}},
				"internal_domains": []string{"b.example.com", "a.example.com"},
			},
		},
		{
			"assets": map[string]any{
				"dns_records":      []map[string]any{{"domain": "example.com", "type": "NS"}},
				"ip_addresses":     []map[string]any{{"address": "5.6.7.8"}},
				"internal_domains": []string{"a.example.com", "c.example.com"},
			},
		},
	}

	got := MergeReports(results, reports)

	dnsRecords, ok := got.Assets["dns_records"].([]map[string]any)
	assert.Assert(t, ok)
	assert.Assert(t, is.Len(dnsRecords, 2))

	ipAddresses, ok := got.Assets["ip_addresses"].([]map[string]any)
	assert.Assert(t, ok)
	assert.Assert(t, is.Len(ipAddresses, 2))

	internalDomains, ok := got.Assets["internal_domains"].([]string)
	assert.Assert(t, ok)
	assert.DeepEqual(t, internalDomains, []string{"a.example.com", "b.example.com", "c.example.com"})
}

func TestMergeReportsFindingsUnion(t *testing.T) {
	results := []DomainScanResult{{Domain: "example.com", InternalScanID: "scan-1", Status: "completed"}}

	reports := []map[string]any{
		{
			"findings": map[string]any{
				"security_violations": []string{"malware"},
				"risks":               []string{"phishing"},
				"is_malicious":        false,
			},
		},
		{
			"findings": map[string]any{
				"security_violations":      []string{"malware", "spam"},
				"is_malicious":             true,
				"missing_compliance_links": "- [ ] privacy_policy",
				"agent_readiness":          map[string]any{"level": int64(1), "checklist": "- [ ] no MCP server card"},
			},
		},
	}

	got := MergeReports(results, reports)

	assert.DeepEqual(t, got.Findings["security_violations"], []string{"malware", "spam"})
	assert.DeepEqual(t, got.Findings["risks"], []string{"phishing"})
	assert.Equal(t, got.Findings["is_malicious"], true)
	assert.Equal(t, got.Findings["missing_compliance_links"], "- [ ] privacy_policy")
	assert.DeepEqual(t, got.Findings["agent_readiness"], []map[string]any{
		{"level": int64(1), "checklist": "- [ ] no MCP server card"},
	})
}

func TestMergeReportsPlatformCompliancMetaFirstNonEmptyWins(t *testing.T) {
	results := []DomainScanResult{
		{Domain: "example.com", InternalScanID: "scan-1", Status: "failed"},
		{Domain: "example.org", InternalScanID: "scan-2", Status: "completed"},
	}

	reports := []map[string]any{
		{
			"platform":   map[string]any{"name": "Openlane"},
			"compliance": map[string]any{"is_soc2": true},
			"meta":       map[string]any{"rank": 1},
			"registrar":  map[string]any{"registrar": "Cloudflare, Inc."},
		},
		{
			"platform":   map[string]any{"name": "should not win"},
			"compliance": map[string]any{"is_soc2": false},
			"meta":       map[string]any{"rank": 2},
			"registrar":  map[string]any{"registrar": "should not win"},
		},
	}

	got := MergeReports(results, reports)

	assert.Equal(t, got.Platform["name"], "Openlane")
	assert.Equal(t, got.Compliance["is_soc2"], true)
	assert.Equal(t, got.Meta["rank"], 1)
	assert.Equal(t, got.Registrar["registrar"], "Cloudflare, Inc.")
}

func TestMergeReportsPreservesScansIncludingFailed(t *testing.T) {
	results := []DomainScanResult{
		{Domain: "example.com", InternalScanID: "scan-1", Status: "completed"},
		{Domain: "bad-domain.example", InternalScanID: "scan-2", Status: "failed"},
	}

	got := MergeReports(results, []map[string]any{{}})

	assert.DeepEqual(t, got.Scans, results)
}

func TestMergeReportsSingleDomainIsTrivialCase(t *testing.T) {
	results := []DomainScanResult{{Domain: "example.com", InternalScanID: "scan-1", Status: "completed"}}

	reports := []map[string]any{
		{
			"vendors": []map[string]any{{"name": "Vercel"}},
		},
	}

	got := MergeReports(results, reports)

	assert.Assert(t, is.Len(got.Scans, 1))
	assert.Assert(t, is.Len(got.Vendors, 1))
	assert.Equal(t, got.Vendors[0]["name"], "Vercel")
}
