package email

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// renderSubprocessorMessage renders the subprocessor notification through its dispatcher and returns the
// resulting HTML body, exercising the same decode + interpolate + render path used for a real send
func renderSubprocessorMessage(t *testing.T, req SubprocessorNotificationRequest) string {
	t.Helper()

	dispatcher, ok := DispatcherByKey(SubprocessorNotificationOp.Name())
	assert.Assert(t, ok)

	client := &Client{Config: *MockRuntimeConfig()}

	payload, err := json.Marshal(req)
	assert.NilError(t, err)

	msg, err := dispatcher.RenderMessage(context.Background(), client, payload)
	assert.NilError(t, err)

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
	assert.Assert(t, ok)

	client := &Client{Config: *MockRuntimeConfig()}
	payload, err := json.Marshal(req)
	assert.NilError(t, err)

	msg, err := dispatcher.RenderMessage(context.Background(), client, payload)
	assert.NilError(t, err)

	assert.Check(t, is.Equal("<https://trust.openlane.io/api/unsubscribe?token=tok_abc123>", msg.Headers["List-Unsubscribe"]))
	assert.Check(t, is.Equal("List-Unsubscribe=One-Click", msg.Headers["List-Unsubscribe-Post"]))
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
	assert.Check(t, is.Equal("SecureCorp Trust Center Subprocessor Update", branded))

	generic := dispatcher.Subject(cfg, SubprocessorNotificationRequest{})
	assert.Check(t, is.Equal(subprocessorNotificationSubject, generic))
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
	assert.Check(t, is.Contains(html, "SecureCorp has updated its subprocessors"))

	// trust center identity is applied: company name, and the logo exactly once (header slot
	// only, no in-body hero logo)
	assert.Check(t, is.Contains(html, "SecureCorp"))
	assert.Check(t, is.Equal(1, strings.Count(html, "https://securecorp.example.com/logo.png")))

	// the accent surfaces only as decorative borders; containers and buttons keep the system styling
	assert.Check(t, is.Contains(html, "border-top:4px solid #3fc2b4"))
	assert.Check(t, is.Contains(html, "border-left:3px solid #3fc2b4"))
	assert.Check(t, is.Contains(html, "background-color:#14171e")) // the CTA button keeps the system color
	assert.Check(t, !strings.Contains(html, "background-color:#3fc2b4"))

	// platform marketing chrome is stripped from trust center subscriber emails
	assert.Check(t, !strings.Contains(html, MockRuntimeConfig().Tagline))
	assert.Check(t, !strings.Contains(html, MockRuntimeConfig().HeaderText))

	// the changed vendor list renders just names and change labels (no logo, link, or countries)
	assert.Check(t, is.Contains(html, "Amazon Web Services"))
	assert.Check(t, is.Contains(html, "Stripe"))
	assert.Check(t, is.Contains(html, "Twilio"))
	assert.Check(t, is.Contains(html, "Added"))
	assert.Check(t, is.Contains(html, "Updated"))
	assert.Check(t, is.Contains(html, "Removed"))

	// the link back to the trust center is the call-to-action button
	assert.Check(t, is.Contains(html, `href="https://securecorp.example.com/trust"`))

	// the per-recipient unsubscribe link is resolved from the token
	assert.Check(t, is.Contains(html, "https://securecorp.example.com/unsubscribe?token=tok_dolores"))
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
	assert.Check(t, !strings.Contains(html, "SecureCorp"))
	assert.Check(t, is.Contains(html, cfg.CompanyName))
	assert.Check(t, is.Contains(html, "updated our subprocessors"))
	assert.Check(t, !strings.Contains(html, "border-top:4px solid"))
}
