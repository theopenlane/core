package domainscan

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestMergeStrings(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want []string
	}{
		{
			name: "unions and dedupes preserving first-seen order",
			a:    []string{"SOC 2", "GDPR"},
			b:    []string{"GDPR", "ISO 27001"},
			want: []string{"SOC 2", "GDPR", "ISO 27001"},
		},
		{
			name: "empty values are dropped",
			a:    []string{"", "SOC 2"},
			b:    []string{""},
			want: []string{"SOC 2"},
		},
		{
			name: "both empty returns empty slice",
			a:    nil,
			b:    nil,
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeStrings(tt.a, tt.b)

			assert.Check(t, is.DeepEqual(tt.want, got))
		})
	}
}

func TestMergeCompanyProfiles(t *testing.T) {
	homepage := &CompanyProfile{
		Name:         "Acme",
		Description:  "Acme does things",
		SSOSupported: false,
		MFASupported: false,
		Systems:      []System{{Name: "Console", Summary: "The web console"}},
		Customers:    []string{"Globex"},
		SocialLinks:  SocialLinks{LinkedIn: "https://linkedin.com/company/acme"},
	}

	pricing := &CompanyProfile{
		Name:                 "Acme should not override",
		SSOSupported:         true,
		MFASupported:         true,
		SocialLoginSupported: true,
		CredentialsSupported: true,
		PasskeySupported:     true,
		Systems:              []System{{Name: "Console", Summary: "duplicate, should be dropped"}, {Name: "API", Summary: "The public API"}},
		Customers:            []string{"Initech"},
		SocialLinks:          SocialLinks{LinkedIn: "https://linkedin.com/company/other", Twitter: "https://twitter.com/acme"},
	}

	got := mergeCompanyProfiles(homepage, nil, pricing)

	assert.Check(t, is.Equal("Acme", got.Name), "first non-empty value wins")
	assert.Check(t, is.Equal("Acme does things", got.Description))
	assert.Check(t, is.Equal(true, got.SSOSupported), "sso true on any page should carry through")
	assert.Check(t, is.Equal(true, got.MFASupported))
	assert.Check(t, is.Equal(true, got.SocialLoginSupported))
	assert.Check(t, is.Equal(true, got.CredentialsSupported))
	assert.Check(t, is.Equal(true, got.PasskeySupported))
	assert.Check(t, is.DeepEqual([]System{{Name: "Console", Summary: "The web console"}, {Name: "API", Summary: "The public API"}}, got.Systems))
	assert.Check(t, is.DeepEqual([]string{"Globex", "Initech"}, got.Customers))
	assert.Check(t, is.Equal("https://linkedin.com/company/acme", got.SocialLinks.LinkedIn), "existing social link should not be overwritten")
	assert.Check(t, is.Equal("https://twitter.com/acme", got.SocialLinks.Twitter), "empty social link should be filled from a later page")
}

func TestMergeCompanyProfilesAllNil(t *testing.T) {
	got := mergeCompanyProfiles(nil, nil)

	assert.Check(t, is.Equal("", got.Name))
	assert.Check(t, is.Equal(false, got.SSOSupported))
	assert.Check(t, is.Len(got.Systems, 0))
}
