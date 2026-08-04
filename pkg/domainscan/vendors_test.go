package domainscan

import (
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

func TestRegistrableDomainExported(t *testing.T) {
	got, ok := RegistrableDomain("app.hubspot.com")

	assert.Check(t, ok)
	assert.Check(t, is.Equal("hubspot.com", got))
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
	g.add("name:acme", "Acme", "", "Unknown", []string{"CRM"})
	g.add("acme.com", "Acme", "", "https://acme.com", []string{"Analytics"})
	g.add("name:other", "Other", "", "https://other.com", nil)

	got := g.finalize()

	want := []Vendor{
		{Name: "Acme", URL: "https://acme.com", Categories: []string{"Analytics", "CRM"}},
		{Name: "Other", URL: "https://other.com"},
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

	want := []Vendor{
		{Name: "Cloudflare", URL: "https://cloudflare.com", Categories: []string{"Analytics"}},
	}

	assert.Check(t, is.DeepEqual(want, groups.finalize()))
}

func TestFilterRedundantGoogle(t *testing.T) {
	tests := []struct {
		name    string
		vendors []Vendor
		want    []Vendor
	}{
		{
			name: "plain Google is dropped when a specific Google product is present",
			vendors: []Vendor{
				{Name: "Google"},
				{Name: "Google Workspace"},
				{Name: "Acme"},
			},
			want: []Vendor{
				{Name: "Google Workspace"},
				{Name: "Acme"},
			},
		},
		{
			name: "plain Google is kept when no specific Google product is present",
			vendors: []Vendor{
				{Name: "Google"},
				{Name: "Acme"},
			},
			want: []Vendor{
				{Name: "Google"},
				{Name: "Acme"},
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
	vendors := []Vendor{
		{Name: "Acme"},
		{Name: "rfc-editor"},
		{Name: "GitHub"},
	}

	got := filterDeniedVendors(vendors, []string{"RFC-Editor"})

	want := []Vendor{
		{Name: "Acme"},
		{Name: "GitHub"},
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

	want := []Vendor{
		{Name: "Segment", URL: "Unknown"},
		{Name: "GitHub", URL: "https://github.com"},
		{Name: "AWS", URL: "https://aws.amazon.com"},
		{Name: "Cloudflare", URL: "https://cloudflare.com"},
	}

	assert.Check(t, is.DeepEqual(want, got))
}

func TestMergeEnrichmentVendorsCapturesLegalName(t *testing.T) {
	enrichment := Enrichment{
		Compliance: &CompliancePage{
			Subprocessors: []string{"Cloudflare, Inc", "Atlassian Corporation Plc"},
		},
	}

	groups := newVendorGroups()
	mergeEnrichmentVendors(enrichment, groups)

	got := groups.finalize()

	want := []Vendor{
		{Name: "Cloudflare", LegalName: "Cloudflare, Inc", URL: "https://cloudflare.com"},
		{Name: "Atlassian", LegalName: "Atlassian Corporation Plc", URL: "Unknown"},
	}

	assert.Check(t, is.DeepEqual(want, got))
}
