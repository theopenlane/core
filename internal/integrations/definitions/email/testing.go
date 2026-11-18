package email

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/theopenlane/newman/providers/mock"

	"github.com/theopenlane/core/v2/internal/ent/generated"
)

// MockRuntimeConfig returns a RuntimeEmailConfig backed by the mock provider
// for use in integration test suites
func MockRuntimeConfig() *RuntimeEmailConfig {
	return &RuntimeEmailConfig{ //nolint:gosec // test-only mock credentials
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
		APIURL:         "https://api.example.com",
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
		LogoRemoteURL:            new("https://securecorp.example.com/logo.png"),
		PrimaryColor:             "#0f3d3a",
		AccentColor:              "#3fc2b4",
		BackgroundColor:          "#e8eaed",
		SecondaryBackgroundColor: "#ffffff",
		ForegroundColor:          "#14171e",
	}
}

// cloudflareTrustCenterBranding returns a production-like trust center branding overlay for the
// trust center email fixtures, using Cloudflare's real logo and brand orange so test sends preview
// how an actual branded customer email renders (logo resolves, accent border shows a real palette)
func cloudflareTrustCenterBranding() TrustCenterBranding {
	return TrustCenterBranding{
		CompanyName: "Cloudflare",
		LogoURL:     "https://upload.wikimedia.org/wikipedia/commons/thumb/4/4b/Cloudflare_Logo.svg/330px-Cloudflare_Logo.svg.png",
		AccentColor: "#f6821f",
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
// When defaultBranding is true, the branding is stripped from brandable emails so they render with the
// Openlane system defaults (the fallback used when no branding is configured), letting both branding
// variants be previewed. Returns nil when no fixture is defined
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
			Email:       toEmail,
			InviterName: "Jane Smith",
			OrgName:     "Acme Corp",
			Role:        "admin",
			Token:       "test-invite-token-12345",
			NewUser:     true,
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
			RecipientInfo:       r,
			TrustCenterBranding: cloudflareTrustCenterBranding(),
			OrgName:             "Cloudflare",
			Token:               "test-subscribe-token-12345",
			VerifyURL:           "https://www.cloudflare.com/trust-hub/subscribe/verify?token=test-subscribe-token-12345",
			UnsubscribeURL:      "https://www.cloudflare.com/trust-hub/unsubscribe?token=test-subscribe-token-12345",
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
			Email: toEmail, Recipients: []string{toEmail}, FirstName: r.FirstName, LastName: r.LastName,
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
		"QuestionnaireAccessRequest": QuestionnaireAccessRequest{
			RecipientInfo:   r,
			OrgName:         cloudflareTrustCenterBranding().CompanyName,
			AssessmentName:  "Acceptable Use Policy",
			AuthURL:         "https://questionnaire.example.com/questionnaire?token=test",
			CampaignName:    "Annual Policy Sign-Off",
			Subject:         "Action required: sign off on the {{ .questionnaireName }}",
			Preheader:       "{{ .orgName }} needs your acknowledgement of the {{ .questionnaireName }}",
			Title:           "Hi {{ .firstName }}, please review and sign off on the {{ .questionnaireName }}",
			Intros:          []string{"{{ .orgName }} has updated the {{ .questionnaireName }} and every employee needs to acknowledge it.", "Read through the policy and confirm that you understand and agree to it."},
			ButtonText:      "Review and Sign Off",
			LogoURL:         cloudflareTrustCenterBranding().LogoURL,
			PrimaryColor:    cloudflareTrustCenterBranding().AccentColor,
			ButtonColor:     cloudflareTrustCenterBranding().AccentColor,
			ButtonTextColor: "#ffffff",
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
			RecipientInfo:   r,
			CampaignName:    "Cloudflare Customer Update",
			Subject:         "{{ .companyName }} - Security update for {{ .firstName }}",
			Preheader:       "A quick update from the {{ .companyName }} security team",
			Title:           "Hi {{ .firstName }}, here is your quarterly security update",
			Intros:          []string{"Our latest SOC 2 Type II report is now available.", "Visit the trust hub to review the report and our updated subprocessor list."},
			ButtonText:      "Visit Trust Hub",
			ButtonLink:      "https://www.cloudflare.com/trust-hub/",
			Outros:          []string{"Questions? Reply to this email and our team will help out."},
			CompanyName:     cloudflareTrustCenterBranding().CompanyName,
			Corporation:     "Cloudflare, Inc.",
			CompanyAddress:  "101 Townsend St, San Francisco, CA 94107",
			LogoURL:         cloudflareTrustCenterBranding().LogoURL,
			PrimaryColor:    cloudflareTrustCenterBranding().AccentColor,
			ButtonColor:     cloudflareTrustCenterBranding().AccentColor,
			ButtonTextColor: "#ffffff",
		},
		"SubprocessorNotificationRequest": SubprocessorNotificationRequest{
			Email:               toEmail,
			FirstName:           r.FirstName,
			LastName:            r.LastName,
			UnsubscribeToken:    "test-unsubscribe-token-12345",
			TrustCenterBranding: cloudflareTrustCenterBranding(),
			Subprocessors: []SubprocessorEntry{
				{Name: "Amazon Web Services", Change: "Added"},
				{Name: "Stripe", Change: "Updated"},
				{Name: "Twilio", Change: "Removed"},
			},
			TrustCenterURL: "https://www.cloudflare.com/trust-hub/",
			UnsubscribeURL: "https://www.cloudflare.com/trust-hub/unsubscribe?token={{ .unsubscribeToken }}",
		},
		TrustCenterUpdateTemplate: TrustCenterUpdateRequest{
			Email:               toEmail,
			FirstName:           r.FirstName,
			LastName:            r.LastName,
			UnsubscribeToken:    "test-unsubscribe-token-12345",
			TrustCenterBranding: cloudflareTrustCenterBranding(),
			PostTitle:           "SOC 2 Type II report published",
			PostText:            "Our latest SOC 2 Type II report is now available in the trust center.\nRequest access to review the full report.",
			TrustCenterURL:      "https://www.cloudflare.com/trust-hub/",
			UnsubscribeURL:      "https://www.cloudflare.com/trust-hub/unsubscribe?token={{ .unsubscribeToken }}",
		},
	}

	fixture, ok := fixtures[name]
	if !ok {
		return nil
	}

	if defaultBranding {
		fixture = withDefaultBranding(fixture)
	}

	data, err := json.Marshal(fixture)
	if err != nil {
		return nil
	}

	return data
}

