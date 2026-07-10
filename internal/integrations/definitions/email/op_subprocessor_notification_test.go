package email

import (
	"context"
	"encoding/json"
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
		Subject:        "We've updated our subprocessors",
		Title:          "We've updated our subprocessors",
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

// TestSubprocessorNotificationBrandingOverride verifies the subprocessor notification applies the trust
// center branding overlay when present, renders each changed vendor (name, countries, change) without a
// logo, and resolves the per-recipient unsubscribe link from the token
func TestSubprocessorNotificationBrandingOverride(t *testing.T) {
	req := SubprocessorNotificationRequest{
		RecipientInfo:  RecipientInfo{Email: "dolores@example.com", FirstName: "Dolores", UnsubscribeToken: "tok_dolores"},
		Subject:        "SecureCorp subprocessor update",
		Title:          "We've updated our subprocessors",
		Intros:         []string{"Our subprocessor list has changed."},
		Subprocessors:  sampleSubprocessorEntries(),
		ButtonText:     "View subprocessors",
		ButtonLink:     "https://securecorp.example.com/trust",
		UnsubscribeURL: "https://securecorp.example.com/unsubscribe?token={{ .unsubscribeToken }}",
		CompanyName:    "SecureCorp",
		LogoURL:        "https://securecorp.example.com/logo.png",
		PrimaryColor:   "#0f3d3a",
		ButtonColor:    "#3fc2b4",
	}

	html := renderSubprocessorMessage(t, req)

	// trust center branding overlay is applied (header logo + company name)
	assert.Contains(t, html, "SecureCorp")
	assert.Contains(t, html, "https://securecorp.example.com/logo.png")

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
// system branding when the trust center supplies no branding overlay values
func TestSubprocessorNotificationSystemBrandingFallback(t *testing.T) {
	cfg := MockRuntimeConfig()

	req := SubprocessorNotificationRequest{
		RecipientInfo: RecipientInfo{Email: "dolores@example.com", FirstName: "Dolores"},
		Subject:       "Subprocessor update",
		Title:         "We've updated our subprocessors",
		Subprocessors: sampleSubprocessorEntries(),
	}

	html := renderSubprocessorMessage(t, req)

	// no trust center company name leaks in; the system config branding is used instead
	assert.NotContains(t, html, "SecureCorp")
	assert.Contains(t, html, cfg.CompanyName)
}
