package domainscan

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestMergeTrustCenterIntoCompliancePage(t *testing.T) {
	comp := &CompliancePage{
		URL:             "https://example.com/privacy",
		PageType:        "privacy_policy",
		Title:           "Privacy Policy",
		Summary:         "summary",
		Frameworks:      []string{"GDPR"},
		SOC2Certified:   false,
		Subprocessors:   []string{"AWS"},
		Controls:        []string{"encryption at rest"},
		ComplianceLinks: []ComplianceLink{{URL: "https://example.com/dpa", Type: "dpa"}},
	}

	tc := &TrustCenterPage{
		HostedBy:      "Vanta",
		Frameworks:    []string{"SOC 2", "GDPR"},
		SOC2Certified: true,
		Controls:      []string{"MFA enforced"},
		Documents:     []TrustDocument{{Name: "SOC 2 Type II Report", Public: false}},
		Subprocessors: []string{"GCP"},
	}

	got := mergeTrustCenterIntoCompliancePage(comp, tc, "https://trust.example.com")

	// scalar fields from comp are preserved
	assert.Check(t, is.Equal(comp.URL, got.URL))
	assert.Check(t, is.Equal(comp.PageType, got.PageType))
	assert.Check(t, is.Equal(comp.Title, got.Title))
	assert.Check(t, is.Equal(comp.Summary, got.Summary))

	assert.Check(t, is.DeepEqual([]string{"SOC 2", "GDPR"}, got.Frameworks))
	assert.Check(t, is.Equal(true, got.SOC2Certified))
	assert.Check(t, is.DeepEqual([]string{"GCP", "Vanta", "AWS"}, got.Subprocessors))
	assert.Check(t, is.DeepEqual([]string{"MFA enforced", "encryption at rest"}, got.Controls))
	assert.Check(t, is.Equal("Vanta", got.TrustCenterHostedBy))
	assert.Check(t, is.DeepEqual(tc.Documents, got.Documents))
	assert.Check(t, is.DeepEqual([]ComplianceLink{
		{URL: "https://example.com/dpa", Type: "dpa"},
		{URL: "https://trust.example.com", Type: "trust_center"},
	}, got.ComplianceLinks))

	// original comp is untouched
	assert.Check(t, is.Equal(false, comp.SOC2Certified))
}

func TestMergeTrustCenterIntoCompliancePage_SelfHostedIsNotASubprocessor(t *testing.T) {
	comp := &CompliancePage{}
	tc := &TrustCenterPage{
		HostedBy:      "self-hosted",
		Subprocessors: []string{"AWS"},
	}

	got := mergeTrustCenterIntoCompliancePage(comp, tc, "https://trust.example.com")

	assert.Check(t, is.DeepEqual([]string{"AWS"}, got.Subprocessors))
}

func TestMergeTrustCenterPages(t *testing.T) {
	t.Run("nil pages are skipped", func(t *testing.T) {
		got := mergeTrustCenterPages(nil, nil)

		assert.Check(t, is.Equal("", got.HostedBy))
		assert.Check(t, is.Len(got.Frameworks, 0))
	})

	t.Run("unions list fields across pages", func(t *testing.T) {
		a := &TrustCenterPage{
			HostedBy:      "self-hosted",
			Frameworks:    []string{"SOC 2"},
			SOC2Certified: false,
			Controls:      []string{"MFA enforced"},
			Documents:     []TrustDocument{{Name: "SOC 2 Report"}},
			Subprocessors: []string{"AWS"},
		}
		b := &TrustCenterPage{
			HostedBy:      "Vanta",
			Frameworks:    []string{"SOC 2", "ISO 27001"},
			SOC2Certified: true,
			Controls:      []string{"encryption at rest"},
			Documents:     []TrustDocument{{Name: "SOC 2 Report"}, {Name: "ISO Cert"}},
			Subprocessors: []string{"GCP"},
		}

		got := mergeTrustCenterPages(a, b, nil)

		// first non-empty, non-"self-hosted" HostedBy wins
		assert.Check(t, is.Equal("Vanta", got.HostedBy))
		assert.Check(t, is.DeepEqual([]string{"SOC 2", "ISO 27001"}, got.Frameworks))
		assert.Check(t, is.Equal(true, got.SOC2Certified))
		assert.Check(t, is.DeepEqual([]string{"MFA enforced", "encryption at rest"}, got.Controls))
		assert.Check(t, is.DeepEqual([]TrustDocument{{Name: "SOC 2 Report"}, {Name: "ISO Cert"}}, got.Documents))
		assert.Check(t, is.DeepEqual([]string{"AWS", "GCP"}, got.Subprocessors))
	})

	t.Run("first non-self-hosted HostedBy is kept even if seen later", func(t *testing.T) {
		a := &TrustCenterPage{HostedBy: "self-hosted"}
		b := &TrustCenterPage{HostedBy: "Drata"}
		c := &TrustCenterPage{HostedBy: "Vanta"}

		got := mergeTrustCenterPages(a, b, c)

		assert.Check(t, is.Equal("Drata", got.HostedBy))
	})
}

func TestMergeTrustDocuments(t *testing.T) {
	tests := []struct {
		name string
		a    []TrustDocument
		b    []TrustDocument
		want []TrustDocument
	}{
		{
			name: "unions and dedupes by name, preserving first-seen order",
			a:    []TrustDocument{{Name: "SOC 2 Report", Public: true}},
			b:    []TrustDocument{{Name: "SOC 2 Report", Public: false}, {Name: "ISO Cert", Public: true}},
			want: []TrustDocument{{Name: "SOC 2 Report", Public: true}, {Name: "ISO Cert", Public: true}},
		},
		{
			name: "documents with empty names are dropped",
			a:    []TrustDocument{{Name: ""}},
			b:    []TrustDocument{{Name: "ISO Cert"}},
			want: []TrustDocument{{Name: "ISO Cert"}},
		},
		{
			name: "both empty returns empty slice",
			a:    nil,
			b:    nil,
			want: []TrustDocument{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeTrustDocuments(tt.a, tt.b)

			assert.Check(t, is.DeepEqual(tt.want, got))
		})
	}
}

func TestMergeComplianceLinks(t *testing.T) {
	tests := []struct {
		name string
		a    []ComplianceLink
		b    []ComplianceLink
		want []ComplianceLink
	}{
		{
			name: "unions and dedupes by url preserving first-seen order",
			a:    []ComplianceLink{{URL: "https://example.com/privacy", Type: "privacy_policy"}},
			b: []ComplianceLink{
				{URL: "https://example.com/privacy", Type: "privacy_policy_dup"},
				{URL: "https://example.com/dpa", Type: "dpa"},
			},
			want: []ComplianceLink{
				{URL: "https://example.com/privacy", Type: "privacy_policy"},
				{URL: "https://example.com/dpa", Type: "dpa"},
			},
		},
		{
			name: "links with empty url are dropped",
			a:    []ComplianceLink{{URL: "", Type: "other"}},
			b:    []ComplianceLink{{URL: "https://example.com/dpa", Type: "dpa"}},
			want: []ComplianceLink{{URL: "https://example.com/dpa", Type: "dpa"}},
		},
		{
			name: "both empty returns empty slice",
			a:    nil,
			b:    nil,
			want: []ComplianceLink{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeComplianceLinks(tt.a, tt.b)

			assert.Check(t, is.DeepEqual(tt.want, got))
		})
	}
}
