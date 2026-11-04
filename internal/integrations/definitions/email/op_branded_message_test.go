package email

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

const brandedMessageKey = "BrandedMessageRequest"

// TestBrandedMessageButtonRendered verifies the CTA action is built only when both the
// button text and link are provided
func TestBrandedMessageButtonRendered(t *testing.T) {
	op := testDispatcher[BrandedMessageRequest](t, brandedMessageKey)

	withButton := op.Build(RuntimeEmailConfig{}, BrandedMessageRequest{
		Title:      "Hello",
		ButtonText: "Go",
		ButtonLink: "https://example.com",
	})
	assert.Check(t, is.Len(withButton.Actions, 1))
	assert.Check(t, is.Equal("Go", withButton.Actions[0].Button.Text))
	assert.Check(t, is.Equal("https://example.com", withButton.Actions[0].Button.Link))

	noLink := op.Build(RuntimeEmailConfig{}, BrandedMessageRequest{
		Title:      "Hello",
		ButtonText: "Go",
	})
	assert.Check(t, is.Len(noLink.Actions, 0))
}

// TestBrandedMessageConfigAppliesRequestBranding verifies customer-supplied identity and footer
// fields are sourced from the request and override the inbound config (Option 1)
func TestBrandedMessageConfigAppliesRequestBranding(t *testing.T) {
	op := testDispatcher[BrandedMessageRequest](t, brandedMessageKey)

	req := BrandedMessageRequest{
		CompanyName:    "Acme",
		Corporation:    "Acme, Inc.",
		CompanyAddress: "1 Market St",
		Copyright:      "(c) Acme",
		Tagline:        "Tag",
		TermsURL:       "https://acme.com/terms",
		PrivacyURL:     "https://acme.com/privacy",
		UnsubscribeURL: "https://acme.com/unsub",
		LogoURL:        "https://acme.com/logo.png",
		HeaderLogoURL:  "https://acme.com/icon.png",
		Social:         []SocialLink{{Platform: "X", IconURL: "https://acme.com/x.png", URL: "https://x.com/acme"}},
	}

	// MockRuntimeConfig carries inherited (non-customer) branding that must be replaced
	out := op.Config(*MockRuntimeConfig(), req)

	assert.Check(t, is.Equal("Acme", out.CompanyName))
	assert.Check(t, is.Equal("Acme, Inc.", out.Corporation))
	assert.Check(t, is.Equal("1 Market St", out.CompanyAddress))
	assert.Check(t, is.Equal("(c) Acme", out.Copyright))
	assert.Check(t, is.Equal("Tag", out.Tagline))
	assert.Check(t, is.Equal("https://acme.com/terms", out.TermsURL))
	assert.Check(t, is.Equal("https://acme.com/privacy", out.PrivacyURL))
	assert.Check(t, is.Equal("https://acme.com/unsub", out.UnsubscribeURL))
	assert.Check(t, is.Equal("https://acme.com/logo.png", out.LogoURL))
	assert.Check(t, is.Equal("https://acme.com/icon.png", out.HeaderLogoURL))
	assert.Check(t, is.Len(out.Social, 1))
	assert.Check(t, is.Equal("X", out.Social[0].Platform))
}

// TestBrandedMessageConfigStripsInheritedBranding verifies an empty request clears inherited
// branding so system defaults never leak into a customer message, while functional delivery
// fields are preserved
func TestBrandedMessageConfigStripsInheritedBranding(t *testing.T) {
	op := testDispatcher[BrandedMessageRequest](t, brandedMessageKey)

	base := *MockRuntimeConfig()
	out := op.Config(base, BrandedMessageRequest{})

	assert.Check(t, is.Len(out.CompanyName, 0))
	assert.Check(t, is.Len(out.Corporation, 0))
	assert.Check(t, is.Len(out.CompanyAddress, 0))
	assert.Check(t, is.Len(out.LogoURL, 0))
	assert.Check(t, is.Len(out.HeaderLogoURL, 0))
	assert.Check(t, is.Len(out.Tagline, 0))
	assert.Check(t, is.Len(out.Social, 0))
	assert.Check(t, is.Len(out.HeaderText, 0))

	assert.Check(t, is.Equal(base.FromEmail, out.FromEmail))
	assert.Check(t, is.Equal(base.ProductURL, out.ProductURL))
}

// TestBrandedMessageCustomerSelectable verifies the entry is exposed via the customer-facing catalog
func TestBrandedMessageCustomerSelectable(t *testing.T) {
	d, ok := DispatcherByKey(brandedMessageKey)
	assert.Check(t, ok)

	cs := d.Registration().CustomerSelectable
	assert.Check(t, cs != nil)
	assert.Check(t, *cs)
}

