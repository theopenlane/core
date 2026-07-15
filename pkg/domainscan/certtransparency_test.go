package domainscan

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestParseCertTransparencySubdomains(t *testing.T) {
	tests := []struct {
		name    string
		entries []crtSHEntry
		apex    string
		want    []string
	}{
		{
			name: "extracts and sorts distinct subdomains",
			entries: []crtSHEntry{
				{NameValue: "www.example.com\nexample.com"},
				{NameValue: "api.example.com"},
			},
			apex: "example.com",
			want: []string{"api.example.com", "www.example.com"},
		},
		{
			name: "dedupes repeated entries",
			entries: []crtSHEntry{
				{NameValue: "www.example.com"},
				{NameValue: "www.example.com"},
			},
			apex: "example.com",
			want: []string{"www.example.com"},
		},
		{
			name: "strips wildcard prefix",
			entries: []crtSHEntry{
				{NameValue: "*.staging.example.com"},
			},
			apex: "example.com",
			want: []string{"staging.example.com"},
		},
		{
			name: "excludes the apex itself and unrelated domains",
			entries: []crtSHEntry{
				{NameValue: "example.com\nother.com\nwww.example.com"},
			},
			apex: "example.com",
			want: []string{"www.example.com"},
		},
		{
			name: "excludes names that merely share a suffix with the apex",
			entries: []crtSHEntry{
				{NameValue: "notexample.com"},
			},
			apex: "example.com",
			want: nil,
		},
		{
			name:    "no entries returns nil",
			entries: nil,
			apex:    "example.com",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCertTransparencySubdomains(tt.entries, tt.apex)

			assert.Check(t, is.DeepEqual(tt.want, got))
		})
	}
}

func TestParseCertTransparencySubdomainsCapsResults(t *testing.T) {
	entries := make([]crtSHEntry, 0, maxCertSubdomains+10)
	for i := 0; i < maxCertSubdomains+10; i++ {
		entries = append(entries, crtSHEntry{NameValue: fmt.Sprintf("sub%d.example.com", i)})
	}

	got := parseCertTransparencySubdomains(entries, "example.com")

	assert.Check(t, is.Len(got, maxCertSubdomains))
}
