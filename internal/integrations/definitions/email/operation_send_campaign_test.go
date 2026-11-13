package email

import (
	"encoding/json"
	"errors"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/integrations/templatekit"
)

// TestUnsubscribeTokenFromMetadata verifies the per-recipient unsubscribe token is read from
// campaign target metadata only when present and string-typed
func TestUnsubscribeTokenFromMetadata(t *testing.T) {
	assert.Check(t, is.Equal("tok_abc", unsubscribeTokenFromMetadata(map[string]any{MetadataUnsubscribeTokenKey: "tok_abc"})))
	assert.Check(t, is.Len(unsubscribeTokenFromMetadata(map[string]any{}), 0))
	assert.Check(t, is.Len(unsubscribeTokenFromMetadata(nil), 0))
	assert.Check(t, is.Len(unsubscribeTokenFromMetadata(map[string]any{MetadataUnsubscribeTokenKey: 123}), 0))
}

// TestUnsubscribeTokenInterpolatedIntoPayload verifies the per-recipient unsubscribe token is an
// available template variable, so a template's unsubscribeURL default resolves to a unique link
// for each recipient at render time
func TestUnsubscribeTokenInterpolatedIntoPayload(t *testing.T) {
	client := &Client{Config: *MockRuntimeConfig()}

	defaults := map[string]any{
		"unsubscribeURL": "https://tc.example.com/unsubscribe?token={{ .unsubscribeToken }}",
	}

	payload, err := templatekit.BuildDispatchPayload(defaults, RecipientInfo{
		Email:            "person@example.com",
		UnsubscribeToken: "tok_abc123",
	})
	assert.Check(t, err)

	resolved, err := interpolatePayload(client, payload, nil)
	assert.Check(t, err)

	var req BrandedMessageRequest
	assert.Check(t, json.Unmarshal(resolved, &req))
	assert.Check(t, is.Equal("https://tc.example.com/unsubscribe?token=tok_abc123", req.UnsubscribeURL))
}

func TestSplitFullName_FirstAndLast(t *testing.T) {
	first, last := splitFullName("Alice Smith")

	assert.Check(t, is.Equal("Alice", first))
	assert.Check(t, is.Equal("Smith", last))
}

func TestSplitFullName_FirstOnly(t *testing.T) {
	first, last := splitFullName("Alice")

	assert.Check(t, is.Equal("Alice", first))
	assert.Check(t, is.Equal("", last))
}

func TestSplitFullName_Empty(t *testing.T) {
	first, last := splitFullName("")

	assert.Check(t, is.Equal("", first))
	assert.Check(t, is.Equal("", last))
}

func TestSplitFullName_Whitespace(t *testing.T) {
	first, last := splitFullName("   ")

	assert.Check(t, is.Equal("", first))
	assert.Check(t, is.Equal("", last))
}

func TestSplitFullName_LeadingTrailingSpaces(t *testing.T) {
	first, last := splitFullName("  Alice Smith  ")

	assert.Check(t, is.Equal("Alice", first))
	assert.Check(t, is.Equal("Smith", last))
}

func TestSplitFullName_MultipleNames(t *testing.T) {
	// Only splits on the first space; "Mary Smith" stays in last
	first, last := splitFullName("Jane Mary Smith")

	assert.Check(t, is.Equal("Jane", first))
	assert.Check(t, is.Equal("Mary Smith", last))
}

func TestSplitFullName_Unicode(t *testing.T) {
	first, last := splitFullName("José García")

	assert.Check(t, is.Equal("José", first))
	assert.Check(t, is.Equal("García", last))
}

func TestValidateCampaignTemplateKey(t *testing.T) {
	testCases := []struct {
		name         string
		campaignType enums.CampaignType
		key          string
		wantErr      bool
	}{
		{name: "questionnaire campaign with access template", campaignType: enums.CampaignTypeQuestionnaire, key: questionnaireAccessKey},
		{name: "questionnaire campaign with branded template", campaignType: enums.CampaignTypeQuestionnaire, key: brandedMessageKey, wantErr: true},
		{name: "custom campaign with branded template", campaignType: enums.CampaignTypeCustom, key: brandedMessageKey},
		{name: "custom campaign with access template", campaignType: enums.CampaignTypeCustom, key: questionnaireAccessKey, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateCampaignTemplateKey(tc.campaignType, tc.key)

			assert.Check(t, is.Equal(tc.wantErr, errors.Is(err, ErrCampaignTemplateKeyMismatch)))
		})
	}
}
