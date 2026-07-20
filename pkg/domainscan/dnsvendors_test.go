package domainscan

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestVendorNameFromHostname(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		wantName   string
		wantDomain string
	}{
		{
			name:       "exact host override wins",
			host:       "admin.google.com",
			wantName:   "Google Workspace",
			wantDomain: "google.com",
		},
		{
			name:       "domain override wins",
			host:       "hubspot.com",
			wantName:   "Hubspot",
			wantDomain: "hubspot.com",
		},
		{
			name:       "alsoKnownAs alias resolves through the naive label, not just an exact domain override",
			host:       "hubspotemail.net",
			wantName:   "Hubspot",
			wantDomain: "hubspotemail.net",
		},
		{
			name:       "GoDaddy's default nameserver domain resolves via alias",
			host:       "ns1.domaincontrol.com",
			wantName:   "GoDaddy",
			wantDomain: "domaincontrol.com",
		},
		{
			name:       "Namecheap's default nameserver domain resolves via alias",
			host:       "dns1.registrar-servers.com",
			wantName:   "Namecheap",
			wantDomain: "registrar-servers.com",
		},
		{
			name:       "unknown domain falls back to naive title case",
			host:       "mail.example.com",
			wantName:   "Example",
			wantDomain: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotDomain := vendorNameFromHostname(tt.host)

			assert.Check(t, is.Equal(tt.wantName, gotName))
			assert.Check(t, is.Equal(tt.wantDomain, gotDomain))
		})
	}
}

func TestVerificationVendorName(t *testing.T) {
	tests := []struct {
		name   string
		record string
		want   string
		wantOK bool
	}{
		{
			name:   "hyphen-delimited tag",
			record: "google-site-verification=abc123",
			want:   "Google",
			wantOK: true,
		},
		{
			name:   "underscore-delimited tag",
			record: "cloudflare_dashboard_sso=abc123",
			want:   "Cloudflare",
			wantOK: true,
		},
		{
			name:   "no delimiter uses the whole key",
			record: "adobeidverification=abc123",
			want:   "Adobeidverification",
			wantOK: true,
		},
		{
			name:   "mail-protocol tags are excluded",
			record: "v=spf1 include:_spf.example.com ~all",
			want:   "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := verificationVendorName(tt.record)

			assert.Check(t, is.Equal(tt.wantOK, ok))
			assert.Check(t, is.Equal(tt.want, got))
		})
	}
}
