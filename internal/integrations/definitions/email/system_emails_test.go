package email

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSystemEmailSubjects verifies subject line generation for all system email operations
func TestSystemEmailSubjects(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "Acme",
		ProductURL:  "https://app.acme.com",
	}

	tests := []struct {
		name     string
		subject  string
		contains []string
	}{
		{
			name: "verify email",
			subject: verifyEmail.Subject(cfg, VerifyEmailRequest{
				RecipientInfo: RecipientInfo{Email: "a@b.com"},
				Token:         "tok123",
			}),
			contains: []string{"verify", "Acme"},
		},
		{
			name:     "welcome",
			subject:  welcomeEmail.Subject(cfg, WelcomeRequest{}),
			contains: []string{"Welcome", "Acme"},
		},
		{
			name: "invite",
			subject: inviteEmail.Subject(cfg, InviteRequest{
				InviterName: "Bob",
				Token:       "tok",
			}),
			contains: []string{"Bob", "Acme"},
		},
		{
			name: "invite joined",
			subject: inviteJoinedEmail.Subject(cfg, InviteJoinedRequest{
				OrgName: "OrgX",
			}),
			contains: []string{"Acme"},
		},
		{
			name: "password reset request",
			subject: resetRequestEmail.Subject(cfg, PasswordResetEmailRequest{
				Token: "tok",
			}),
			contains: []string{"Acme", "Password Reset"},
		},
		{
			name:     "password reset success",
			subject:  resetSuccessEmail.Subject(cfg, PasswordResetSuccessRequest{}),
			contains: []string{"Acme", "Password Reset", "Confirmation"},
		},
		{
			name: "subscribe",
			subject: subscribeEmail.Subject(cfg, SubscribeRequest{
				Token:   "tok",
				OrgName: "OrgY",
			}),
			contains: []string{"subscribed", "Acme"},
		},
		{
			name: "verify billing",
			subject: verifyBillingEmail.Subject(cfg, VerifyBillingRequest{
				Token: "tok",
			}),
			contains: []string{"verify", "billing", "Acme"},
		},
		{
			name: "tc nda request",
			subject: tcNDARequestEmail.Subject(cfg, TrustCenterNDARequestEmail{
				OrgName: "SecureCorp",
			}),
			contains: []string{"SecureCorp", "NDA"},
		},
		{
			name: "tc nda signed",
			subject: tcNDASignedEmail.Subject(cfg, TrustCenterNDASignedEmail{
				OrgName: "SecureCorp",
			}),
			contains: []string{"SecureCorp", "NDA", "Signed"},
		},
		{
			name: "tc auth",
			subject: tcAuthEmail.Subject(cfg, TrustCenterAuthEmail{
				OrgName: "SecureCorp",
			}),
			contains: []string{"SecureCorp", "Trust Center"},
		},
		{
			name: "questionnaire auth",
			subject: questionnaireAuthEmail.Subject(cfg, QuestionnaireAuthEmail{
				AssessmentName: "SOC2 Review",
			}),
			contains: []string{"SOC2 Review", "Acme"},
		},
		{
			name: "billing changed",
			subject: billingChangedEmail.Subject(cfg, BillingEmailChangedEmail{
				OrgName: "BillOrg",
			}),
			contains: []string{"Billing", "BillOrg"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.NotEmpty(t, tc.subject)
			for _, s := range tc.contains {
				assert.Contains(t, tc.subject, s)
			}
		})
	}
}

// TestVerifyEmailURLConstruction verifies the verify email content builds the correct URL
func TestVerifyEmailURLConstruction(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "TestCo",
		ProductURL:  "https://app.testco.com",
	}

	req := VerifyEmailRequest{
		RecipientInfo: RecipientInfo{FirstName: "Alice"},
		Token:         "abc123",
	}

	content := verifyEmail.Content(cfg, req)

	require.Len(t, content.Body.Actions, 1)
	assert.Equal(t, "https://app.testco.com/verify?token=abc123", content.Body.Actions[0].Button.Link)
	assert.Equal(t, "Confirm Email", content.Body.Actions[0].Button.Text)
	assert.Equal(t, "Alice", content.Body.Name)
}

