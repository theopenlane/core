package domainscan

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestNormalizeComplianceType(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{value: "privacy_policy", want: "privacy_policy"},
		// punctuation and case are handled mechanically, so none of these needs an alias entry
		{value: "Privacy Policy", want: "privacy_policy"},
		{value: "privacy-policy", want: "privacy_policy"},
		{value: "privacypolicy", want: "privacy_policy"},
		{value: "privacy__policy", want: "privacy_policy"},
		{value: "  PRIVACY-Policy  ", want: "privacy_policy"},
		{value: "TERMS_OF_SERVICE", want: "terms_of_service"},
		{value: "termsofservice", want: "terms_of_service"},
		{value: "sub-processors", want: "subprocessors"},
		// these are different words for the same document, which is what the alias table is for
		{value: "privacy", want: "privacy_policy"},
		{value: "tos", want: "terms_of_service"},
		{value: "terms of use", want: "terms_of_service"},
		{value: "data-processing", want: "dpa"},
		{value: "data processing agreement", want: "dpa"},
		{value: "subprocessor", want: "subprocessors"},
		{value: "trust", want: "trust_center"},
		{value: "cookies", want: "cookie_policy"},
		{value: "", want: ""},
		{value: "   ", want: ""},
		// an unrecognized value is returned normalized rather than dropped, so a caller can
		// still compare it and a future type needs no alias to work
		{value: "Bug Bounty", want: "bug_bounty"},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			assert.Check(t, is.Equal(tt.want, NormalizeComplianceType(tt.value)))
		})
	}
}

func TestIsCanonicalComplianceType(t *testing.T) {
	assert.Check(t, IsCanonicalComplianceType("privacy"))
	assert.Check(t, IsCanonicalComplianceType("Terms of Service"))
	assert.Check(t, !IsCanonicalComplianceType("other"))
	assert.Check(t, !IsCanonicalComplianceType("cookie_settings"))
	assert.Check(t, !IsCanonicalComplianceType(""))
}

func TestMatchesComplianceType(t *testing.T) {
	assert.Check(t, MatchesComplianceType("Privacy Policy", "privacy_policy"))
	assert.Check(t, MatchesComplianceType("tos", "terms_of_service"))
	assert.Check(t, !MatchesComplianceType("", "privacy_policy"))
	assert.Check(t, !MatchesComplianceType("security", "privacy_policy"))
}
