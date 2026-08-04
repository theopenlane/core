package domainscan

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/pkg/jsonx"
)

func TestBuildScanReportNilResult(t *testing.T) {
	enrichment := Enrichment{
		Company: &CompanyProfile{
			Name: "Acme",
			Systems: []System{
				{Name: "Console", Summary: "The web console"},
			},
		},
		Compliance: &CompliancePage{
			Frameworks: []string{"SOC 2"},
		},
		DNS: &DNSVendorInfo{
			Vendors: []DNSVendor{
				{Name: "Cloudflare", URL: "https://cloudflare.com"},
			},
		},
	}

	data := BuildScanReport(nil, enrichment, nil, nil)

	// URL-scanner-only sections must be absent when result is nil, not merely empty
	_, hasExternalScanID := data["external_scan_id"]
	assert.Check(t, !hasExternalScanID)

	_, hasURL := data["url"]
	assert.Check(t, !hasURL)

	_, hasMeta := data["meta"]
	assert.Check(t, !hasMeta)

	_, hasBranding := data["branding"]
	assert.Check(t, !hasBranding)

	var report ScanReport
	assert.NilError(t, jsonx.RoundTrip(data, &report))

	// enrichment-derived sections should still be populated
	assert.Assert(t, report.Platform != nil)
	assert.Check(t, is.Equal("Acme", report.Platform.Name))

	assert.Assert(t, is.Len(report.Systems, 1))
	assert.Check(t, is.Equal("Console", report.Systems[0].SystemName))

	assert.Assert(t, report.Compliance != nil)
	assert.Check(t, is.DeepEqual([]string{"SOC 2"}, report.Compliance.Frameworks))

	assert.Assert(t, is.Len(report.Vendors, 1))
	assert.Check(t, is.Equal("Cloudflare", report.Vendors[0].Name))

	assert.Check(t, is.Len(report.Findings.SecurityViolations, 0))
}
