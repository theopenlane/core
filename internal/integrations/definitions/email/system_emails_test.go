package email

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testDispatcher looks up a registered email dispatcher by key and returns the concrete Operation[T]
func testDispatcher[T Recipient](t *testing.T, key string) Operation[T] {
	t.Helper()

	d, ok := DispatcherByKey(key)
	require.True(t, ok, "dispatcher %q not registered", key)

	return d.(Operation[T])
}

// TestSystemEmailSubjects verifies subject line generation for all system email operations
func TestSystemEmailSubjects(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "Acme",
		ProductURL:  "https://app.acme.com",
	}

	verify := testDispatcher[VerifyEmailRequest](t, "VerifyEmailRequest")
	welcome := testDispatcher[WelcomeRequest](t, "WelcomeRequest")
	invite := testDispatcher[InviteRequest](t, "InviteRequest")
	inviteJoined := testDispatcher[InviteJoinedRequest](t, "InviteJoinedRequest")
	resetRequest := testDispatcher[PasswordResetEmailRequest](t, "PasswordResetEmailRequest")
	resetSuccess := testDispatcher[PasswordResetSuccessRequest](t, "PasswordResetSuccessRequest")
	subscribe := testDispatcher[SubscribeRequest](t, "SubscribeRequest")
	verifyBilling := testDispatcher[VerifyBillingRequest](t, "VerifyBillingRequest")
	tcNDARequest := testDispatcher[TrustCenterNDARequestEmail](t, "TrustCenterNDARequestEmail")
	tcNDASigned := testDispatcher[TrustCenterNDASignedEmail](t, "TrustCenterNDASignedEmail")
	tcAuth := testDispatcher[TrustCenterAuthEmail](t, "TrustCenterAuthEmail")
	questionnaireAuth := testDispatcher[QuestionnaireAuthEmail](t, "QuestionnaireAuthEmail")
	billingChanged := testDispatcher[BillingEmailChangedEmail](t, "BillingEmailChangedEmail")

	tests := []struct {
		name     string
		subject  string
		contains []string
	}{
		{
			name: "verify email",
			subject: verify.Subject(cfg, VerifyEmailRequest{
				RecipientInfo: RecipientInfo{Email: "a@b.com"},
				Token:         "tok123",
			}),
			contains: []string{"verify", "Acme"},
		},
		{
			name:     "welcome",
			subject:  welcome.Subject(cfg, WelcomeRequest{}),
			contains: []string{"Welcome", "Acme"},
		},
		{
			name: "invite",
			subject: invite.Subject(cfg, InviteRequest{
				InviterName: "Bob",
				Token:       "tok",
			}),
			contains: []string{"Bob", "Acme"},
		},
		{
			name: "invite joined",
			subject: inviteJoined.Subject(cfg, InviteJoinedRequest{
				OrgName: "OrgX",
			}),
			contains: []string{"Acme"},
		},
		{
			name: "password reset request",
			subject: resetRequest.Subject(cfg, PasswordResetEmailRequest{
				Token: "tok",
			}),
			contains: []string{"Acme", "Password Reset"},
		},
		{
			name:     "password reset success",
			subject:  resetSuccess.Subject(cfg, PasswordResetSuccessRequest{}),
			contains: []string{"Acme", "Password Reset", "Confirmation"},
		},
		{
			name: "subscribe",
			subject: subscribe.Subject(cfg, SubscribeRequest{
				Token:   "tok",
				OrgName: "OrgY",
			}),
			contains: []string{"subscribed", "Acme"},
		},
		{
			name: "verify billing",
			subject: verifyBilling.Subject(cfg, VerifyBillingRequest{
				Token: "tok",
			}),
			contains: []string{"verify", "billing", "Acme"},
		},
		{
			name: "tc nda request",
			subject: tcNDARequest.Subject(cfg, TrustCenterNDARequestEmail{
				OrgName: "SecureCorp",
			}),
			contains: []string{"SecureCorp", "NDA"},
		},
		{
			name: "tc nda signed",
			subject: tcNDASigned.Subject(cfg, TrustCenterNDASignedEmail{
				OrgName: "SecureCorp",
			}),
			contains: []string{"SecureCorp", "NDA", "Signed"},
		},
		{
			name: "tc auth",
			subject: tcAuth.Subject(cfg, TrustCenterAuthEmail{
				OrgName: "SecureCorp",
			}),
			contains: []string{"SecureCorp", "Trust Center"},
		},
		{
			name: "questionnaire auth",
			subject: questionnaireAuth.Subject(cfg, QuestionnaireAuthEmail{
				AssessmentName: "SOC2 Review",
			}),
			contains: []string{"SOC2 Review", "Acme"},
		},
		{
			name: "billing changed",
			subject: billingChanged.Subject(cfg, BillingEmailChangedEmail{
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

	body := testDispatcher[VerifyEmailRequest](t, "VerifyEmailRequest").Build(cfg, req)

	require.Len(t, body.Actions, 1)
	assert.Equal(t, "https://app.testco.com/verify?token=abc123", body.Actions[0].Button.Link)
	assert.Equal(t, "Confirm Email", body.Actions[0].Button.Text)
	assert.Equal(t, "Alice", body.Name)
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

	body := testDispatcher[InviteRequest](t, "InviteRequest").Build(cfg, req)

	require.Len(t, body.Actions, 1)
	assert.Equal(t, "https://app.testco.com/invite?token=inv-tok-456", body.Actions[0].Button.Link)
	assert.Equal(t, "Accept Invite", body.Actions[0].Button.Text)
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

	body := testDispatcher[InviteRequest](t, "InviteRequest").Build(cfg, req)

	require.NotEmpty(t, body.Intros.Unsafe)
	found := false

	for _, intro := range body.Intros.Unsafe {
		if strings.Contains(string(intro), "ADMIN") {
			found = true
		}
	}

	assert.True(t, found, "expected uppercased role in intro")
}

func TestInviteIntroEscapesUnsafeInputs(t *testing.T) {
	body := testDispatcher[InviteRequest](t, "InviteRequest").Build(RuntimeEmailConfig{
		CompanyName: "TestCo <script>",
		ProductURL:  "https://app.testco.com",
	}, InviteRequest{
		InviterName: `Alice <img src=x onerror=alert(1)>`,
		OrgName:     "Eng & <Ops>",
		Role:        "admin<script>",
		Token:       "tok",
	})

	require.NotEmpty(t, body.Intros.Unsafe)
	intro := string(body.Intros.Unsafe[0])

	assert.NotContains(t, intro, "<script>")
	assert.NotContains(t, intro, "<img")
	assert.Contains(t, intro, "TestCo &lt;script&gt;")
	assert.Contains(t, intro, "Alice &lt;img src=x onerror=alert(1)&gt;")
	assert.Contains(t, intro, "<strong>Eng &amp; &lt;Ops&gt;</strong>")
	assert.Contains(t, intro, "ADMIN&lt;SCRIPT&gt;")
	assert.NotContains(t, intro, "&amp;amp;")
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

	body := testDispatcher[InviteRequest](t, "InviteRequest").Build(cfg, req)

	require.NotEmpty(t, body.Intros.Unsafe)
	for _, intro := range body.Intros.Unsafe {
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

	body := testDispatcher[PasswordResetEmailRequest](t, "PasswordResetEmailRequest").Build(cfg, req)

	require.Len(t, body.Actions, 1)
	assert.Equal(t, "https://app.testco.com/password-reset?token=reset-xyz", body.Actions[0].Button.Link)
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

	body := testDispatcher[SubscribeRequest](t, "SubscribeRequest").Build(cfg, req)

	require.Len(t, body.Actions, 1)
	assert.Equal(t, "https://app.testco.com/subscribe/verify?token=sub-tok", body.Actions[0].Button.Link)
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

	body := testDispatcher[VerifyBillingRequest](t, "VerifyBillingRequest").Build(cfg, req)

	require.Len(t, body.Actions, 1)
	assert.Equal(t, "https://app.testco.com/billing/verify?token=bill-tok", body.Actions[0].Button.Link)
}

// TestTrustCenterNDARequestButtonColor verifies the NDA request uses trust center theme colors
func TestTrustCenterNDARequestButtonColor(t *testing.T) {
	op := testDispatcher[TrustCenterNDARequestEmail](t, "TrustCenterNDARequestEmail")

	body := op.Build(RuntimeEmailConfig{}, TrustCenterNDARequestEmail{
		OrgName: "SecureCorp",
		NDAURL:  "https://trust.securecorp.com/sign",
	})

	require.Len(t, body.Actions, 1)
	assert.Equal(t, tcButtonColor, body.Actions[0].Button.Color)
	assert.Equal(t, tcButtonTextColor, body.Actions[0].Button.TextColor)
	assert.Equal(t, "https://trust.securecorp.com/sign", body.Actions[0].Button.Link)
}

// TestTrustCenterNDASignedContent verifies the NDA signed confirmation content
func TestTrustCenterNDASignedContent(t *testing.T) {
	op := testDispatcher[TrustCenterNDASignedEmail](t, "TrustCenterNDASignedEmail")

	body := op.Build(RuntimeEmailConfig{}, TrustCenterNDASignedEmail{
		OrgName:        "SecureCorp",
		TrustCenterURL: "https://trust.securecorp.com",
	})

	require.Len(t, body.Actions, 1)
	assert.Equal(t, "Visit Trust Center", body.Actions[0].Button.Text)
	assert.Equal(t, "https://trust.securecorp.com", body.Actions[0].Button.Link)
	assert.Contains(t, body.Intros.Paragraphs[0], "SecureCorp")
}

// TestTrustCenterNDASignedAttachment verifies attachment options are returned when data is present
func TestTrustCenterNDASignedAttachment(t *testing.T) {
	op := testDispatcher[TrustCenterNDASignedEmail](t, "TrustCenterNDASignedEmail")

	opts := op.MessageOptions(RuntimeEmailConfig{}, TrustCenterNDASignedEmail{
		AttachmentFilename: "nda-signed.pdf",
		AttachmentData:     []byte("pdf-content"),
	})

	require.Len(t, opts, 1)
}

// TestTrustCenterNDASignedNoAttachment verifies nil options when no attachment
func TestTrustCenterNDASignedNoAttachment(t *testing.T) {
	op := testDispatcher[TrustCenterNDASignedEmail](t, "TrustCenterNDASignedEmail")

	opts := op.MessageOptions(RuntimeEmailConfig{}, TrustCenterNDASignedEmail{})

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

	body := testDispatcher[QuestionnaireAuthEmail](t, "QuestionnaireAuthEmail").Build(cfg, req)

	require.Len(t, body.Actions, 1)
	assert.Equal(t, "https://q.auditco.com/auth?tok=abc", body.Actions[0].Button.Link)
	assert.Contains(t, body.Title, "AuditCo")
	assert.Contains(t, body.Intros.Paragraphs[0], "SOC2 Review")
}

// TestQuestionnaireFromOverride verifies the from address override when configured
func TestQuestionnaireFromOverride(t *testing.T) {
	cfg := RuntimeEmailConfig{
		QuestionnaireEmail: "questionnaire@acme.com",
	}

	op := testDispatcher[QuestionnaireAuthEmail](t, "QuestionnaireAuthEmail")
	opts := op.MessageOptions(cfg, QuestionnaireAuthEmail{})

	require.Len(t, opts, 1)
}

// TestQuestionnaireNoFromOverride verifies nil when no questionnaire email configured
func TestQuestionnaireNoFromOverride(t *testing.T) {
	op := testDispatcher[QuestionnaireAuthEmail](t, "QuestionnaireAuthEmail")
	opts := op.MessageOptions(RuntimeEmailConfig{}, QuestionnaireAuthEmail{})

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

	body := testDispatcher[BillingEmailChangedEmail](t, "BillingEmailChangedEmail").Build(cfg, req)

	assert.Equal(t, "Billing Email Changed", body.Title)
	assert.Contains(t, body.Intros.Paragraphs[0], "BillOrg")

	require.NotEmpty(t, body.Dictionary.Cells)
	assert.Equal(t, "old@billing.com", body.Dictionary.Cells[0].Value)
	assert.Equal(t, "new@billing.com", body.Dictionary.Cells[1].Value)
}

// TestAllEmailOperationsCount verifies every dispatcher surfaces exactly one registration
func TestAllEmailOperationsCount(t *testing.T) {
	ops := AllEmailOperations()

	assert.Len(t, ops, len(dispatchers))
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
