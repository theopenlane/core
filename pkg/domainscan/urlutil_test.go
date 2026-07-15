package domainscan

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		want   string
		wantOK bool
	}{
		{
			name:   "bare host gets https scheme",
			rawURL: "example.com",
			want:   "https://example.com",
			wantOK: true,
		},
		{
			name:   "scheme already present is preserved",
			rawURL: "http://example.com",
			want:   "http://example.com",
			wantOK: true,
		},
		{
			name:   "path and query are preserved",
			rawURL: "example.com/path?x=1",
			want:   "https://example.com/path?x=1",
			wantOK: true,
		},
		{
			name:   "empty string has no hostname",
			rawURL: "",
			wantOK: false,
		},
		{
			name:   "scheme with no host fails",
			rawURL: "https://",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := normalizeURL(tt.rawURL)

			assert.Check(t, is.Equal(tt.wantOK, ok))

			if tt.wantOK {
				assert.Check(t, is.Equal(tt.want, got.String()))
			}
		})
	}
}

func TestApexDomain(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		want   string
		wantOK bool
	}{
		{
			name:   "subdomain collapses to registrable domain",
			rawURL: "www.mail.example.co.uk",
			want:   "example.co.uk",
			wantOK: true,
		},
		{
			name:   "bare domain is its own apex",
			rawURL: "example.com",
			want:   "example.com",
			wantOK: true,
		},
		{
			name:   "with scheme and path",
			rawURL: "https://app.example.com/login",
			want:   "example.com",
			wantOK: true,
		},
		{
			name:   "invalid url fails",
			rawURL: "https://",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := apexDomain(tt.rawURL)

			assert.Check(t, is.Equal(tt.wantOK, ok))
			assert.Check(t, is.Equal(tt.want, got))
		})
	}
}

func TestTrustCenterURLs(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		want   []string
		wantOK bool
	}{
		{
			name:   "derives one candidate per subdomain prefix",
			rawURL: "example.com",
			want: []string{
				"https://trust.example.com",
				"https://security.example.com",
				"https://compliance.example.com",
			},
			wantOK: true,
		},
		{
			name:   "path and query are stripped",
			rawURL: "https://www.example.com/some/path?x=1",
			want: []string{
				"https://trust.example.com",
				"https://security.example.com",
				"https://compliance.example.com",
			},
			wantOK: true,
		},
		{
			name:   "invalid url fails",
			rawURL: "https://",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := trustCenterURLs(tt.rawURL)

			assert.Check(t, is.Equal(tt.wantOK, ok))
			assert.Check(t, is.DeepEqual(tt.want, got))
		})
	}
}

func TestStatusPageURL(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		want   string
		wantOK bool
	}{
		{
			name:   "derives status subdomain of apex",
			rawURL: "example.com",
			want:   "https://status.example.com",
			wantOK: true,
		},
		{
			name:   "path and query are stripped",
			rawURL: "https://www.example.com/some/path?x=1",
			want:   "https://status.example.com",
			wantOK: true,
		},
		{
			name:   "invalid url fails",
			rawURL: "https://",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := statusPageURL(tt.rawURL)

			assert.Check(t, is.Equal(tt.wantOK, ok))
			assert.Check(t, is.Equal(tt.want, got))
		})
	}
}
