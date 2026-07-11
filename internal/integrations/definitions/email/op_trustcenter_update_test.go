package email

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/newman"
)

// TestTrustCenterUpdateRegistered verifies the trust center update entry registers under its own
// unique identifier, distinct from the customer-selectable branded message key, so campaigns keyed
// by it resolve a dispatcher
func TestTrustCenterUpdateRegistered(t *testing.T) {
	assert.NotEqual(t, BrandedMessageOp.Name(), TrustCenterUpdateTemplate)

	_, ok := DispatcherByKey(TrustCenterUpdateTemplate)
	assert.True(t, ok)
}

// TestTrustCenterUpdateNotCustomerSelectable verifies the entry is excluded from the customer-facing
// catalog: the trust center update message is a system send, never authored from the picker
func TestTrustCenterUpdateNotCustomerSelectable(t *testing.T) {
	d, ok := DispatcherByKey(TrustCenterUpdateTemplate)
	assert.True(t, ok)

	cs := d.Registration().CustomerSelectable
	assert.NotNil(t, cs)
	assert.False(t, *cs)

	for _, sel := range CustomerSelectableDispatchers() {
		assert.NotEqual(t, TrustCenterUpdateTemplate, sel.Name())
	}
}

// TestTrustCenterUpdateSubject verifies the operation composes the subject from the trust center's
// company name and the post title, falling back to the generic update subject
func TestTrustCenterUpdateSubject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		companyName string
		title       string
		expected    string
	}{
		{
			name:        "company and title",
			companyName: "SecureCorp",
			title:       "SOC 2 report published",
			expected:    "SecureCorp Trust Center Update: SOC 2 report published",
		},
		{
			name:        "company only",
			companyName: "SecureCorp",
			expected:    "SecureCorp Trust Center Update",
		},
		{
			name:     "title only",
			title:    "SOC 2 report published",
			expected: "Trust center update: SOC 2 report published",
		},
		{
			name:     "neither",
			expected: "Trust center update",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, trustCenterUpdateSubject(tc.companyName, tc.title))
		})
	}
}

// renderTrustCenterUpdateMessage renders the trust center update through its dispatcher, exercising
// the same decode + interpolate + render path used for a real send
func renderTrustCenterUpdateMessage(t *testing.T, client *Client, req TrustCenterUpdateRequest) *newman.EmailMessage {
	t.Helper()

	dispatcher, ok := DispatcherByKey(TrustCenterUpdateTemplate)
	require.True(t, ok)

	payload, err := json.Marshal(req)
	require.NoError(t, err)

	msg, err := dispatcher.RenderMessage(context.Background(), client, payload)
	require.NoError(t, err)

	return msg
}

// TestTrustCenterUpdateRender verifies the post notification leads the subject with the company, uses
// the post title as the headline with a single attribution line, renders the post body as written,
// applies the branding overlay, and links back to the trust center
func TestTrustCenterUpdateRender(t *testing.T) {
	req := TrustCenterUpdateRequest{
		RecipientInfo: RecipientInfo{Email: "dolores@example.com", FirstName: "Dolores", UnsubscribeToken: "tok_dolores"},
		TrustCenterBranding: TrustCenterBranding{
			CompanyName: "SecureCorp",
			LogoURL:     "https://securecorp.example.com/logo.png",
			AccentColor: "#3fc2b4",
		},
		PostTitle:      "SOC 2 report published",
		PostText:       "Our latest SOC 2 Type II report is now available.\nRequest access to review the full report.",
		TrustCenterURL: "https://securecorp.example.com/trust",
		UnsubscribeURL: "https://securecorp.example.com/unsubscribe?token={{ .unsubscribeToken }}",
	}

	msg := renderTrustCenterUpdateMessage(t, &Client{Config: *MockRuntimeConfig()}, req)

	assert.Equal(t, "SecureCorp Trust Center Update: SOC 2 report published", msg.GetSubject())

	html := msg.GetHTML()

	// the post title is the headline with a single attribution line naming the company
	assert.Contains(t, html, "SOC 2 report published")
	assert.Contains(t, html, "SecureCorp has published a new update to their trust center.")

	// plain-text post body renders as paragraphs
	assert.Contains(t, html, "<p>Our latest SOC 2 Type II report is now available.</p>")
	assert.Contains(t, html, "<p>Request access to review the full report.</p>")

	// identity and trust center link are applied, with the accent as a decorative border only and
	// the logo rendered exactly once (header slot only)
	assert.Equal(t, 1, strings.Count(html, "https://securecorp.example.com/logo.png"))
	assert.Contains(t, html, `href="https://securecorp.example.com/trust"`)
	assert.Contains(t, html, "border-top:4px solid #3fc2b4")
	assert.NotContains(t, html, "background-color:#3fc2b4")

	// the per-recipient unsubscribe link is resolved from the token
	assert.Contains(t, html, "https://securecorp.example.com/unsubscribe?token=tok_dolores")
}

