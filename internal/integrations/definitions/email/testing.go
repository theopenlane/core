package email

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/samber/lo"
	"github.com/theopenlane/newman/providers/mock"

	"github.com/theopenlane/core/internal/ent/generated"
)

// MockRuntimeConfig returns a RuntimeEmailConfig backed by the mock provider
// for use in integration test suites
func MockRuntimeConfig() *RuntimeEmailConfig {
	return &RuntimeEmailConfig{ //nolint:gosec // test-only mock credentials
		APIKey:         "mock-api-key",
		Provider:       ProviderMock,
		FromEmail:      "test@example.com",
		CompanyName:    "MITB Company",
		CompanyAddress: "123 Paper street",
		Corporation:    "MITB Corp",
		SupportEmail:   "support@example.com",
		LogoURL:        "https://example.com/logo.png",
		HeaderLogoURL:  "https://example.com/icon.png",
		HeaderText:     "MITB Portal",
		RootURL:        "https://example.com",
		ProductURL:     "https://app.example.com",
		DocsURL:        "https://docs.example.com",
		TermsURL:       "https://example.com/terms",
		PrivacyURL:     "https://example.com/privacy",
		Tagline:        "Compliance without the busywork",
		Social: []SocialLink{
			{Platform: "GitHub", IconURL: "https://example.com/github.png", URL: "https://github.com/example"},
			{Platform: "LinkedIn", IconURL: "https://example.com/linkedin.png", URL: "https://linkedin.com/company/example"},
		},
	}
}

// MockSenderFromClient extracts the *mock.EmailSender from an Client.
// Returns nil if the sender is not a mock
func MockSenderFromClient(client any) *mock.EmailSender {
	ec, ok := client.(*Client)
	if !ok {
		return nil
	}

	ms, ok := ec.Sender.(*mock.EmailSender)
	if !ok {
		return nil
	}

	return ms
}

// TrustCenterSettingFixture returns a sample trust center setting with branding populated, used to
// preview and test what a trust center update email looks like with customer branding applied
func TrustCenterSettingFixture() *generated.TrustCenterSetting {
	return &generated.TrustCenterSetting{
		CompanyName:              "SecureCorp",
		LogoRemoteURL:            lo.ToPtr("https://securecorp.example.com/logo.png"),
		PrimaryColor:             "#0f3d3a",
		AccentColor:              "#3fc2b4",
		BackgroundColor:          "#e8eaed",
		SecondaryBackgroundColor: "#ffffff",
		ForegroundColor:          "#14171e",
	}
}