// TestInviteEmailURLConstruction verifies the invite email content builds the correct URL
func TestInviteEmailURLConstruction(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "TestCo",
		ProductURL:  "https://app.testco.com",
	}

	req := InviteRequest{
		RecipientInfo: RecipientInfo{FirstName: "Bob"},
		InviterName:   "Alice",
		OrgName:       "Engineering",
		Role:          "admin",
		Token:         "inv-tok-456",
	}

	content := inviteEmail.Content(cfg, req)

	require.Len(t, content.Body.Actions, 1)
	assert.Equal(t, "https://app.testco.com/invite?token=inv-tok-456", content.Body.Actions[0].Button.Link)
	assert.Equal(t, "Accept Invite", content.Body.Actions[0].Button.Text)
}

// TestInviteEmailRoleInIntro verifies the role is uppercased in the intro
func TestInviteEmailRoleInIntro(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "TestCo",
		ProductURL:  "https://app.testco.com",
	}

	req := InviteRequest{
		InviterName: "Alice",
		OrgName:     "Eng",
		Role:        "admin",
		Token:       "tok",
	}

	content := inviteEmail.Content(cfg, req)

	require.NotEmpty(t, content.Body.IntrosUnsafe)
	found := false
	for _, intro := range content.Body.IntrosUnsafe {
		if strings.Contains(string(intro), "ADMIN") {
			found = true
		}
	}

	assert.True(t, found, "expected uppercased role in intro")
}

// TestInviteEmailNoRole verifies the intro omits role text when empty
func TestInviteEmailNoRole(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "TestCo",
		ProductURL:  "https://app.testco.com",
	}

	req := InviteRequest{
		InviterName: "Alice",
		OrgName:     "Eng",
		Token:       "tok",
	}

	content := inviteEmail.Content(cfg, req)

	require.NotEmpty(t, content.Body.IntrosUnsafe)
	for _, intro := range content.Body.IntrosUnsafe {
		assert.NotContains(t, string(intro), "role of")
	}
}

// TestPasswordResetURLConstruction verifies the reset URL is correctly built
func TestPasswordResetURLConstruction(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "TestCo",
		ProductURL:  "https://app.testco.com",
	}

	req := PasswordResetEmailRequest{
		RecipientInfo: RecipientInfo{FirstName: "Charlie"},
		Token:         "reset-xyz",
	}

	content := resetRequestEmail.Content(cfg, req)

	require.Len(t, content.Body.Actions, 1)
	assert.Equal(t, "https://app.testco.com/password-reset?token=reset-xyz", content.Body.Actions[0].Button.Link)
}

// TestSubscribeVerifyURLConstruction verifies the subscribe verification URL
func TestSubscribeVerifyURLConstruction(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "TestCo",
		ProductURL:  "https://app.testco.com",
	}

	req := SubscribeRequest{
		RecipientInfo: RecipientInfo{FirstName: "Dana"},
		OrgName:       "SubOrg",
		Token:         "sub-tok",
	}

	content := subscribeEmail.Content(cfg, req)

	require.Len(t, content.Body.Actions, 1)
	assert.Equal(t, "https://app.testco.com/subscribe/verify?token=sub-tok", content.Body.Actions[0].Button.Link)
}

// TestVerifyBillingURLConstruction verifies the billing verification URL
func TestVerifyBillingURLConstruction(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "TestCo",
		ProductURL:  "https://app.testco.com",
	}

	req := VerifyBillingRequest{
		Token: "bill-tok",
	}

	content := verifyBillingEmail.Content(cfg, req)

	require.Len(t, content.Body.Actions, 1)
	assert.Equal(t, "https://app.testco.com/billing/verify?token=bill-tok", content.Body.Actions[0].Button.Link)
}

// TestTrustCenterNDARequestButtonColor verifies the NDA request uses trust center theme colors
func TestTrustCenterNDARequestButtonColor(t *testing.T) {
	content := tcNDARequestEmail.Content(RuntimeEmailConfig{}, TrustCenterNDARequestEmail{
		OrgName: "SecureCorp",
		NDAURL:  "https://trust.securecorp.com/sign",
	})

	require.Len(t, content.Body.Actions, 1)
	assert.Equal(t, tcButtonColor, content.Body.Actions[0].Button.Color)
	assert.Equal(t, tcButtonTextColor, content.Body.Actions[0].Button.TextColor)
	assert.Equal(t, "https://trust.securecorp.com/sign", content.Body.Actions[0].Button.Link)
}

