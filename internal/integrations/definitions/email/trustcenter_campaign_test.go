package email

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/ent/generated"
)

// TestRenderTrustCenterCampaignMessages verifies a trust center update renders with branding pulled
// from the trust center setting, per-recipient content interpolation, and a tokenized unsubscribe link
func TestRenderTrustCenterCampaignMessages(t *testing.T) {
	dispatcher, ok := DispatcherByKey(brandedMessageKey)
	assert.True(t, ok)

	client := &Client{Config: *MockRuntimeConfig()}

	setting := TrustCenterSettingFixture()
	template := TrustCenterUpdateTemplateFixture()

	targets := []*generated.CampaignTarget{
		{
			ID:       "ct_1",
			Email:    "dolores@example.com",
			FullName: "Dolores Abernathy",
			Metadata: map[string]any{MetadataUnsubscribeTokenKey: "tok_dolores"},
		},
	}

	messages, targetIDs, failed := renderTrustCenterCampaignMessages(
		context.Background(), client, dispatcher, template, setting, nil, CampaignContext{CampaignName: "June Update"}, targets,
	)

	assert.Equal(t, 0, failed)
	assert.Len(t, messages, 1)
	assert.Equal(t, []string{"ct_1"}, targetIDs)

	html := messages[0].GetHTML()
	assert.Contains(t, html, "SecureCorp")                                                   // trust center company branding
	assert.Contains(t, html, "https://securecorp.example.com/logo.png")                      // trust center logo
	assert.Contains(t, html, "Hi Dolores")                                                   // per-recipient content interpolation
	assert.Contains(t, html, "https://securecorp.example.com/unsubscribe?token=tok_dolores") // per-recipient unsubscribe link
}

// TestApplyTrustCenterBrandingFallsBackToConfig verifies that a nil/empty trust center setting
// falls back to the runtime email config as the standard branding
func TestApplyTrustCenterBrandingFallsBackToConfig(t *testing.T) {
	cfg := *MockRuntimeConfig()

	var req BrandedMessageRequest
	applyTrustCenterBranding(&req, nil, cfg)

	assert.Equal(t, cfg.CompanyName, req.CompanyName)
	assert.Equal(t, cfg.LogoURL, req.LogoURL)
	assert.Equal(t, cfg.ButtonColor, req.ButtonColor)
}
