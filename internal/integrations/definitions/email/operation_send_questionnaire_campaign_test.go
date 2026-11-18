package email

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQuestionnairePayloadDropsSystemKeysFromDefaults(t *testing.T) {
	dispatch := questionnaireDispatch{
		defaults: map[string]any{
			"recipients":      []any{"someone@else.com"},
			"email":           "someone@else.com",
			"auth_url":        "https://evil.example.com",
			"firstName":       "Mallory",
			"tags":            []any{map[string]any{"name": "x", "value": "y"}},
			"assessment_name": "Other",
			"campaignId":      "other-campaign",
			"title":           "Hi {{ .firstName }}",
		},
		overlay: CampaignContext{CampaignID: "camp-1"},
	}

	input := QuestionnaireAuthEmail{
		RecipientInfo:  RecipientInfo{Email: "vendor@example.com"},
		AssessmentName: "SOC 2 Vendor Review",
		AuthURL:        "https://q.example.com/auth",
	}

	payload, err := questionnairePayload(dispatch, input)
	assert.Assert(t, err)

	var req QuestionnaireAccessRequest
	assert.Assert(t, json.Unmarshal(payload, &req))

	assert.Check(t, is.Len(req.Recipients, 0))
	assert.Check(t, is.Len(req.Tags, 0))
	assert.Check(t, is.Equal("vendor@example.com", req.Email))
	assert.Check(t, is.Equal("", req.FirstName))
	assert.Check(t, is.Equal("https://q.example.com/auth", req.AuthURL))
	assert.Check(t, is.Equal("SOC 2 Vendor Review", req.AssessmentName))
	assert.Check(t, is.Equal("camp-1", req.CampaignID))
	assert.Check(t, is.Equal("Hi {{ .firstName }}", req.Title))
}
