package email

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.Len(t, withButton.Actions, 1)
	assert.Equal(t, "Go", withButton.Actions[0].Button.Text)
	assert.Equal(t, "https://example.com", withButton.Actions[0].Button.Link)

	noLink := op.Build(RuntimeEmailConfig{}, BrandedMessageRequest{
		Title:      "Hello",
		ButtonText: "Go",
	})
	assert.Empty(t, noLink.Actions)
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

	assert.Equal(t, "Acme", out.CompanyName)
	assert.Equal(t, "Acme, Inc.", out.Corporation)
	assert.Equal(t, "1 Market St", out.CompanyAddress)
	assert.Equal(t, "(c) Acme", out.Copyright)
	assert.Equal(t, "Tag", out.Tagline)
	assert.Equal(t, "https://acme.com/terms", out.TermsURL)
	assert.Equal(t, "https://acme.com/privacy", out.PrivacyURL)
	assert.Equal(t, "https://acme.com/unsub", out.UnsubscribeURL)
	assert.Equal(t, "https://acme.com/logo.png", out.LogoURL)
	assert.Equal(t, "https://acme.com/icon.png", out.HeaderLogoURL)
	require.Len(t, out.Social, 1)
	assert.Equal(t, "X", out.Social[0].Platform)
}

// TestBrandedMessageConfigStripsInheritedBranding verifies an empty request clears inherited
// branding so system defaults never leak into a customer message, while functional delivery
// fields are preserved
func TestBrandedMessageConfigStripsInheritedBranding(t *testing.T) {
	op := testDispatcher[BrandedMessageRequest](t, brandedMessageKey)

	base := *MockRuntimeConfig()
	out := op.Config(base, BrandedMessageRequest{})

	assert.Empty(t, out.CompanyName)
	assert.Empty(t, out.Corporation)
	assert.Empty(t, out.CompanyAddress)
	assert.Empty(t, out.LogoURL)
	assert.Empty(t, out.HeaderLogoURL)
	assert.Empty(t, out.Tagline)
	assert.Empty(t, out.Social)
	assert.Empty(t, out.HeaderText)

	assert.Equal(t, base.FromEmail, out.FromEmail)
	assert.Equal(t, base.ProductURL, out.ProductURL)
}

// TestBrandedMessageCustomerSelectable verifies the entry is exposed via the customer-facing catalog
func TestBrandedMessageCustomerSelectable(t *testing.T) {
	d, ok := DispatcherByKey(brandedMessageKey)
	require.True(t, ok)

	cs := d.Registration().CustomerSelectable
	require.NotNil(t, cs)
	assert.True(t, *cs)
}

// TestBrandedMessageExamplePayloadValid verifies the catalog example unmarshals and populates the
// fields needed to render a meaningful preview
func TestBrandedMessageExamplePayloadValid(t *testing.T) {
	d, ok := DispatcherByKey(brandedMessageKey)
	require.True(t, ok)

	raw := d.ExamplePayload()
	require.NotEmpty(t, raw)

	var req BrandedMessageRequest
	require.NoError(t, json.Unmarshal(raw, &req))

	assert.NotEmpty(t, req.Email)
	assert.NotEmpty(t, req.Subject)
	assert.NotEmpty(t, req.Title)
	assert.NotEmpty(t, req.CompanyName)
}

// TestBrandedMessageUISchemaValid verifies the builder UI schema is a JSON object carrying the
// authoring order used by the form
func TestBrandedMessageUISchemaValid(t *testing.T) {
	d, ok := DispatcherByKey(brandedMessageKey)
	require.True(t, ok)

	raw := d.BuilderUISchema()
	require.NotEmpty(t, raw)

	var ui map[string]any
	require.NoError(t, json.Unmarshal(raw, &ui))
	// only the irreducible multi-line hint remains; field types/widgets come from the reflected schema
	assert.Contains(t, ui, "intros")
}

// TestBrandedMessageConfigSchemaScoping verifies the reflected catalog schema excludes per-send
// fields and carries the format hints that replace the hand-written UI schema
func TestBrandedMessageConfigSchemaScoping(t *testing.T) {
	d, ok := DispatcherByKey(brandedMessageKey)
	require.True(t, ok)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(d.Registration().ConfigSchema, &schema))

	defs, ok := schema["$defs"].(map[string]any)
	require.True(t, ok, "schema should have $defs")

	entry, ok := defs["BrandedMessageRequest"].(map[string]any)
	require.True(t, ok, "schema should define BrandedMessageRequest")

	props, ok := entry["properties"].(map[string]any)
	require.True(t, ok, "BrandedMessageRequest should have properties")

	// per-send fields promoted from RecipientInfo / CampaignContext are excluded from authoring
	for _, perSend := range []string{"email", "recipients", "firstName", "lastName", "tags", "campaignId", "campaignName"} {
		assert.NotContains(t, props, perSend, "per-send field %q should be excluded from the authorable schema", perSend)
	}

	// authorable content/branding fields remain
	for _, authorable := range []string{"subject", "title", "intros", "buttonText", "companyName"} {
		assert.Contains(t, props, authorable, "authorable field %q should be present", authorable)
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
	require.True(t, ok, "field %q missing from schema", field)
	assert.Equal(t, want, p["format"], "field %q format", field)
}

// TestRenderCatalogPreviewExample verifies the catalog preview renders the example to non-empty HTML,
// interpolates the recipient first name, and applies the example branding instead of leaking the
// inbound config's branding
func TestRenderCatalogPreviewExample(t *testing.T) {
	d, ok := DispatcherByKey(brandedMessageKey)
	require.True(t, ok)

	client := &Client{Config: *MockRuntimeConfig()}

	html, err := RenderCatalogPreview(context.Background(), client, d, nil)
	require.NoError(t, err)
	require.NotEmpty(t, html)

	assert.Contains(t, html, "Hi Jordan, welcome aboard")
	assert.Contains(t, html, "View Onboarding")
	assert.NotContains(t, html, "MITB")
}

// TestRenderCatalogPreviewDraftOverridesExample verifies in-progress draft values override the
// example so the live preview reflects what the author has typed
func TestRenderCatalogPreviewDraftOverridesExample(t *testing.T) {
	d, ok := DispatcherByKey(brandedMessageKey)
	require.True(t, ok)

	client := &Client{Config: *MockRuntimeConfig()}

	draft := map[string]any{
		"title":      "Quarterly Update",
		"buttonText": "Read More",
	}

	html, err := RenderCatalogPreview(context.Background(), client, d, draft)
	require.NoError(t, err)

	assert.Contains(t, html, "Quarterly Update")
	assert.Contains(t, html, "Read More")
	assert.NotContains(t, html, "welcome aboard")
}
