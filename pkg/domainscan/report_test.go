package domainscan

import (
	"sort"
	"testing"

	"github.com/cloudflare/cloudflare-go/v7/url_scanner"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestRegistrableDomain(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		want   string
	}{
		{
			name:   "empty url returns empty",
			rawURL: "",
			want:   "",
		},
		{
			name:   "url with no host returns empty",
			rawURL: "https://",
			want:   "",
		},
		{
			name:   "subdomain collapses to registrable domain",
			rawURL: "https://www.example.com/path",
			want:   "example.com",
		},
		{
			name:   "private suffix keeps the vendor's brand domain",
			rawURL: "https://d3e54v103j8qbb.cloudfront.net",
			want:   "cloudfront.net",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registrableDomain(tt.rawURL)

			assert.Check(t, is.Equal(tt.want, got))
		})
	}
}

func TestIcannRegistrableDomain(t *testing.T) {
	tests := []struct {
		name   string
		host   string
		want   string
		wantOK bool
	}{
		{
			name:   "icann suffix uses effective tld plus one",
			host:   "www.google.com",
			want:   "google.com",
			wantOK: true,
		},
		{
			name:   "private suffix returns the suffix itself",
			host:   "myapp.herokuapp.com",
			want:   "herokuapp.com",
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := icannRegistrableDomain(tt.host)

			assert.Check(t, is.Equal(tt.wantOK, ok))
			assert.Check(t, is.Equal(tt.want, got))
		})
	}
}

func TestDomainVendorName(t *testing.T) {
	tests := []struct {
		name   string
		domain string
		want   string
	}{
		{name: "capitalizes the first label", domain: "google.com", want: "Google"},
		{name: "single label domain", domain: "localhost", want: "Localhost"},
		{name: "empty domain returns empty", domain: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domainVendorName(tt.domain)

			assert.Check(t, is.Equal(tt.want, got))
		})
	}
}

func TestVendorGroupsAddAndFinalize(t *testing.T) {
	g := newVendorGroups()

	// two signals for the same vendor arrive under different keys (a domain key and a
	// name key) and should merge into a single finalized entry
	g.add("name:acme", "Acme", "Unknown", []string{"CRM"})
	g.add("acme.com", "Acme", "https://acme.com", []string{"Analytics"})
	g.add("name:other", "Other", "https://other.com", nil)

	got := g.finalize()

	want := []map[string]any{
		{"name": "Acme", "url": "https://acme.com", "categories": []string{"Analytics", "CRM"}},
		{"name": "Other", "url": "https://other.com"},
	}

	assert.Check(t, is.DeepEqual(want, got))
}

func TestGroupWappaDetectionsCanonicalizesVendorName(t *testing.T) {
	// Wappalyzer reports this technology under its own app name, but it should
	// collapse into the "Cloudflare" alias group rather than surface as a
	// separate vendor sharing the same domain
	wappaData := []url_scanner.ScanGetResponseMetaProcessorsWappaData{
		{
			App:     "Cloudflare Browser Insights",
			Website: "https://cloudflare.com",
			Categories: []url_scanner.ScanGetResponseMetaProcessorsWappaDataCategory{
				{Name: "Analytics"},
			},
		},
	}

	groups, technologies := groupWappaDetections(wappaData, map[string]bool{}, map[string]bool{})

	assert.Check(t, is.Len(technologies, 0))

	want := []map[string]any{
		{"name": "Cloudflare", "url": "https://cloudflare.com", "categories": []string{"Analytics"}},
	}

	assert.Check(t, is.DeepEqual(want, groups.finalize()))
}