// TestTrustCenterNDASignedContent verifies the NDA signed confirmation content
func TestTrustCenterNDASignedContent(t *testing.T) {
	content := tcNDASignedEmail.Content(RuntimeEmailConfig{}, TrustCenterNDASignedEmail{
		OrgName:        "SecureCorp",
		TrustCenterURL: "https://trust.securecorp.com",
	})

	require.Len(t, content.Body.Actions, 1)
	assert.Equal(t, "Visit Trust Center", content.Body.Actions[0].Button.Text)
	assert.Equal(t, "https://trust.securecorp.com", content.Body.Actions[0].Button.Link)
	assert.Contains(t, content.Body.Intros[0], "SecureCorp")
}

// TestTrustCenterNDASignedAttachment verifies attachment options are returned when data is present
func TestTrustCenterNDASignedAttachment(t *testing.T) {
	req := TrustCenterNDASignedEmail{
		AttachmentFilename: "nda-signed.pdf",
		AttachmentData:     []byte("pdf-content"),
	}

	opts := tcNDASignedEmail.MessageOptions(RuntimeEmailConfig{}, req)

	require.Len(t, opts, 1)
}

// TestTrustCenterNDASignedNoAttachment verifies nil options when no attachment
func TestTrustCenterNDASignedNoAttachment(t *testing.T) {
	opts := tcNDASignedEmail.MessageOptions(RuntimeEmailConfig{}, TrustCenterNDASignedEmail{})

	assert.Nil(t, opts)
}

// TestQuestionnaireAuthContent verifies questionnaire auth email content
func TestQuestionnaireAuthContent(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "AuditCo",
	}

	req := QuestionnaireAuthEmail{
		AssessmentName: "SOC2 Review",
		AuthURL:        "https://q.auditco.com/auth?tok=abc",
	}

	content := questionnaireAuthEmail.Content(cfg, req)

	require.Len(t, content.Body.Actions, 1)
	assert.Equal(t, "https://q.auditco.com/auth?tok=abc", content.Body.Actions[0].Button.Link)
	assert.Contains(t, content.Body.Title, "AuditCo")
	assert.Contains(t, content.Body.Intros[0], "SOC2 Review")
}

// TestQuestionnaireFromOverride verifies the from address override when configured
func TestQuestionnaireFromOverride(t *testing.T) {
	cfg := RuntimeEmailConfig{
		QuestionnaireEmail: "questionnaire@acme.com",
	}

	opts := questionnaireAuthEmail.MessageOptions(cfg, QuestionnaireAuthEmail{})

	require.Len(t, opts, 1)
}

// TestQuestionnaireNoFromOverride verifies nil when no questionnaire email configured
func TestQuestionnaireNoFromOverride(t *testing.T) {
	opts := questionnaireAuthEmail.MessageOptions(RuntimeEmailConfig{}, QuestionnaireAuthEmail{})

	assert.Nil(t, opts)
}

// TestBillingChangedContent verifies billing change notification content
func TestBillingChangedContent(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName:  "TestCo",
		SupportEmail: "support@testco.com",
	}

	req := BillingEmailChangedEmail{
		OrgName:         "BillOrg",
		OldBillingEmail: "old@billing.com",
		NewBillingEmail: "new@billing.com",
		ChangedAt:       time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC),
	}

	content := billingChangedEmail.Content(cfg, req)

	assert.Equal(t, "Billing Email Changed", content.Body.Title)
	assert.Contains(t, content.Body.Intros[0], "BillOrg")

	require.NotEmpty(t, content.Body.ContentBlocks)
	block := string(content.Body.ContentBlocks[0])
	assert.Contains(t, block, "old@billing.com")
	assert.Contains(t, block, "new@billing.com")
}

// TestAllEmailOperationsCount verifies the expected number of operations are registered
func TestAllEmailOperationsCount(t *testing.T) {
	ops := AllEmailOperations()

	assert.Len(t, ops, 13)
}

// TestAllEmailOperationsHaveNames verifies each operation registration has a non-empty name
func TestAllEmailOperationsHaveNames(t *testing.T) {
	ops := AllEmailOperations()

	for _, op := range ops {
		assert.NotEmpty(t, op.Name, "operation should have a name")
		assert.NotEmpty(t, op.Topic, "operation should have a topic")
		assert.NotNil(t, op.Handle, "operation should have a handler")
	}
}

// TestAllEmailOperationsUniqueNames verifies no duplicate operation names
func TestAllEmailOperationsUniqueNames(t *testing.T) {
	ops := AllEmailOperations()

	seen := make(map[string]struct{}, len(ops))
	for _, op := range ops {
		_, duplicate := seen[op.Name]
		assert.False(t, duplicate, "duplicate operation name: %s", op.Name)
		seen[op.Name] = struct{}{}
	}
}
