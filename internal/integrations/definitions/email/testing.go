package email

import (
	"encoding/json"
	"time"

	"github.com/theopenlane/newman/providers/mock"
)

// MockRuntimeConfig returns a RuntimeEmailConfig backed by the mock provider
// for use in integration test suites
func MockRuntimeConfig() *RuntimeEmailConfig {
	return &RuntimeEmailConfig{
		APIKey:         "mock-api-key",
		Provider:       ProviderMock,
		FromEmail:      "test@example.com",
		CompanyName:    "MITB Company",
		CompanyAddress: "123 Paper street",
		Corporation:    "MITB Corp",
		SupportEmail:   "support@example.com",
		LogoURL:        "https://example.com/logo.png",
		RootURL:        "https://example.com",
		ProductURL:     "https://app.example.com",
		DocsURL:        "https://docs.example.com",
	}
}

// MockSenderFromClient extracts the *mock.EmailSender from an EmailClient.
// Returns nil if the sender is not a mock
func MockSenderFromClient(client any) *mock.EmailSender {
	ec, ok := client.(*EmailClient)
	if !ok {
		return nil
	}

	ms, ok := ec.Sender.(*mock.EmailSender)
	if !ok {
		return nil
	}

	return ms
}

// TestRecipient returns a RecipientInfo with the given email and placeholder name fields
func TestRecipient(toEmail string) RecipientInfo {
	return RecipientInfo{
		Email:     toEmail,
		FirstName: "Wilfred",
		LastName:  "Netherton",
	}
}

// TestFixture returns a scaffolded JSON payload for the given dispatcher name.
// All required fields are populated with realistic test values; URL-resolved fields
// (NDAURL, AuthURL) are set directly so PreHooks that require a database are bypassed.
// Returns nil when no fixture is defined for the dispatcher
func TestFixture(name, toEmail string) json.RawMessage {
	r := TestRecipient(toEmail)

	fixtures := map[string]any{
		"VerifyEmailRequest": VerifyEmailRequest{
			RecipientInfo: r,
			Token:         "test-verify-token-12345",
		},
		"WelcomeRequest": WelcomeRequest{
			RecipientInfo: r,
		},
		"InviteRequest": InviteRequest{
			RecipientInfo: r,
			InviterName:   "Jane Smith",
			OrgName:       "Acme Corp",
			Role:          "admin",
			Token:         "test-invite-token-12345",
		},
		"InviteJoinedRequest": InviteJoinedRequest{
			RecipientInfo: r,
			OrgName:       "Acme Corp",
		},
		"PasswordResetEmailRequest": PasswordResetEmailRequest{
			RecipientInfo: r,
			Token:         "test-reset-token-12345",
		},
		"PasswordResetSuccessRequest": PasswordResetSuccessRequest{
			RecipientInfo: r,
		},
		"SubscribeRequest": SubscribeRequest{
			RecipientInfo: r,
			OrgName:       "Acme Corp",
			Token:         "test-subscribe-token-12345",
		},
		"VerifyBillingRequest": VerifyBillingRequest{
			RecipientInfo: r,
			Token:         "test-billing-token-12345",
		},
		"TrustCenterNDARequestEmail": TrustCenterNDARequestEmail{
			RecipientInfo: r,
			OrgName:       "SecureCorp",
			NDAURL:        "https://trustcenter.example.com/securecorp/access/sign-nda?token=test",
		},
		"TrustCenterNDASignedEmail": TrustCenterNDASignedEmail{
			RecipientInfo:      r,
			OrgName:            "SecureCorp",
			TrustCenterURL:     "https://trustcenter.example.com/securecorp",
			AttachmentFilename: "SecureCorp-NDA-Signed.pdf",
			AttachmentData:     []byte("%PDF-1.4 test NDA document content"),
		},
		"TrustCenterAuthEmail": TrustCenterAuthEmail{
			RecipientInfo: r,
			OrgName:       "SecureCorp",
			AuthURL:       "https://trustcenter.example.com/securecorp/auth?token=test",
		},
		"QuestionnaireAuthEmail": QuestionnaireAuthEmail{
			RecipientInfo:  r,
			AssessmentName: "SOC 2 Type II Review",
			AuthURL:        "https://questionnaire.example.com/auth?token=test",
		},
		"BillingEmailChangedEmail": BillingEmailChangedEmail{
			RecipientInfo:   r,
			OrgName:         "Acme Corp",
			OldBillingEmail: "old-billing@acme.com",
			NewBillingEmail: "new-billing@acme.com",
			ChangedAt:       time.Now().UTC(),
		},
		"BrandedMessageRequest": BrandedMessageRequest{
			RecipientInfo: r,
			CampaignContext: CampaignContext{
				CampaignName:        "Test Campaign",
				CampaignDescription: "A test campaign to verify branded message rendering",
			},
			Subject:    "{{ .companyName }} - Test Branded Message",
			Preheader:  "Hi {{ .firstName }}, this is a test branded message",
			Title:      "Welcome to {{ .companyName }}",
			Intros:     []string{"Hi {{ .firstName }}, this branded message was sent via the email-test CLI.", "All templates are rendering correctly for {{ .companyName }}."},
			ButtonText: "Visit Dashboard",
			ButtonLink: "{{ .rootURL }}/campaigns",
			Outros:     []string{"If you received this, the branded message template is working as expected.", "Questions? Reach us at {{ .supportEmail }}."},
		},
	}

	fixture, ok := fixtures[name]
	if !ok {
		return nil
	}

	data, err := json.Marshal(fixture)
	if err != nil {
		return nil
	}

	return data
}
