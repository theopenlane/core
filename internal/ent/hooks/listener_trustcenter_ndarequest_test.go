package hooks

import (
	"testing"

	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/validator"
)

const ndaApprovalTestDomain = "theopenlane.io"

func TestValidateEmailRestrictions(t *testing.T) {
	client := &generated.Client{
		EmailVerifier: &validator.EmailVerifier{
			Client: emailverifier.NewVerifier(),
		},
	}

	tests := []struct {
		name    string
		email   string
		setting models.TrustCenterNDARequestSetting
		blocked bool
	}{
		{
			name:    "invalid email",
			blocked: true,
			setting: models.TrustCenterNDARequestSetting{
				AllowDisposableEmail: true,
				AllowRoleAccount:     true,
			},
		},
		{
			name:  "work email cannot bypass role restriction",
			email: "support@stripe.com",
			setting: models.TrustCenterNDARequestSetting{
				WorkEmailOnly: true,
			},
			blocked: true,
		},
		{
			name:  "disposable rules cannot bypass work email restriction",
			email: "signer@mailinator.com",
			setting: models.TrustCenterNDARequestSetting{
				AllowDisposableEmail: true,
				WorkEmailOnly:        true,
			},
			blocked: true,
		},
		{
			name:    "disposable email is blocked when disposable emails are not allowed",
			email:   "signer@mailinator.com",
			blocked: true,
		},
		{
			name:  "disposable emails allowed",
			email: "signer@mailinator.com",
			setting: models.TrustCenterNDARequestSetting{
				AllowDisposableEmail: true,
			},
		},
		{
			name:  "disposable email blocked by work email only",
			email: "signer@mailinator.com",
			setting: models.TrustCenterNDARequestSetting{
				WorkEmailOnly: true,
			},
			blocked: true,
		},
		{
			name:  "role account at a disposable domain is blocked",
			email: "support@mailinator.com",
			setting: models.TrustCenterNDARequestSetting{
				AllowRoleAccount: true,
			},
			blocked: true,
		},
		{
			name:  "role account at a disposable domain is allowed when both types are allowed",
			email: "support@mailinator.com",
			setting: models.TrustCenterNDARequestSetting{
				AllowRoleAccount:     true,
				AllowDisposableEmail: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &generated.TrustCenterNDARequest{Email: tt.email}

			blocked := validateEmailRules(t.Context(), client, tt.setting, req)
			assert.Equal(t, tt.blocked, blocked)
		})
	}
}

func TestValidateDomainList(t *testing.T) {
	tests := []struct {
		name     string
		setting  models.TrustCenterNDARequestSetting
		approved bool
		matched  bool
	}{
		{
			name: "lists disabled",
			setting: models.TrustCenterNDARequestSetting{
				DomainAllowlist: []string{ndaApprovalTestDomain},
				DomainBlocklist: []string{ndaApprovalTestDomain},
			},
		},
		{
			name: "blocked",
			setting: models.TrustCenterNDARequestSetting{
				UseDomainBlocklist: true,
				DomainBlocklist:    []string{ndaApprovalTestDomain},
			},
			matched: true,
		},
		{
			name: "allowed",
			setting: models.TrustCenterNDARequestSetting{
				UseDomainAllowlist: true,
				DomainAllowlist:    []string{ndaApprovalTestDomain},
			},
			approved: true,
			matched:  true,
		},
		{
			name: "blocklist wins",
			setting: models.TrustCenterNDARequestSetting{
				UseDomainBlocklist: true,
				UseDomainAllowlist: true,
				DomainBlocklist:    []string{ndaApprovalTestDomain},
				DomainAllowlist:    []string{ndaApprovalTestDomain},
			},
			matched: true,
		},
		{
			name: "allowlist miss continues",
			setting: models.TrustCenterNDARequestSetting{
				UseDomainAllowlist: true,
				DomainAllowlist:    []string{"example.com"},
			},
		},
		{
			name: "blocklist miss continues",
			setting: models.TrustCenterNDARequestSetting{
				UseDomainBlocklist: true,
				DomainBlocklist:    []string{"example.com"},
			},
		},
		{
			name: "empty allowlist continues",
			setting: models.TrustCenterNDARequestSetting{
				UseDomainAllowlist: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			approved, matched := validateDomainList(ndaApprovalTestDomain, tt.setting)
			assert.Equal(t, tt.approved, approved)
			assert.Equal(t, tt.matched, matched)
		})
	}
}

func TestValidateContact(t *testing.T) {
	approved, err := validateContact(t.Context(), nil, "signer@"+ndaApprovalTestDomain, models.TrustCenterNDARequestSetting{})
	require.NoError(t, err)
	assert.False(t, approved)
}

func TestValidateExistingDomain(t *testing.T) {
	request := &generated.TrustCenterNDARequest{
		ID:            "request",
		TrustCenterID: "trust-center",
	}
	approved, err := validateDomainFromNDARequest(t.Context(), nil, request, ndaApprovalTestDomain, models.TrustCenterNDARequestSetting{})
	require.NoError(t, err)
	assert.False(t, approved)
}

func TestValidateContactDomain(t *testing.T) {
	approved, err := validateContactDomain(t.Context(), nil, ndaApprovalTestDomain, models.TrustCenterNDARequestSetting{})
	require.NoError(t, err)
	assert.False(t, approved)
}
