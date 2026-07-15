package email

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// renderSubprocessorMessage renders the subprocessor notification through its dispatcher and returns the
// resulting HTML body, exercising the same decode + interpolate + render path used for a real send
func renderSubprocessorMessage(t *testing.T, req SubprocessorNotificationRequest) string {
	t.Helper()

	dispatcher, ok := DispatcherByKey(SubprocessorNotificationOp.Name())
	require.True(t, ok)

	client := &Client{Config: *MockRuntimeConfig()}

	payload, err := json.Marshal(req)
	require.NoError(t, err)

	msg, err := dispatcher.RenderMessage(context.Background(), client, payload)
	require.NoError(t, err)

	return msg.GetHTML()
}

// TestSubprocessorNotificationListUnsubscribeHeader verifies the RFC 8058 one-click header is derived
// from the per-recipient footer unsubscribe URL: same origin + token, pointed at /api/unsubscribe with the
// page slug dropped (the token alone identifies the subscriber)
func TestSubprocessorNotificationListUnsubscribeHeader(t *testing.T) {
	t.Parallel()

	req := SubprocessorNotificationRequest{
		RecipientInfo:  RecipientInfo{Email: "dolores@example.com"},
		Subprocessors:  sampleSubprocessorEntries(),
		UnsubscribeURL: "https://trust.openlane.io/securecorp/unsubscribe?token=tok_abc123",
	}

	dispatcher, ok := DispatcherByKey(SubprocessorNotificationOp.Name())
	require.True(t, ok)

	client := &Client{Config: *MockRuntimeConfig()}
	payload, err := json.Marshal(req)
	require.NoError(t, err)

	msg, err := dispatcher.RenderMessage(context.Background(), client, payload)
	require.NoError(t, err)

	assert.Equal(t, "<https://trust.openlane.io/api/unsubscribe?token=tok_abc123>", msg.Headers["List-Unsubscribe"])
	assert.Equal(t, "List-Unsubscribe=One-Click", msg.Headers["List-Unsubscribe-Post"])
}

// sampleSubprocessorEntries returns one entry of each change kind for rendering assertions
func sampleSubprocessorEntries() []SubprocessorEntry {
	return []SubprocessorEntry{
		{Name: "Amazon Web Services", Change: "Added"},
		{Name: "Stripe", Change: "Updated"},
		{Name: "Twilio", Change: "Removed"},
	}
}

// TestSubprocessorNotificationSubject verifies the operation composes the subject from the trust
// center's company name, falling back to the generic copy when there is none
func TestSubprocessorNotificationSubject(t *testing.T) {
	t.Parallel()

	dispatcher := testDispatcher[SubprocessorNotificationRequest](t, SubprocessorNotificationOp.Name())

	cfg := *MockRuntimeConfig()

	branded := dispatcher.Subject(cfg, SubprocessorNotificationRequest{
		TrustCenterBranding: TrustCenterBranding{CompanyName: "SecureCorp"},
	})
	assert.Equal(t, "SecureCorp Trust Center Subprocessor Update", branded)

	generic := dispatcher.Subject(cfg, SubprocessorNotificationRequest{})
	assert.Equal(t, subprocessorNotificationSubject, generic)
}

// TestSubprocessorNotificationBrandingOverride verifies the subprocessor notification applies the trust
// center branding overlay when present, composes the copy from the company name, renders each changed
// vendor (name and change) without a logo, and resolves the per-recipient unsubscribe link from the token
func TestSubprocessorNotificationBrandingOverride(t *testing.T) {
	req := SubprocessorNotificationRequest{
		RecipientInfo: RecipientInfo{Email: "dolores@example.com", FirstName: "Dolores", UnsubscribeToken: "tok_dolores"},
		TrustCenterBranding: TrustCenterBranding{
			CompanyName: "SecureCorp",
			LogoURL:     "https://securecorp.example.com/logo.png",
			AccentColor: "#3fc2b4",
		},
		Subprocessors:  sampleSubprocessorEntries(),
		TrustCenterURL: "https://securecorp.example.com/trust",
		UnsubscribeURL: "https://securecorp.example.com/unsubscribe?token={{ .unsubscribeToken }}",
	}

	html := renderSubprocessorMessage(t, req)

	// the copy leads with the trust center's company name
	assert.Contains(t, html, "SecureCorp has updated its subprocessors")

	// trust center identity is applied: company name, and the logo exactly once (header slot
	// only, no in-body hero logo)
	assert.Contains(t, html, "SecureCorp")
	assert.Equal(t, 1, strings.Count(html, "https://securecorp.example.com/logo.png"))

	// the accent surfaces only as decorative borders; containers and buttons keep the system styling
	assert.Contains(t, html, "border-top:4px solid #3fc2b4")
	assert.Contains(t, html, "border-left:3px solid #3fc2b4")
	assert.Contains(t, html, "background-color:#14171e") // the CTA button keeps the system color
	assert.NotContains(t, html, "background-color:#3fc2b4")

	// platform marketing chrome is stripped from trust center subscriber emails
	assert.NotContains(t, html, MockRuntimeConfig().Tagline)
	assert.NotContains(t, html, MockRuntimeConfig().HeaderText)

	// the changed vendor list renders just names and change labels (no logo, link, or countries)
	assert.Contains(t, html, "Amazon Web Services")
	assert.Contains(t, html, "Stripe")
	assert.Contains(t, html, "Twilio")
	assert.Contains(t, html, "Added")
	assert.Contains(t, html, "Updated")
	assert.Contains(t, html, "Removed")

	// the link back to the trust center is the call-to-action button
	assert.Contains(t, html, `href="https://securecorp.example.com/trust"`)

	// the per-recipient unsubscribe link is resolved from the token
	assert.Contains(t, html, "https://securecorp.example.com/unsubscribe?token=tok_dolores")
}

// TestSubprocessorNotificationSystemBrandingFallback verifies the notification falls back to the Openlane
// system branding and the generic copy when the trust center supplies no branding overlay values
func TestSubprocessorNotificationSystemBrandingFallback(t *testing.T) {
	cfg := MockRuntimeConfig()

	req := SubprocessorNotificationRequest{
		RecipientInfo: RecipientInfo{Email: "dolores@example.com", FirstName: "Dolores"},
		Subprocessors: sampleSubprocessorEntries(),
	}

	html := renderSubprocessorMessage(t, req)

	// no trust center company name leaks in; the system config branding and generic copy are used
	// instead (the title's apostrophe renders HTML-escaped), and no accent border is drawn
	assert.NotContains(t, html, "SecureCorp")
	assert.Contains(t, html, cfg.CompanyName)
	assert.Contains(t, html, "updated our subprocessors")
	assert.NotContains(t, html, "border-top:4px solid")
}
