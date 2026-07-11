package email

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/ent/generated"
)

// TestRenderTrustCenterCampaignMessages verifies a trust center update renders with branding pulled
// from the trust center setting, the post content from the campaign metadata, and a tokenized
// per-recipient unsubscribe link
func TestRenderTrustCenterCampaignMessages(t *testing.T) {
	client := &Client{Config: *MockRuntimeConfig()}

	setting := TrustCenterSettingFixture()

	// resolve the dispatcher from the trust center update key, mirroring the campaign dispatch path
	dispatcher, ok := DispatcherByKey(TrustCenterUpdateTemplate)
	assert.True(t, ok)

	// the campaign metadata carries the post data, as the automated triggers supply it
	metadata := map[string]any{
		"postTitle":      "SOC 2 report published",
		"postText":       "We've updated our subprocessor list.",
		"trustCenterURL": "https://securecorp.example.com/trust",
		"unsubscribeURL": "https://securecorp.example.com/unsubscribe?token={{ .unsubscribeToken }}",
	}

	targets := []*generated.CampaignTarget{
		{
			ID:       "ct_1",
			Email:    "dolores@example.com",
			FullName: "Dolores Abernathy",
			Metadata: map[string]any{MetadataUnsubscribeTokenKey: "tok_dolores"},
		},
	}

	messages, targetIDs, failed := renderTrustCenterCampaignMessages(
		context.Background(), client, dispatcher, setting, metadata, CampaignContext{CampaignName: "June Update"}, targets,
	)

	assert.Equal(t, 0, failed)
	assert.Len(t, messages, 1)
	assert.Equal(t, []string{"ct_1"}, targetIDs)

	assert.Equal(t, "SecureCorp Trust Center Update: SOC 2 report published", messages[0].GetSubject())

	html := messages[0].GetHTML()
	assert.Contains(t, html, "SecureCorp has published a new update to their trust center.") // operation-composed attribution line
	assert.Contains(t, html, "SOC 2 report published")                                       // post title as the headline
	assert.Contains(t, html, "<p>We&#39;ve updated our subprocessor list.</p>")              // post body, escaped paragraph
	assert.Contains(t, html, "https://securecorp.example.com/logo.png")                      // trust center logo branding
	assert.Contains(t, html, "https://securecorp.example.com/unsubscribe?token=tok_dolores") // per-recipient unsubscribe link
}

// TestRenderTrustCenterCampaignMessagesNilSetting verifies a trust center with no setting renders
// with the generic copy and the system config branding
func TestRenderTrustCenterCampaignMessagesNilSetting(t *testing.T) {
	cfg := MockRuntimeConfig()
	client := &Client{Config: *cfg}

	dispatcher, ok := DispatcherByKey(TrustCenterUpdateTemplate)
	assert.True(t, ok)

	metadata := map[string]any{
		"postText": "We published a new update.",
	}

	targets := []*generated.CampaignTarget{
		{ID: "ct_1", Email: "dolores@example.com"},
	}

	messages, _, failed := renderTrustCenterCampaignMessages(
		context.Background(), client, dispatcher, nil, metadata, CampaignContext{}, targets,
	)

	assert.Equal(t, 0, failed)
	assert.Len(t, messages, 1)

	html := messages[0].GetHTML()
	assert.Contains(t, html, "A new update was published to the trust center.")
	assert.Contains(t, html, "We published a new update.")
	assert.Contains(t, html, cfg.CompanyName)
}