// TestTrustCenterUpdateRenderHTMLPost verifies an HTML post body renders as authored (lists and
// emphasis preserved) with dangerous content stripped by the email scrubber
func TestTrustCenterUpdateRenderHTMLPost(t *testing.T) {
	req := TrustCenterUpdateRequest{
		RecipientInfo: RecipientInfo{Email: "dolores@example.com"},
		PostTitle:     "June update",
		PostText:      `<p>We shipped:</p><ul><li>New <strong>SOC 2</strong> report</li><li>Refreshed policies</li></ul><script>alert("x")</script>`,
	}

	msg := renderTrustCenterUpdateMessage(t, &Client{Config: *MockRuntimeConfig()}, req)

	html := msg.GetHTML()

	// authored formatting is preserved
	assert.Contains(t, html, "<li>New <strong>SOC 2</strong> report</li>")
	assert.Contains(t, html, "<li>Refreshed policies</li>")

	// dangerous content is stripped
	assert.NotContains(t, html, "<script>")
}

// TestTrustCenterUpdateRenderFallbacks verifies a post with no title and no trust center branding
// renders the generic copy with the system branding
func TestTrustCenterUpdateRenderFallbacks(t *testing.T) {
	cfg := MockRuntimeConfig()

	req := TrustCenterUpdateRequest{
		RecipientInfo: RecipientInfo{Email: "dolores@example.com"},
		PostText:      "We published a new update.",
	}

	msg := renderTrustCenterUpdateMessage(t, &Client{Config: *cfg}, req)

	assert.Equal(t, DefaultTrustCenterUpdateTitle, msg.GetSubject())

	html := msg.GetHTML()
	assert.Contains(t, html, DefaultTrustCenterUpdateTitle)
	assert.Contains(t, html, "A new update was published to the trust center.")
	assert.Contains(t, html, "We published a new update.")
	assert.NotContains(t, html, "SecureCorp")
	assert.Contains(t, html, cfg.CompanyName)
}

// TestTrustCenterUpdateContent verifies the campaign metadata conversion carries the post data under
// the request's JSON keys and drops the empty per-target recipient email key
func TestTrustCenterUpdateContent(t *testing.T) {
	t.Parallel()

	content, err := TrustCenterUpdateContent(TrustCenterUpdateRequest{
		PostTitle:      "SOC 2 report published",
		PostText:       "We published a new update.",
		TrustCenterURL: "https://securecorp.example.com/trust",
		UnsubscribeURL: "https://securecorp.example.com/unsubscribe?token={{ .unsubscribeToken }}",
	})
	require.NoError(t, err)

	assert.Equal(t, "SOC 2 report published", content["postTitle"])
	assert.Equal(t, "We published a new update.", content["postText"])
	assert.Equal(t, "https://securecorp.example.com/trust", content["trustCenterURL"])

	unsubscribeURL, ok := content["unsubscribeURL"].(string)
	assert.True(t, ok)
	assert.Contains(t, unsubscribeURL, "{{ .unsubscribeToken }}")

	_, hasEmail := content["email"]
	assert.False(t, hasEmail)
}
