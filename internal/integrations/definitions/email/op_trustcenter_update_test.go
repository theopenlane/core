package email

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/theopenlane/newman"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// TestTrustCenterUpdateRegistered verifies the trust center update entry registers under its own
// unique identifier, distinct from the customer-selectable branded message key, so campaigns keyed
// by it resolve a dispatcher
func TestTrustCenterUpdateRegistered(t *testing.T) {
	assert.Check(t, BrandedMessageOp.Name() != TrustCenterUpdateTemplate)

	_, ok := DispatcherByKey(TrustCenterUpdateTemplate)
	assert.Check(t, ok)
}

// TestTrustCenterUpdateNotCustomerSelectable verifies the entry is excluded from the customer-facing
// catalog: the trust center update message is a system send, never authored from the picker
func TestTrustCenterUpdateNotCustomerSelectable(t *testing.T) {
	d, ok := DispatcherByKey(TrustCenterUpdateTemplate)
	assert.Check(t, ok)

	cs := d.Registration().CustomerSelectable
	assert.Check(t, cs != nil)
	assert.Check(t, !(*cs))

	for _, sel := range CustomerSelectableDispatchers() {
		assert.Check(t, TrustCenterUpdateTemplate != sel.Name())
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
			assert.Check(t, is.Equal(tc.expected, trustCenterUpdateSubject(tc.companyName, tc.title)))
		})
	}
}

// renderTrustCenterUpdateMessage renders the trust center update through its dispatcher, exercising
// the same decode + interpolate + render path used for a real send
func renderTrustCenterUpdateMessage(t *testing.T, client *Client, req TrustCenterUpdateRequest) *newman.EmailMessage {
	t.Helper()

	dispatcher, ok := DispatcherByKey(TrustCenterUpdateTemplate)
	assert.Assert(t, ok)

	payload, err := json.Marshal(req)
	assert.NilError(t, err)

	msg, err := dispatcher.RenderMessage(context.Background(), client, payload)
	assert.NilError(t, err)

	return msg
}

// TestTrustCenterUpdateRender verifies the post notification leads the subject with the company, uses
// the post title as the headline with a single attribution line, renders the post body as written,
// applies the branding overlay, and links back to the trust center
func TestTrustCenterUpdateRender(t *testing.T) {
	req := TrustCenterUpdateRequest{
		Email: "dolores@example.com", FirstName: "Dolores", UnsubscribeToken: "tok_dolores",
		CompanyName:    "SecureCorp",
		LogoURL:        "https://securecorp.example.com/logo.png",
		AccentColor:    "#3fc2b4",
		PostTitle:      "SOC 2 report published",
		PostText:       "Our latest SOC 2 Type II report is now available.\nRequest access to review the full report.",
		TrustCenterURL: "https://securecorp.example.com/trust",
		UnsubscribeURL: "https://securecorp.example.com/unsubscribe?token={{ .unsubscribeToken }}",
	}

	msg := renderTrustCenterUpdateMessage(t, &Client{Config: *MockRuntimeConfig()}, req)

	assert.Check(t, is.Equal("SecureCorp Trust Center Update: SOC 2 report published", msg.GetSubject()))

	html := msg.GetHTML()

	// the post title is the headline with a single attribution line naming the company
	assert.Check(t, is.Contains(html, "SOC 2 report published"))
	assert.Check(t, is.Contains(html, "SecureCorp has published a new update to their trust center."))

	// plain-text post body renders as paragraphs
	assert.Check(t, is.Contains(html, "<p>Our latest SOC 2 Type II report is now available.</p>"))
	assert.Check(t, is.Contains(html, "<p>Request access to review the full report.</p>"))

	// identity and trust center link are applied, with the accent as a decorative border only and
	// the logo rendered exactly once (header slot only)
	assert.Check(t, is.Equal(1, strings.Count(html, "https://securecorp.example.com/logo.png")))
	assert.Check(t, is.Contains(html, `href="https://securecorp.example.com/trust"`))
	assert.Check(t, is.Contains(html, "border-top:4px solid #3fc2b4"))
	assert.Check(t, !strings.Contains(html, "background-color:#3fc2b4"))

	// the per-recipient unsubscribe link is resolved from the token
	assert.Check(t, is.Contains(html, "https://securecorp.example.com/unsubscribe?token=tok_dolores"))
}

// TestTrustCenterUpdateRenderHTMLPost verifies an HTML post body renders as authored (lists and
// emphasis preserved) with dangerous content stripped by the email scrubber
func TestTrustCenterUpdateRenderHTMLPost(t *testing.T) {
	req := TrustCenterUpdateRequest{
		Email:     "dolores@example.com",
		PostTitle: "June update",
		PostText:  `<p>We shipped:</p><ul><li>New <strong>SOC 2</strong> report</li><li>Refreshed policies</li></ul><script>alert("x")</script>`,
	}

	msg := renderTrustCenterUpdateMessage(t, &Client{Config: *MockRuntimeConfig()}, req)

	html := msg.GetHTML()

	// authored formatting is preserved
	assert.Check(t, is.Contains(html, "<li>New <strong>SOC 2</strong> report</li>"))
	assert.Check(t, is.Contains(html, "<li>Refreshed policies</li>"))

	// dangerous content is stripped
	assert.Check(t, !strings.Contains(html, "<script>"))
}

// TestTrustCenterUpdateRenderFallbacks verifies a post with no title and no trust center branding
// renders the generic copy with the system branding
func TestTrustCenterUpdateRenderFallbacks(t *testing.T) {
	cfg := MockRuntimeConfig()

	req := TrustCenterUpdateRequest{
		Email:    "dolores@example.com",
		PostText: "We published a new update.",
	}

	msg := renderTrustCenterUpdateMessage(t, &Client{Config: *cfg}, req)

	assert.Check(t, is.Equal(DefaultTrustCenterUpdateTitle, msg.GetSubject()))

	html := msg.GetHTML()
	assert.Check(t, is.Contains(html, DefaultTrustCenterUpdateTitle))
	assert.Check(t, is.Contains(html, "A new update was published to the trust center."))
	assert.Check(t, is.Contains(html, "We published a new update."))
	assert.Check(t, !strings.Contains(html, "SecureCorp"))
	assert.Check(t, is.Contains(html, cfg.CompanyName))
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
	assert.NilError(t, err)

	assert.Check(t, is.Equal("SOC 2 report published", content["postTitle"]))
	assert.Check(t, is.Equal("We published a new update.", content["postText"]))
	assert.Check(t, is.Equal("https://securecorp.example.com/trust", content["trustCenterURL"]))

	unsubscribeURL, ok := content["unsubscribeURL"].(string)
	assert.Check(t, ok)
	assert.Check(t, is.Contains(unsubscribeURL, "{{ .unsubscribeToken }}"))

	_, hasEmail := content["email"]
	assert.Check(t, !hasEmail)
}