// TestBrandedMessageExamplePayloadValid verifies the catalog example unmarshals and populates the
// fields needed to render a meaningful preview
func TestBrandedMessageExamplePayloadValid(t *testing.T) {
	d, ok := DispatcherByKey(brandedMessageKey)
	assert.Check(t, ok)

	raw := d.ExamplePayload()
	assert.Check(t, len(raw) > 0)

	var req BrandedMessageRequest
	assert.Check(t, json.Unmarshal(raw, &req))

	assert.Check(t, len(req.Email) > 0)
	assert.Check(t, len(req.Subject) > 0)
	assert.Check(t, len(req.Title) > 0)
	assert.Check(t, len(req.CompanyName) > 0)
}

// TestBrandedMessageUISchemaValid verifies the builder UI schema is a JSON object carrying the
// authoring order used by the form
func TestBrandedMessageUISchemaValid(t *testing.T) {
	d, ok := DispatcherByKey(brandedMessageKey)
	assert.Check(t, ok)

	raw := d.BuilderUISchema()
	assert.Check(t, len(raw) > 0)

	var ui map[string]any
	assert.Check(t, json.Unmarshal(raw, &ui))
	// only the irreducible multi-line hint remains; field types/widgets come from the reflected schema
	assert.Check(t, is.Contains(ui, "intros"))
}

// TestBrandedMessageConfigSchemaScoping verifies the reflected catalog schema excludes per-send
// fields and carries the format hints that replace the hand-written UI schema
func TestBrandedMessageConfigSchemaScoping(t *testing.T) {
	d, ok := DispatcherByKey(brandedMessageKey)
	assert.Check(t, ok)

	var schema map[string]any
	assert.Check(t, json.Unmarshal(d.Registration().Input.Schema, &schema))

	defs, ok := schema["$defs"].(map[string]any)
	assert.Check(t, ok, "schema should have $defs")

	entry, ok := defs["BrandedMessageRequest"].(map[string]any)
	assert.Check(t, ok, "schema should define BrandedMessageRequest")

	props, ok := entry["properties"].(map[string]any)
	assert.Check(t, ok, "BrandedMessageRequest should have properties")

	// per-send fields promoted from RecipientInfo / CampaignContext are excluded from authoring
	for _, perSend := range []string{"email", "recipients", "firstName", "lastName", "tags", "campaignId", "campaignName"} {
		_, found := props[perSend]
		assert.Check(t, !found, "per-send field %q should be excluded from the authorable schema", perSend)
	}

	// authorable content/branding fields remain
	for _, authorable := range []string{"subject", "title", "intros", "buttonText", "companyName"} {
		assert.Check(t, is.Contains(props, authorable), "authorable field %q should be present", authorable)
	}

	// format hints are reflected so the UI selects widgets from the schema, not a hand-written uiSchema
	assertSchemaFormat(t, props, "primaryColor", "color")
	assertSchemaFormat(t, props, "buttonColor", "color")
	assertSchemaFormat(t, props, "buttonLink", "uri")
	assertSchemaFormat(t, props, "logoURL", "uri")
}

// assertSchemaFormat asserts a reflected property carries the expected JSON Schema format
func assertSchemaFormat(t *testing.T, props map[string]any, field, want string) {
	t.Helper()

	p, ok := props[field].(map[string]any)
	assert.Check(t, ok, "field %q missing from schema", field)
	assert.Check(t, is.Equal(want, p["format"]), "field %q format", field)
}

// TestRenderCatalogPreviewExample verifies the catalog preview renders the example to non-empty HTML,
// interpolates the recipient first name, and applies the example branding instead of leaking the
// inbound config's branding
func TestRenderCatalogPreviewExample(t *testing.T) {
	d, ok := DispatcherByKey(brandedMessageKey)
	assert.Check(t, ok)

	client := &Client{Config: *MockRuntimeConfig()}

	html, err := RenderCatalogPreview(context.Background(), client, d, nil)
	assert.Check(t, err)
	assert.Check(t, len(html) > 0)

	assert.Check(t, is.Contains(html, "Hi Jordan, welcome aboard"))
	assert.Check(t, is.Contains(html, "View Onboarding"))
	assert.Check(t, !strings.Contains(html, "MITB"))
}

// TestRenderCatalogPreviewDraftOverridesExample verifies in-progress draft values override the
// example so the live preview reflects what the author has typed
func TestRenderCatalogPreviewDraftOverridesExample(t *testing.T) {
	d, ok := DispatcherByKey(brandedMessageKey)
	assert.Check(t, ok)

	client := &Client{Config: *MockRuntimeConfig()}

	draft := map[string]any{
		"title":      "Quarterly Update",
		"buttonText": "Read More",
	}

	html, err := RenderCatalogPreview(context.Background(), client, d, draft)
	assert.Check(t, err)

	assert.Check(t, is.Contains(html, "Quarterly Update"))
	assert.Check(t, is.Contains(html, "Read More"))
	assert.Check(t, !strings.Contains(html, "welcome aboard"))
}
