package ssoutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluate(t *testing.T) {
	t.Parallel()

	exemptDomains := []string{"AuditFirm.com"}

	tests := []struct {
		name         string
		in           EnforcementInput
		wantMustSSO  bool
		wantExempt   bool
		wantReason   string
		wantTFA      bool
		wantEnforced bool
	}{
		{
			name:         "sso not enforced",
			in:           EnforcementInput{SSOEnforced: false, IDPAuthTested: true, IsMember: true, Email: "user@corp.com"},
			wantMustSSO:  false,
			wantExempt:   false,
			wantEnforced: false,
		},
		{
			name:         "enforced but connection not tested is inert",
			in:           EnforcementInput{SSOEnforced: true, IDPAuthTested: false, IsMember: true, Email: "user@corp.com"},
			wantMustSSO:  false,
			wantExempt:   false,
			wantEnforced: false,
		},
		{
			name:         "enforced member must sso",
			in:           EnforcementInput{SSOEnforced: true, IDPAuthTested: true, IsMember: true, Email: "user@corp.com"},
			wantMustSSO:  true,
			wantExempt:   false,
			wantEnforced: true,
		},
		{
			name:         "owner is exempt",
			in:           EnforcementInput{SSOEnforced: true, IDPAuthTested: true, IsMember: true, IsOwner: true, Email: "owner@corp.com"},
			wantMustSSO:  false,
			wantExempt:   true,
			wantReason:   ExemptReasonOwner,
			wantEnforced: true,
		},
		{
			name:         "member with explicit exemption",
			in:           EnforcementInput{SSOEnforced: true, IDPAuthTested: true, IsMember: true, MemberExempt: true, Email: "auditor@corp.com"},
			wantMustSSO:  false,
			wantExempt:   true,
			wantReason:   ExemptReasonUser,
			wantEnforced: true,
		},
		{
			name:         "member whose domain is exempt case insensitive",
			in:           EnforcementInput{SSOEnforced: true, IDPAuthTested: true, IsMember: true, ExemptDomains: exemptDomains, Email: "auditor@auditfirm.com"},
			wantMustSSO:  false,
			wantExempt:   true,
			wantReason:   ExemptReasonDomain,
			wantEnforced: true,
		},
		{
			name:         "non member with exempt domain is not exempt",
			in:           EnforcementInput{SSOEnforced: true, IDPAuthTested: true, IsMember: false, ExemptDomains: exemptDomains, Email: "auditor@auditfirm.com"},
			wantMustSSO:  true,
			wantExempt:   false,
			wantEnforced: true,
		},
		{
			name:         "non exempt domain member must sso",
			in:           EnforcementInput{SSOEnforced: true, IDPAuthTested: true, IsMember: true, ExemptDomains: exemptDomains, Email: "user@corp.com"},
			wantMustSSO:  true,
			wantExempt:   false,
			wantEnforced: true,
		},
		{
			name:         "support session is exempt",
			in:           EnforcementInput{SSOEnforced: true, IDPAuthTested: true, SupportSession: true, Email: "support@system.theopenlane.io"},
			wantMustSSO:  false,
			wantExempt:   true,
			wantReason:   ExemptReasonSupport,
			wantEnforced: true,
		},
		{
			name:         "tfa required regardless of exemption",
			in:           EnforcementInput{SSOEnforced: true, IDPAuthTested: true, TFAEnforced: true, IsMember: true, IsOwner: true, Email: "owner@corp.com"},
			wantMustSSO:  false,
			wantExempt:   true,
			wantReason:   ExemptReasonOwner,
			wantTFA:      true,
			wantEnforced: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Evaluate(tt.in)

			assert.Equal(t, tt.wantMustSSO, got.MustSSO, "MustSSO")
			assert.Equal(t, tt.wantExempt, got.Exempt, "Exempt")
			assert.Equal(t, tt.wantReason, got.ExemptReason, "ExemptReason")
			assert.Equal(t, tt.wantTFA, got.TFARequired, "TFARequired")
			assert.Equal(t, tt.wantEnforced, got.SSOEnforced, "SSOEnforced")
		})
	}
}