// TrustCenterUpdateTemplateFixture returns a sample email template configured for a trust center
// update, with branded message content and a tokenized unsubscribe link
func TrustCenterUpdateTemplateFixture() *generated.EmailTemplate {
	return &generated.EmailTemplate{
		Key: TrustCenterUpdateTemplate,
		Defaults: map[string]any{
			"subject":        "{{ .companyName }} trust center update",
			"title":          "Hi {{ .firstName }}, an update from {{ .companyName }}",
			"intros":         []any{"We've updated our subprocessor list.", "Review the changes in our trust center."},
			"buttonText":     "View Trust Center",
			"buttonLink":     "https://securecorp.example.com/trust",
			"unsubscribeURL": "https://securecorp.example.com/unsubscribe?token={{ .unsubscribeToken }}",
		},
	}
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
// When defaultBranding is true, the trust center branding overlay is stripped from trust-center emails
// so they render with the Openlane system defaults (the fallback used when a trust center configures no
// branding), letting both branding variants be previewed. Returns nil when no fixture is defined
func TestFixture(name, toEmail string, defaultBranding bool) json.RawMessage {
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
			// production sets only Email on the invite recipient (no name), so there is no greeting name
			RecipientInfo: RecipientInfo{Email: toEmail},
			InviterName:   "Jane Smith",
			OrgName:       "Acme Corp",
			Role:          "admin",
			Token:         "test-invite-token-12345",
			NewUser:       true,
		},
		"InviteJoinedRequest": InviteJoinedRequest{
			RecipientInfo: RecipientInfo{Email: toEmail},
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
			RecipientInfo:  r,
			OrgName:        "Acme Corp",
			Token:          "test-subscribe-token-12345",
			VerifyURL:      "https://trustcenter.example.com/acme/subscribe/verify?token=test-subscribe-token-12345",
			UnsubscribeURL: "https://trustcenter.example.com/acme/unsubscribe?token=test-subscribe-token-12345",
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
			TrustCenterURL:     "https://trustcenter.example.com/securecorp?token=test",
			AttachmentFilename: "signed_nda_file.pdf",
			AttachmentData:     testAttestedNDAPDF(),
		},
		"TrustCenterAuthEmail": TrustCenterAuthEmail{
			RecipientInfo: r,
			OrgName:       "SecureCorp",
			AuthURL:       "https://trustcenter.example.com/securecorp?token=test",
		},
		"TrustCenterNDAApprovalRequestEmail": TrustCenterNDAApprovalRequestEmail{
			RecipientInfo:  RecipientInfo{Email: toEmail, Recipients: []string{toEmail}, FirstName: r.FirstName, LastName: r.LastName},
			OrgName:        "SecureCorp",
			RequesterName:  "Dolores Abernathy",
			RequesterEmail: "dolores.abernathy@example.com",
		},
		"QuestionnaireAuthEmail": QuestionnaireAuthEmail{
			RecipientInfo:  r,
			OrgName:        "Acme Corp",
			AssessmentName: "SOC 2 Type II Review",
			AuthURL:        "https://questionnaire.example.com/questionnaire?token=test",
		},
		"BillingEmailChangedEmail": BillingEmailChangedEmail{
			RecipientInfo:   r,
			OrgName:         "Acme Corp",
			OldBillingEmail: "old-billing@acme.com",
			NewBillingEmail: "new-billing@acme.com",
			ChangedAt:       time.Now().UTC(),
		},
		"OrgDeletionNoticeEmail": OrgDeletionNoticeEmail{
			RecipientInfo: r,
			OrgName:       "Acme Corp",
			DeletionDate:  time.Now().UTC().AddDate(0, 0, 7), //nolint:mnd
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
			ButtonLink: "https://app.example.com/campaigns",
			Outros:     []string{"If you received this, the branded message template is working as expected.", "Questions? Reach us at {{ .supportemail }}."},
		},
		"SubprocessorNotificationRequest": SubprocessorNotificationRequest{
			RecipientInfo: RecipientInfo{
				Email:            toEmail,
				FirstName:        r.FirstName,
				LastName:         r.LastName,
				UnsubscribeToken: "test-unsubscribe-token-12345",
			},
			Subject:   "SecureCorp subprocessor update",
			Preheader: "Review the latest changes to our subprocessor list",
			Title:     "We've updated our subprocessors",
			Intros:    []string{"The subprocessors we use have changed. You can review the full list anytime in our trust center."},
			Subprocessors: []SubprocessorEntry{
				{Name: "Amazon Web Services", Change: "Added"},
				{Name: "Stripe", Change: "Updated"},
				{Name: "Twilio", Change: "Removed"},
			},
			ButtonText:          "View subprocessors",
			ButtonLink:          "https://securecorp.example.com/trust",
			UnsubscribeURL:      "https://securecorp.example.com/unsubscribe?token={{ .unsubscribeToken }}",
			CompanyName:         "SecureCorp",
			LogoURL:             "https://securecorp.example.com/logo.png",
			PrimaryColor:        "#0f3d3a",
			ButtonColor:         "#3fc2b4",
			BodyBackgroundColor: "#e8eaed",
			CardBackgroundColor: "#ffffff",
			TextColor:           "#14171e",
		},
	}

	fixture, ok := fixtures[name]
	if !ok {
		return nil
	}

	// preview the Openlane fallback by clearing the trust center branding overlay
	if defaultBranding {
		if sub, ok := fixture.(SubprocessorNotificationRequest); ok {
			sub.CompanyName = ""
			sub.Corporation = ""
			sub.LogoURL = ""
			sub.PrimaryColor = ""
			sub.ButtonColor = ""
			sub.BodyBackgroundColor = ""
			sub.CardBackgroundColor = ""
			sub.TextColor = ""
			fixture = sub
		}
	}

	data, err := json.Marshal(fixture)
	if err != nil {
		return nil
	}

	return data
}

const testPDFFontSize = 14

// testAttestedNDAPDF generates a valid two-page PDF simulating an NDA with attestation page
func testAttestedNDAPDF() []byte {
	page1 := testMinimalPDF("Non-Disclosure Agreement — SecureCorp")
	page2 := testMinimalPDF("Signature Certification — Test Fixture")

	var buf bytes.Buffer

	if err := api.MergeRaw([]io.ReadSeeker{
		bytes.NewReader(page1),
		bytes.NewReader(page2),
	}, &buf, false, nil); err != nil {
		return page1
	}

	return buf.Bytes()
}

// testMinimalPDF creates a valid single-page PDF with the given title text
func testMinimalPDF(title string) []byte {
	page := map[string]any{
		"paper":  "A4P",
		"origin": "UpperLeft",
		"fonts": map[string]any{
			"f": map[string]any{"name": "Helvetica-Bold", "size": testPDFFontSize},
		},
		"pages": map[string]any{
			"1": map[string]any{
				"content": map[string]any{
					"text": []map[string]any{
						{"value": title, "pos": [2]float64{20, 20}, "font": map[string]any{"name": "$f"}},
					},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(page)

	var buf bytes.Buffer

	_ = api.Create(nil, bytes.NewReader(jsonData), &buf, nil)

	return buf.Bytes()
}
