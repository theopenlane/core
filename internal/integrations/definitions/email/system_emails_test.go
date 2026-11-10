package email

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/integrations/registry"
)

// testDispatcher looks up a registered email dispatcher by key and returns the concrete Operation[T]
func testDispatcher[T Recipient](t *testing.T, key string) Operation[T] {
	t.Helper()

	d, ok := DispatcherByKey(key)
	assert.Assert(t, ok, "dispatcher %q not registered", key)

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
	orgDeletion := testDispatcher[OrgDeletionNoticeEmail](t, "OrgDeletionNoticeEmail")

	tests := []struct {
		name     string
		subject  string
		contains []string
	}{
		{
			name: "verify email",
			subject: verify.Subject(cfg, VerifyEmailRequest{
				Email: "a@b.com",
				Token: "tok123",
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
			contains: []string{"subscription", "OrgY"},
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
				OrgName:        "Acme",
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
		{
			name: "org deletion notice",
			subject: orgDeletion.Subject(cfg, OrgDeletionNoticeEmail{
				OrgName:      "DeadOrg",
				DeletionDate: time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
			}),
			contains: []string{"Deletion", "DeadOrg"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Assert(t, len(tc.subject) > 0)
			for _, s := range tc.contains {
				assert.Check(t, is.Contains(tc.subject, s))
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
		FirstName: "Alice",
		Token:     "abc123",
	}

	body := testDispatcher[VerifyEmailRequest](t, "VerifyEmailRequest").Build(cfg, req)

	assert.Assert(t, is.Len(body.Actions, 1))
	assert.Check(t, is.Equal("https://app.testco.com/verify?token=abc123", body.Actions[0].Button.Link))
	assert.Check(t, is.Equal("Confirm Email", body.Actions[0].Button.Text))
	assert.Check(t, is.Equal("Alice", body.Name))
}

// TestInviteEmailURLConstruction verifies the invite email content builds the correct URL
func TestInviteEmailURLConstruction(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "TestCo",
		ProductURL:  "https://app.testco.com",
	}

	req := InviteRequest{
		FirstName: "Bob", Email: "bob@example.com",
		InviterName: "Alice",
		OrgName:     "Engineering",
		Role:        "admin",
		Token:       "inv-tok-456",
	}

	body := testDispatcher[InviteRequest](t, "InviteRequest").Build(cfg, req)

	assert.Assert(t, is.Len(body.Actions, 1))
	assert.Check(t, is.Equal("https://app.testco.com/invite?token=inv-tok-456&email=bob%40example.com", body.Actions[0].Button.Link))
	assert.Check(t, is.Equal("Accept Invite", body.Actions[0].Button.Text))
}

// TestInviteEmailRoleInIntro verifies the role appears in the intro
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

	assert.Assert(t, len(body.Intros.Paragraphs) > 0)
	found := false

	for _, intro := range body.Intros.Paragraphs {
		if strings.Contains(intro, "role of admin") {
			found = true
		}
	}

	assert.Check(t, found, "expected role in intro")
}

// TestInviteIntroEscapesUnsafeInputs verifies user-controlled fields are HTML-escaped in the rendered
// email, so the plain-text intro paragraphs cannot inject markup
func TestInviteIntroEscapesUnsafeInputs(t *testing.T) {
	dispatcher, ok := DispatcherByKey(InviteOp.Name())
	assert.Assert(t, ok)

	client := &Client{Config: RuntimeEmailConfig{CompanyName: "TestCo <script>", ProductURL: "https://app.testco.com"}}

	payload, err := json.Marshal(InviteRequest{
		InviterName: `Alice <img src=x onerror=alert(1)>`,
		OrgName:     "Eng & <Ops>",
		Role:        "admin<script>",
		Token:       "tok",
	})
	assert.NilError(t, err)

	msg, err := dispatcher.RenderMessage(context.Background(), client, payload)
	assert.NilError(t, err)

	html := msg.GetHTML()

	assert.Check(t, !strings.Contains(html, "<img src=x onerror=alert(1)>"))
	assert.Check(t, !strings.Contains(html, "admin<script>"))
	assert.Check(t, is.Contains(html, "Alice &lt;img src=x onerror=alert(1)&gt;"))
	assert.Check(t, is.Contains(html, "Eng &amp; &lt;Ops&gt;"))
	assert.Check(t, is.Contains(html, "admin&lt;script&gt;"))
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

	assert.Assert(t, len(body.Intros.Paragraphs) > 0)
	for _, intro := range body.Intros.Paragraphs {
		assert.Check(t, !strings.Contains(intro, "role of"))
	}
}

// TestPasswordResetURLConstruction verifies the reset URL is correctly built
func TestPasswordResetURLConstruction(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "TestCo",
		ProductURL:  "https://app.testco.com",
	}

	req := PasswordResetEmailRequest{
		FirstName: "Charlie",
		Token:     "reset-xyz",
	}

	body := testDispatcher[PasswordResetEmailRequest](t, "PasswordResetEmailRequest").Build(cfg, req)

	assert.Assert(t, is.Len(body.Actions, 1))
	assert.Check(t, is.Equal("https://app.testco.com/password-reset?token=reset-xyz", body.Actions[0].Button.Link))
}

// TestSubscribeVerifyURLTrustCenterDomain verifies the confirm button uses the caller-supplied trust
// center domain link when present, so subscribers land on the trust center rather than the API host
func TestSubscribeVerifyURLTrustCenterDomain(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "TestCo",
		ProductURL:  "https://app.testco.com",
		APIURL:      "https://api.testco.com",
	}

	req := SubscribeRequest{
		FirstName: "Dana",
		OrgName:   "SubOrg",
		Token:     "sub-tok",
		VerifyURL: "https://trust.suborg.com/subscribe/verify?token=sub-tok",
	}

	body := testDispatcher[SubscribeRequest](t, "SubscribeRequest").Build(cfg, req)

	assert.Assert(t, is.Len(body.Actions, 1))
	assert.Check(t, is.Equal("https://trust.suborg.com/subscribe/verify?token=sub-tok", body.Actions[0].Button.Link))
}

// TestSubscribeVerifyURLFallback verifies that an organization-level subscriber with no trust center to
// land on falls back to the API-direct verify endpoint
func TestSubscribeVerifyURLFallback(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "TestCo",
		ProductURL:  "https://app.testco.com",
		APIURL:      "https://api.testco.com",
	}

	req := SubscribeRequest{
		FirstName: "Dana",
		OrgName:   "SubOrg",
		Token:     "sub-tok",
	}

	body := testDispatcher[SubscribeRequest](t, "SubscribeRequest").Build(cfg, req)

	assert.Assert(t, is.Len(body.Actions, 1))
	assert.Check(t, is.Equal("https://api.testco.com/v1/subscribe/verify?token=sub-tok", body.Actions[0].Button.Link))
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

	assert.Assert(t, is.Len(body.Actions, 1))
	assert.Check(t, is.Equal("https://app.testco.com/billing/verify?token=bill-tok", body.Actions[0].Button.Link))
}

// TestTrustCenterNDARequestButtonColor verifies the NDA request uses trust center theme colors
func TestTrustCenterNDARequestButtonColor(t *testing.T) {
	op := testDispatcher[TrustCenterNDARequestEmail](t, "TrustCenterNDARequestEmail")

	body := op.Build(RuntimeEmailConfig{}, TrustCenterNDARequestEmail{
		OrgName: "SecureCorp",
		NDAURL:  "https://trust.securecorp.com/sign",
	})

	assert.Assert(t, is.Len(body.Actions, 1))
	assert.Check(t, is.Equal(tcButtonColor, body.Actions[0].Button.Color))
	assert.Check(t, is.Equal(tcButtonTextColor, body.Actions[0].Button.TextColor))
	assert.Check(t, is.Equal("https://trust.securecorp.com/sign", body.Actions[0].Button.Link))
}

// TestTrustCenterNDASignedContent verifies the NDA signed confirmation content
func TestTrustCenterNDASignedContent(t *testing.T) {
	op := testDispatcher[TrustCenterNDASignedEmail](t, "TrustCenterNDASignedEmail")

	body := op.Build(RuntimeEmailConfig{}, TrustCenterNDASignedEmail{
		OrgName:        "SecureCorp",
		TrustCenterURL: "https://trust.securecorp.com",
	})

	assert.Assert(t, is.Len(body.Actions, 1))
	assert.Check(t, is.Equal("Visit Trust Center", body.Actions[0].Button.Text))
	assert.Check(t, is.Equal("https://trust.securecorp.com", body.Actions[0].Button.Link))
	assert.Check(t, is.Contains(body.Intros.Paragraphs[0], "SecureCorp"))
}

// TestTrustCenterNDASignedAttachment verifies attachment options are returned when data is present
func TestTrustCenterNDASignedAttachment(t *testing.T) {
	op := testDispatcher[TrustCenterNDASignedEmail](t, "TrustCenterNDASignedEmail")

	opts := op.MessageOptions(RuntimeEmailConfig{}, TrustCenterNDASignedEmail{
		AttachmentFilename: "nda-signed.pdf",
		AttachmentData:     []byte("pdf-content"),
	})

	assert.Assert(t, is.Len(opts, 1))
}

// TestTrustCenterNDASignedNoAttachment verifies nil options when no attachment
func TestTrustCenterNDASignedNoAttachment(t *testing.T) {
	op := testDispatcher[TrustCenterNDASignedEmail](t, "TrustCenterNDASignedEmail")

	opts := op.MessageOptions(RuntimeEmailConfig{}, TrustCenterNDASignedEmail{})

	assert.Check(t, is.Nil(opts))
}

// TestQuestionnaireAuthContent verifies questionnaire auth email content
func TestQuestionnaireAuthContent(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "AuditCo",
	}

	req := QuestionnaireAuthEmail{
		OrgName:        "AuditCo",
		AssessmentName: "SOC2 Review",
		AuthURL:        "https://q.auditco.com/auth?tok=abc",
	}

	body := testDispatcher[QuestionnaireAuthEmail](t, "QuestionnaireAuthEmail").Build(cfg, req)

	assert.Assert(t, is.Len(body.Actions, 1))
	assert.Check(t, is.Equal("https://q.auditco.com/auth?tok=abc", body.Actions[0].Button.Link))
	assert.Check(t, is.Contains(body.Title, "AuditCo"))
	assert.Check(t, is.Contains(body.Intros.Paragraphs[0], "SOC2 Review"))
}

