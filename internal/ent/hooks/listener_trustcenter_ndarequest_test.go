package hooks

import (
	"testing"

	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/core/v2/internal/ent/generated"
)

const ndaApprovalTestDomain = "theopenlane.io"

func TestValidateEmailRestrictions(t *testing.T) {
	tests := []struct {
		name       string
		valid      bool
		disposable bool
		role       bool
		free       bool
		setting    models.TrustCenterNDARequestSetting
		approved   bool
	}{
		{
			name: "invalid syntax",
			setting: models.TrustCenterNDARequestSetting{
				AllowDisposableEmail: true,
				AllowRoleAccount:     true,
			},
		},
		{
			name:  "work email",
			valid: true,
			setting: models.TrustCenterNDARequestSetting{
				WorkEmailOnly: true,
			},
			approved: true,
		},
		{
			name:       "disposable rejected",
			valid:      true,
			disposable: true,
		},
		{
			name:       "disposable allowed",
			valid:      true,
			disposable: true,
			setting: models.TrustCenterNDARequestSetting{
				AllowDisposableEmail: true,
			},
			approved: true,
		},
		{
			name:       "work only rejects allowed disposable",
			valid:      true,
			disposable: true,
			setting: models.TrustCenterNDARequestSetting{
				AllowDisposableEmail: true,
				WorkEmailOnly:        true,
			},
		},
		{
			name:  "role rejected",
			valid: true,
			role:  true,
		},
		{
			name:  "role allowed",
			valid: true,
			role:  true,
			setting: models.TrustCenterNDARequestSetting{
				AllowRoleAccount: true,
			},
			approved: true,
		},
		{
			name:     "free allowed",
			valid:    true,
			free:     true,
			approved: true,
		},
		{
			name:  "work only rejects free",
			valid: true,
			free:  true,
			setting: models.TrustCenterNDARequestSetting{
				WorkEmailOnly: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &emailverifier.Result{
				Syntax:      emailverifier.Syntax{Valid: tt.valid},
				Disposable:  tt.disposable,
				RoleAccount: tt.role,
				Free:        tt.free,
			}
			assert.Equal(t, tt.approved, validateEmailRules(result, &tt.setting))
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
			approved, matched := validateDomainList(ndaApprovalTestDomain, &tt.setting)
			assert.Equal(t, tt.approved, approved)
			assert.Equal(t, tt.matched, matched)
		})
	}
}

func TestValidateContact(t *testing.T) {
	approved, err := validateContact(t.Context(), nil, "signer@"+ndaApprovalTestDomain, &models.TrustCenterNDARequestSetting{})
	require.NoError(t, err)
	assert.False(t, approved)
}

func TestValidateExistingDomain(t *testing.T) {
	request := &generated.TrustCenterNDARequest{
		ID:            "request",
		TrustCenterID: "trust-center",
	}
	approved, err := validateDomainFromNDARequest(t.Context(), nil, request, ndaApprovalTestDomain, &models.TrustCenterNDARequestSetting{})
	require.NoError(t, err)
	assert.False(t, approved)
}

func TestValidateContactDomain(t *testing.T) {
	approved, err := validateContactDomain(t.Context(), nil, ndaApprovalTestDomain, &models.TrustCenterNDARequestSetting{})
	require.NoError(t, err)
	assert.False(t, approved)
}

func TestValidateEmailRestrictionsWithoutVerifier(t *testing.T) {
	tests := []struct {
		name     string
		setting  models.TrustCenterNDARequestSetting
		approved bool
	}{
		{
			name: "no classification restrictions",
			setting: models.TrustCenterNDARequestSetting{
				AllowDisposableEmail: true,
				AllowRoleAccount:     true,
			},
			approved: true,
		},
		{
			name: "work email requires classification",
			setting: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:        true,
				AllowDisposableEmail: true,
				AllowRoleAccount:     true,
			},
		},
		{
			name: "disposable restriction requires classification",
			setting: models.TrustCenterNDARequestSetting{
				AllowRoleAccount: true,
			},
		},
		{
			name: "role restriction requires classification",
			setting: models.TrustCenterNDARequestSetting{
				AllowDisposableEmail: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.approved, validateEmailRules(nil, &tt.setting))
		})
	}
}

func TestEvaluateRulesWithoutVerifier(t *testing.T) {
	tests := []struct {
		name     string
		setting  models.TrustCenterNDARequestSetting
		approved bool
	}{
		{
			name: "allowlist",
			setting: models.TrustCenterNDARequestSetting{
				AllowDisposableEmail: true,
				AllowRoleAccount:     true,
				UseDomainAllowlist:   true,
				DomainAllowlist:      []string{ndaApprovalTestDomain},
			},
			approved: true,
		},
		{
			name: "blocklist wins",
			setting: models.TrustCenterNDARequestSetting{
				AllowDisposableEmail: true,
				AllowRoleAccount:     true,
				UseDomainAllowlist:   true,
				DomainAllowlist:      []string{ndaApprovalTestDomain},
				UseDomainBlocklist:   true,
				DomainBlocklist:      []string{ndaApprovalTestDomain},
			},
		},
		{
			name: "no positive criteria",
			setting: models.TrustCenterNDARequestSetting{
				AllowDisposableEmail: true,
				AllowRoleAccount:     true,
			},
			approved: true,
		},
		{
			name: "allowlist approves before work restriction",
			setting: models.TrustCenterNDARequestSetting{
				AllowDisposableEmail: true,
				AllowRoleAccount:     true,
				WorkEmailOnly:        true,
				UseDomainAllowlist:   true,
				DomainAllowlist:      []string{ndaApprovalTestDomain},
			},
			approved: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			approved, err := evaluateRules(t.Context(), &generated.Client{}, &generated.TrustCenterNDARequest{Email: "signer@" + ndaApprovalTestDomain}, &tt.setting)
			require.NoError(t, err)
			assert.Equal(t, tt.approved, approved)
		})
	}
}
