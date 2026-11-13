package email

import (
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

const questionnaireAccessKey = "QuestionnaireAccessRequest"

func TestQuestionnaireAccessCustomerSelectable(t *testing.T) {
	d, ok := DispatcherByKey(questionnaireAccessKey)
	assert.Assert(t, ok)

	cs := d.Registration().CustomerSelectable
	assert.Check(t, cs != nil && *cs)
}

func TestQuestionnaireAccessTemplateVariables(t *testing.T) {
	d, ok := DispatcherByKey(questionnaireAccessKey)
	assert.Assert(t, ok)

	names := make(map[string]struct{})
	for _, v := range d.TemplateVariables() {
		assert.Check(t, len(v.Description) > 0, "variable %s missing description", v.Name)
		names[v.Name] = struct{}{}
	}

	assert.Check(t, is.Contains(names, "orgName"))
	assert.Check(t, is.Contains(names, "questionnaireName"))

	branded, ok := DispatcherByKey(brandedMessageKey)
	assert.Assert(t, ok)
	assert.Check(t, is.Len(branded.TemplateVariables(), 0))
}

func TestQuestionnaireAccessConfigSchemaScoping(t *testing.T) {
	d, ok := DispatcherByKey(questionnaireAccessKey)
	assert.Assert(t, ok)

	var schema map[string]any
	assert.Check(t, json.Unmarshal(d.Registration().Input.Schema, &schema))

	defs, ok := schema["$defs"].(map[string]any)
	assert.Assert(t, ok)

	entry, ok := defs[questionnaireAccessKey].(map[string]any)
	assert.Assert(t, ok)

	props, ok := entry["properties"].(map[string]any)
	assert.Assert(t, ok)

	for _, perSend := range []string{"email", "recipients", "tags", "auth_url", "org_name", "assessment_name", "campaignId"} {
		_, found := props[perSend]
		assert.Check(t, !found, "per-send field %q should be excluded from the authorable schema", perSend)
	}

	for _, authorable := range []string{"subject", "preheader", "title", "intros", "buttonText", "logoURL", "headerLogoURL", "buttonColor", "primaryColor"} {
		assert.Check(t, is.Contains(props, authorable), "authorable field %q should be present", authorable)
	}

	for _, excluded := range []string{"buttonLink", "companyName", "social", "unsubscribeURL"} {
		_, found := props[excluded]
		assert.Check(t, !found, "field %q should not be authorable on the questionnaire access email", excluded)
	}
}

func TestQuestionnaireAccessPreviewExample(t *testing.T) {
	d, ok := DispatcherByKey(questionnaireAccessKey)
	assert.Assert(t, ok)

	html, err := RenderCatalogPreview(context.Background(), &Client{Config: *MockRuntimeConfig()}, d, nil)
	assert.Check(t, err)

	assert.Check(t, is.Contains(html, "Acme Security sent you a questionnaire to complete"))
	assert.Check(t, is.Contains(html, "Vendor Security Review questionnaire"))
	assert.Check(t, is.Contains(html, "https://example.com/questionnaire"))
}

func TestQuestionnaireAccessRenderUsesAuthoredContent(t *testing.T) {
	d, ok := DispatcherByKey(questionnaireAccessKey)
	assert.Assert(t, ok)

	payload, err := json.Marshal(QuestionnaireAccessRequest{
		QuestionnaireAuthEmail: QuestionnaireAuthEmail{
			RecipientInfo:  RecipientInfo{Email: "vendor@example.com", FirstName: "Sam"},
			OrgName:        "Acme",
			AssessmentName: "SOC 2 Vendor Review",
			AuthURL:        "https://q.example.com/auth?tok=abc",
		},
		QuestionnaireEmailContent: QuestionnaireEmailContent{
			Subject:        "{{ .orgName }} needs your {{ .questionnaireName }}",
			Title:          "Hi {{ .firstName }}",
			Intros:         []string{"Please finish by Friday."},
			ButtonText:     "Start now",
			BrandOverrides: BrandOverrides{ButtonColor: "#000000", ButtonTextColor: "#ffffff"},
		},
	})
	assert.Assert(t, err)

	msg, err := d.RenderMessage(context.Background(), &Client{Config: *MockRuntimeConfig()}, payload)
	assert.Assert(t, err)

	assert.Check(t, is.Equal("Acme needs your SOC 2 Vendor Review", msg.Subject))
	assert.Check(t, is.Contains(msg.GetHTML(), "Hi Sam"))
	assert.Check(t, is.Contains(msg.GetHTML(), "Please finish by Friday."))
	assert.Check(t, is.Contains(msg.GetHTML(), "Start now"))
	assert.Check(t, is.Contains(msg.GetHTML(), "https://q.example.com/auth?tok=abc"))
}

func TestQuestionnaireBodyDefaultsWhenContentEmpty(t *testing.T) {
	req := QuestionnaireAuthEmail{OrgName: "Acme", AssessmentName: "SOC 2 Vendor Review", AuthURL: "https://q.example.com/auth"}

	body := questionnaireBody(RuntimeEmailConfig{}, req, QuestionnaireEmailContent{})

	assert.Check(t, is.Equal("Acme sent you a questionnaire to complete", body.Title))
	assert.Check(t, is.DeepEqual([]string{"Acme has shared the SOC 2 Vendor Review questionnaire for you to complete."}, body.Intros.Paragraphs))
	assert.Assert(t, is.Len(body.Actions, 1))
	assert.Check(t, is.Equal(defaultQuestionnaireButtonText, body.Actions[0].Button.Text))
	assert.Check(t, is.Equal(tcButtonColor, body.Actions[0].Button.Color))
	assert.Check(t, is.Equal(tcButtonTextColor, body.Actions[0].Button.TextColor))
	assert.Check(t, is.Equal("Access SOC 2 Vendor Review Questionnaire from Acme", questionnaireSubject(req, QuestionnaireEmailContent{})))
}

func TestQuestionnaireBodyAlwaysLinksAuthURL(t *testing.T) {
	req := QuestionnaireAuthEmail{AuthURL: "https://q.example.com/auth"}

	body := questionnaireBody(RuntimeEmailConfig{}, req, QuestionnaireEmailContent{ButtonText: "Go"})

	assert.Assert(t, is.Len(body.Actions, 1))
	assert.Check(t, is.Equal("Go", body.Actions[0].Button.Text))
	assert.Check(t, is.Equal("https://q.example.com/auth", body.Actions[0].Button.Link))
}

func TestQuestionnaireBodyAppliesHeaderLogoOverride(t *testing.T) {
	cfg := RuntimeEmailConfig{HeaderLogoURL: "https://openlane.io/icon.png"}

	body := questionnaireBody(cfg, QuestionnaireAuthEmail{}, QuestionnaireEmailContent{BrandOverrides: BrandOverrides{HeaderLogoURL: "https://acme.com/icon.png"}})

	assert.Assert(t, body.Header.Logo != nil)
	assert.Check(t, is.Equal("https://acme.com/icon.png", body.Header.Logo.Src))
}