// TestQuestionnaireFromOverride verifies the from address override when configured
func TestQuestionnaireFromOverride(t *testing.T) {
	cfg := RuntimeEmailConfig{
		QuestionnaireEmail: "questionnaire@acme.com",
	}

	op := testDispatcher[QuestionnaireAuthEmail](t, "QuestionnaireAuthEmail")
	opts := op.MessageOptions(cfg, QuestionnaireAuthEmail{})

	assert.Assert(t, is.Len(opts, 1))
}

// TestQuestionnaireNoFromOverride verifies nil when no questionnaire email configured
func TestQuestionnaireNoFromOverride(t *testing.T) {
	op := testDispatcher[QuestionnaireAuthEmail](t, "QuestionnaireAuthEmail")
	opts := op.MessageOptions(RuntimeEmailConfig{}, QuestionnaireAuthEmail{})

	assert.Check(t, is.Nil(opts))
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

	assert.Check(t, is.Equal("Billing Email Changed", body.Title))
	assert.Check(t, is.Contains(body.Intros.Paragraphs[0], "BillOrg"))

	assert.Assert(t, len(body.Dictionary.Cells) > 0)
	assert.Check(t, is.Equal("old@billing.com", body.Dictionary.Cells[0].Value))
	assert.Check(t, is.Equal("new@billing.com", body.Dictionary.Cells[1].Value))
}