// withDefaultBranding returns a copy of the fixture with its branding cleared to preview the Openlane fallback
func withDefaultBranding(fixture any) any {
	switch req := fixture.(type) {
	case SubprocessorNotificationRequest:
		req.TrustCenterBranding = TrustCenterBranding{}
		return req
	case TrustCenterUpdateRequest:
		req.TrustCenterBranding = TrustCenterBranding{}
		return req
	case SubscribeRequest:
		req.TrustCenterBranding = TrustCenterBranding{}
		req.OrgName = ""

		return req
	case QuestionnaireAccessRequest:
		req.BrandOverrides = BrandOverrides{}
		return req
	case BrandedMessageRequest:
		req.BrandOverrides = BrandOverrides{}
		return req
	default:
		return fixture
	}
}

const testPDFFontSize = 14

// testAttestedNDAPDF generates a valid two-page PDF simulating an NDA with attestation page
func testAttestedNDAPDF() []byte {
	page1 := testMinimalPDF("Non-Disclosure Agreement — SecureCorp")
	page2 := testMinimalPDF("Signature Certification — Test Fixture")

	var buf bytes.Buffer

	if err := api.MergeRaw(context.Background(), []io.ReadSeeker{
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

	_ = api.Create(context.Background(), nil, bytes.NewReader(jsonData), &buf, nil)

	return buf.Bytes()
}

const (
	// MockResendWebhookSecret is the Svix signing secret MockProviderRuntimeConfig verifies Resend webhooks with
	MockResendWebhookSecret = "whsec_bW9jay1yZXNlbmQtd2ViaG9vay1zZWNyZXQ=" //nolint:gosec // test-only mock credential
	// MockPostmarkWebhookSecret is the basic auth password MockProviderRuntimeConfig verifies Postmark webhooks with
	MockPostmarkWebhookSecret = "mock-postmark-webhook-secret" //nolint:gosec // test-only mock credential
)

// MockProviderRuntimeConfig returns the mock runtime config with provider selected and a placeholder credential, so the definition registers that provider's inbound webhook with the mock webhook secret
func MockProviderRuntimeConfig(provider string) *RuntimeEmailConfig {
	cfg := MockRuntimeConfig()
	cfg.Provider = provider
	cfg.Resend = ResendConfig{APIKey: "mock-api-key", WebhookSecret: MockResendWebhookSecret}                 //nolint:gosec // test-only mock credential
	cfg.Postmark = PostmarkConfig{ServerToken: "mock-server-token", WebhookSecret: MockPostmarkWebhookSecret} //nolint:gosec // test-only mock credential

	return cfg
}

// SvixSignature computes the v1 Svix signature header value for payload under secret, id, and timestamp, as Resend sends it
func SvixSignature(secret, id, timestamp string, payload []byte) string {
	key, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(secret, "whsec_"))
	if err != nil {
		return ""
	}

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(id + "." + timestamp + "."))
	mac.Write(payload)

	return "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
