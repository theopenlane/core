package email

import (
	"encoding/json"

	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// QuestionnaireEmailContent holds the customer-authorable text and branding of the questionnaire access email
type QuestionnaireEmailContent struct {
	MessageContent
	BrandOverrides
}

// defaultQuestionnaireContent returns the default questionnaire wording for the org and questionnaire
func defaultQuestionnaireContent(req QuestionnaireAuthEmail) QuestionnaireEmailContent {
	return QuestionnaireEmailContent{
		Subject:    "Access " + req.AssessmentName + " Questionnaire from " + req.OrgName,
		Preheader:  req.AssessmentName + " — complete your questionnaire from " + req.OrgName,
		Title:      req.OrgName + " sent you a questionnaire to complete",
		Intros:     []string{req.OrgName + " has shared the " + req.AssessmentName + " questionnaire for you to complete."},
		ButtonText: "Access Questionnaire",
	}
}

// QuestionnaireAccessRequest is the customer-selectable catalog entry for questionnaire campaign emails
// the access link, recipient and questionnaire fields are set per send and excluded from the authorable schema
type QuestionnaireAccessRequest struct {
	QuestionnaireAuthEmail `jsonschema:"-"`
	CampaignContext        `jsonschema:"-"`
	QuestionnaireEmailContent
}

// QuestionnaireAccessOp is the typed operation ref used for catalog dispatch
var QuestionnaireAccessOp = types.OperationPayloadOf[QuestionnaireAccessRequest]()

// questionnaireAccessExample seeds the editor and preview with the default questionnaire wording
var questionnaireAccessExample = QuestionnaireAccessRequest{
	QuestionnaireAuthEmail: QuestionnaireAuthEmail{
		RecipientInfo:  RecipientInfo{Email: "jordan.avery@example.com", FirstName: "Jordan", LastName: "Avery"},
		OrgName:        "Acme Security",
		AssessmentName: "Vendor Security Review",
		AuthURL:        "https://example.com/questionnaire",
	},
	CampaignContext:           CampaignContext{CampaignName: "Vendor Security Review"},
	QuestionnaireEmailContent: defaultQuestionnaireContent(QuestionnaireAuthEmail{OrgName: "{{ .orgName }}", AssessmentName: "{{ .questionnaireName }}"}),
}

// questionnaireAccessUISchema carries only UI hints that JSON Schema cannot express
var questionnaireAccessUISchema = json.RawMessage(`{
  "intros": {"items": {"ui:widget": "textarea"}}
}`)

var questionnaireAccess = RegisterEmailOperation(Operation[QuestionnaireAccessRequest]{
	Op:                 QuestionnaireAccessOp,
	Theme:              baseTheme,
	Description:        "Questionnaire access email with customizable text, logos, and colors; the questionnaire access link is always added by the system",
	CustomerSelectable: new(true),
	Example:            questionnaireAccessExample,
	UISchema:           questionnaireAccessUISchema,
	Variables: []OperationVariable{
		{Name: "orgName", Description: "Name of the organization sending the questionnaire", PayloadKey: "org_name"},
		{Name: "questionnaireName", Description: "Name of the questionnaire being sent", PayloadKey: "assessment_name"},
	},
	SystemKeys: payloadKeys[QuestionnaireAuthEmail](),
	Subject: func(_ RuntimeEmailConfig, req QuestionnaireAccessRequest) string {
		return req.Subject
	},
	Build: func(cfg RuntimeEmailConfig, req QuestionnaireAccessRequest) render.ContentBody {
		return questionnaireBody(customQuestionnaireConfig(cfg, req.QuestionnaireEmailContent), req.QuestionnaireAuthEmail, req.QuestionnaireEmailContent)
	},
	Config: func(cfg RuntimeEmailConfig, req QuestionnaireAccessRequest) RuntimeEmailConfig {
		return customQuestionnaireConfig(cfg, req.QuestionnaireEmailContent)
	},
	MessageOptions: func(cfg RuntimeEmailConfig, _ QuestionnaireAccessRequest) []newman.MessageOption {
		return questionnaireFromOption(cfg)
	},
})

// questionnaireBody builds the questionnaire access email body from the content, always linking the button to the auth URL
func questionnaireBody(cfg RuntimeEmailConfig, req QuestionnaireAuthEmail, content QuestionnaireEmailContent) render.ContentBody {
	cfg = questionnaireConfig(cfg, content)

	return render.ContentBody{
		Preheader: content.Preheader,
		Header:    defaultHeader(cfg),
		Title:     content.Title,
		Intros:    render.IntrosBlock{Paragraphs: content.Intros},
		Actions: []render.Action{{
			Button: render.Button{
				Text:      content.ButtonText,
				Link:      req.AuthURL,
				Color:     cfg.ButtonColor,
				TextColor: cfg.ButtonTextColor,
			},
		}},
	}
}

// questionnaireConfig applies the questionnaire button colors and then the authored brand overrides
func questionnaireConfig(cfg RuntimeEmailConfig, content QuestionnaireEmailContent) RuntimeEmailConfig {
	cfg.ButtonColor = tcButtonColor
	cfg.ButtonTextColor = tcButtonTextColor

	return content.apply(cfg)
}

// customQuestionnaireConfig clears the Openlane header logo so a custom template without one renders no header
func customQuestionnaireConfig(cfg RuntimeEmailConfig, content QuestionnaireEmailContent) RuntimeEmailConfig {
	cfg.HeaderLogoURL = ""

	return questionnaireConfig(cfg, content)
}

// questionnaireFromOption returns the questionnaire from address override when one is configured
func questionnaireFromOption(cfg RuntimeEmailConfig) []newman.MessageOption {
	if cfg.QuestionnaireEmail == "" {
		return nil
	}

	return []newman.MessageOption{
		newman.WithFrom(cfg.QuestionnaireEmail),
	}
}