// TestOrgDeletionNoticeContent verifies org deletion notice content
func TestOrgDeletionNoticeContent(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName:  "TestCo",
		ProductURL:   "https://app.testco.com",
		SupportEmail: "support@testco.com",
	}

	req := OrgDeletionNoticeEmail{
		OrgName:      "DeadOrg",
		DeletionDate: time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
	}

	body := testDispatcher[OrgDeletionNoticeEmail](t, "OrgDeletionNoticeEmail").Build(cfg, req)

	assert.Check(t, is.Equal("Organization Deletion Notice", body.Title))
	assert.Check(t, is.Contains(body.Intros.Paragraphs[0], "DeadOrg"))

	assert.Assert(t, len(body.Dictionary.Cells) > 0)
	assert.Check(t, is.Equal("July 1, 2025", body.Dictionary.Cells[0].Value))

	assert.Assert(t, is.Len(body.Actions, 1))
	assert.Check(t, is.Equal("https://app.testco.com"+billingPath, body.Actions[0].Button.Link))
	assert.Check(t, is.Equal("Add Payment Method", body.Actions[0].Button.Text))
	// matches the questionnaire email treatment: trust center teal button on the base theme
	assert.Check(t, is.Equal(tcButtonColor, body.Actions[0].Button.Color))
	assert.Check(t, is.Equal(tcButtonTextColor, body.Actions[0].Button.TextColor))
}

// TestAllEmailOperationsCount verifies every dispatcher surfaces exactly one registration
func TestAllEmailOperationsCount(t *testing.T) {
	ops := AllEmailOperations()

	assert.Check(t, is.Len(ops, len(dispatchers)))
}

// TestAllEmailOperationsHaveNames verifies each operation registration has a non-empty name, a handler, and a topic stamped by the registry
func TestAllEmailOperationsHaveNames(t *testing.T) {
	reg := registry.New()
	assert.NilError(t, reg.RegisterAll(Builder(&RuntimeEmailConfig{}, false)))

	for _, op := range AllEmailOperations() {
		assert.Check(t, len(op.Name) > 0, "operation should have a name")
		assert.Check(t, op.Handle != nil, "operation should have a handler")

		registered, err := reg.Operation(DefinitionID.ID(), op.Name)
		assert.NilError(t, err)
		assert.Check(t, len(registered.Topic) > 0, "operation should have a topic once registered")
	}
}

// TestAllEmailOperationsUniqueNames verifies no duplicate operation names
func TestAllEmailOperationsUniqueNames(t *testing.T) {
	ops := AllEmailOperations()

	seen := make(map[string]struct{}, len(ops))
	for _, op := range ops {
		_, duplicate := seen[op.Name]
		assert.Check(t, !duplicate, "duplicate operation name: %s", op.Name)
		seen[op.Name] = struct{}{}
	}
}
