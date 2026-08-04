package domainscan

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestVendorNameFromASNOrg(t *testing.T) {
	tests := []struct {
		name string
		org  string
		want string
	}{
		{name: "strips Inc suffix and canonicalizes", org: "Cloudflare, Inc.", want: "Cloudflare"},
		{name: "strips LLC suffix", org: "Google LLC", want: "Google"},
		{name: "strips Limited suffix", org: "Datacamp Limited", want: "Datacamp"},
		{name: "strips GmbH suffix", org: "Hetzner Online GmbH", want: "Hetzner Online"},
		{name: "strips trailing registry code and canonicalizes", org: "AMAZON-02", want: "AWS"},
		{name: "empty org returns empty", org: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := vendorNameFromASNOrg(tt.org)

			assert.Check(t, is.Equal(tt.want, got))
		})
	}
}

func TestHostVendorName(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		apexDomain string
		pageASNOrg string
		want       string
	}{
		{
			name:       "third-party host matched by domain",
			host:       "cdn.iubenda.com",
			apexDomain: "theopenlane.io",
			pageASNOrg: "Cloudflare, Inc.",
			want:       "Iubenda",
		},
		{
			name:       "apex domain falls back to the page's hosting ASN org",
			host:       "theopenlane.io",
			apexDomain: "theopenlane.io",
			pageASNOrg: "Cloudflare, Inc.",
			want:       "Cloudflare",
		},
		{
			name:       "apex domain with no ASN org returns empty",
			host:       "theopenlane.io",
			apexDomain: "theopenlane.io",
			pageASNOrg: "",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hostVendorName(tt.host, tt.apexDomain, tt.pageASNOrg)

			assert.Check(t, is.Equal(tt.want, got))
		})
	}
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

	got := buildInternalDomains(enrichment, "example.com", "Cloudflare, Inc.")

	// every host is a subdomain of the apex, so each is attributed to whoever fronts the apex
	want := []DNSRecord{
		{Domain: "app.example.com", Type: "internal", Vendor: "Cloudflare"},
		{Domain: "mail.example.com", Type: "internal", Vendor: "Cloudflare"},
		{Domain: "status.example.com", Type: "internal", Vendor: "Cloudflare"},
		{Domain: "trust.example.com", Type: "internal", Vendor: "Cloudflare"},
	}

	assert.Check(t, is.DeepEqual(want, got))
}