func TestFilterRedundantGoogle(t *testing.T) {
	tests := []struct {
		name    string
		vendors []map[string]any
		want    []map[string]any
	}{
		{
			name: "plain Google is dropped when a specific Google product is present",
			vendors: []map[string]any{
				{"name": "Google"},
				{"name": "Google Workspace"},
				{"name": "Acme"},
			},
			want: []map[string]any{
				{"name": "Google Workspace"},
				{"name": "Acme"},
			},
		},
		{
			name: "plain Google is kept when no specific Google product is present",
			vendors: []map[string]any{
				{"name": "Google"},
				{"name": "Acme"},
			},
			want: []map[string]any{
				{"name": "Google"},
				{"name": "Acme"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterRedundantGoogle(tt.vendors)

			assert.Check(t, is.DeepEqual(tt.want, got))
		})
	}
}

func TestFilterDeniedVendors(t *testing.T) {
	vendors := []map[string]any{
		{"name": "Acme"},
		{"name": "rfc-editor"},
		{"name": "GitHub"},
	}

	got := filterDeniedVendors(vendors, []string{"RFC-Editor"})

	want := []map[string]any{
		{"name": "Acme"},
		{"name": "GitHub"},
	}

	assert.Check(t, is.DeepEqual(want, got))
}

func TestMergeEnrichmentVendors(t *testing.T) {
	enrichment := Enrichment{
		Company: &CompanyProfile{
			Technologies: []string{"Segment"},
			SocialLinks:  SocialLinks{GitHub: "https://github.com/acme"},
		},
		Compliance: &CompliancePage{
			Subprocessors: []string{"AWS"},
		},
		DNS: &DNSVendorInfo{
			Vendors: []DNSVendor{{Name: "Cloudflare", URL: "https://cloudflare.com"}},
		},
	}

	groups := newVendorGroups()
	mergeEnrichmentVendors(enrichment, groups)

	got := groups.finalize()

	want := []map[string]any{
		{"name": "Segment", "url": "Unknown"},
		{"name": "GitHub", "url": "https://github.com"},
		{"name": "AWS", "url": "https://aws.amazon.com"},
		{"name": "Cloudflare", "url": "https://cloudflare.com"},
	}

	assert.Check(t, is.DeepEqual(want, got))
}

func TestBuildInternalDomains(t *testing.T) {
	enrichment := Enrichment{
		Company: &CompanyProfile{
			StatusPageURL:  "https://status.example.com",
			SubdomainLinks: []string{"https://APP.example.com/dashboard", "https://"},
		},
		Compliance: &CompliancePage{
			ComplianceLinks: []ComplianceLink{{URL: "https://trust.example.com/page", Type: "trust_center"}},
		},
		DNS: &DNSVendorInfo{
			Subdomains: []SubdomainDNSInfo{{Host: "mail.example.com"}},
		},
	}

	got := buildInternalDomains(enrichment)

	want := []string{"app.example.com", "mail.example.com", "status.example.com", "trust.example.com"}

	assert.Check(t, is.DeepEqual(want, got))
}

func TestBuildMissingComplianceLinks(t *testing.T) {
	tests := []struct {
		name       string
		compliance *CompliancePage
		want       []string
	}{
		{
			name:       "nil compliance returns nil",
			compliance: nil,
			want:       nil,
		},
		{
			name: "found links and page type are excluded from missing",
			compliance: &CompliancePage{
				PageType: "privacy_policy",
				ComplianceLinks: []ComplianceLink{
					{URL: "https://example.com/terms", Type: "terms_of_service"},
				},
			},
			want: []string{"trust_center", "dpa", "security", "cookie_policy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildMissingComplianceLinks(Enrichment{Compliance: tt.compliance})

			assert.Check(t, is.DeepEqual(tt.want, got))
		})
	}
}

func TestBuildPlatform(t *testing.T) {
	t.Run("nil company returns nil", func(t *testing.T) {
		got := buildPlatform(Enrichment{})

		assert.Check(t, got == nil)
	})

	t.Run("includes optional fields only when present", func(t *testing.T) {
		enrichment := Enrichment{
			Company: &CompanyProfile{
				Name:             "Acme",
				Description:      "Widgets",
				Industry:         "Manufacturing",
				Location:         "Remote",
				EmployeeRange:    "11-50",
				FoundedYear:      "2020",
				EstimatedRevenue: "$1M-$10M",
				SSOSupported:     true,
				SocialLinks:      SocialLinks{GitHub: "https://github.com/acme"},
				StatusPageURL:    "https://status.acme.com",
				Customers:        []string{"Globex"},
			},
		}

		got := buildPlatform(enrichment)

		want := map[string]any{
			"name":              "Acme",
			"description":       "Widgets",
			"industry":          "Manufacturing",
			"location":          "Remote",
			"employee_range":    "11-50",
			"founded_year":      "2020",
			"estimated_revenue": "$1M-$10M",
			"sso_supported":     true,
			"mfa_supported":     false,
			"social_links":      SocialLinks{GitHub: "https://github.com/acme"},
			"status_page_url":   "https://status.acme.com",
			"customers":         []string{"Globex"},
		}

		assert.Check(t, is.DeepEqual(want, got))
	})
}

func TestBuildSystems(t *testing.T) {
	t.Run("nil company returns nil", func(t *testing.T) {
		got := buildSystems(Enrichment{})

		assert.Check(t, got == nil)
	})

	t.Run("prefers full description, falls back to summary", func(t *testing.T) {
		enrichment := Enrichment{
			Company: &CompanyProfile{
				Systems: []System{
					{Name: "Widget", Summary: "short", FullDescription: "long"},
					{Name: "Gadget", Summary: "short only"},
				},
			},
		}

		got := buildSystems(enrichment)

		want := []map[string]any{
			{"system_name": "Widget", "description": "long"},
			{"system_name": "Gadget", "description": "short only"},
		}

		assert.Check(t, is.DeepEqual(want, got))
	})
}

func TestBuildComplianceSection(t *testing.T) {
	t.Run("nil compliance returns nil", func(t *testing.T) {
		got := buildComplianceSection(Enrichment{})

		assert.Check(t, got == nil)
	})

	t.Run("includes optional fields only when present", func(t *testing.T) {
		enrichment := Enrichment{
			Compliance: &CompliancePage{
				Frameworks:          []string{"SOC 2"},
				SOC2Certified:       true,
				Controls:            []string{"MFA enforced"},
				TrustCenterHostedBy: "Vanta",
				Documents:           []TrustDocument{{Name: "SOC 2 Report"}},
			},
		}

		got := buildComplianceSection(enrichment)

		want := map[string]any{
			"frameworks":             []string{"SOC 2"},
			"is_soc2":                true,
			"controls":               []string{"MFA enforced"},
			"trust_center_hosted_by": "Vanta",
			"documents":              []TrustDocument{{Name: "SOC 2 Report"}},
		}

		assert.Check(t, is.DeepEqual(want, got))
	})
}

func TestBuildAgentReadinessChecklistMarkdown(t *testing.T) {
	failedChecks := []map[string]any{
		{"check": "markdown", "message": "missing markdown negotiation"},
		{"check": "mcp", "message": "no MCP server card"},
	}

	got := buildAgentReadinessChecklistMarkdown(failedChecks)

	want := "- [ ] missing markdown negotiation\n- [ ] no MCP server card"

	assert.Check(t, is.Equal(want, got))
}

func TestWalkAgentReadinessChecks(t *testing.T) {
	node := map[string]any{
		"markdown": map[string]any{
			"status":  "fail",
			"message": "missing markdown negotiation",
		},
		"mcp": map[string]any{
			"status":  "pass",
			"message": "ok",
		},
		"nested": map[string]any{
			"deep": map[string]any{
				"status":  "fail",
				"message": "deep failure",
			},
		},
	}

	var failedChecks []map[string]any

	walkAgentReadinessChecks(node, "", &failedChecks)

	sort.Slice(failedChecks, func(i, j int) bool {
		return failedChecks[i]["check"].(string) < failedChecks[j]["check"].(string)
	})

	want := []map[string]any{
		{"check": "markdown", "message": "missing markdown negotiation"},
		{"check": "nested.deep", "message": "deep failure"},
	}

	assert.Check(t, is.DeepEqual(want, failedChecks))
}
