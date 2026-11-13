package email

import (
	"encoding/json"
	"fmt"

	"github.com/samber/lo"
	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// defaultQuestionnaireButtonText is the call-to-action label used when no custom button text is set
const defaultQuestionnaireButtonText = "Access Questionnaire"

// QuestionnaireEmailContent holds the customer-authorable text and branding of the questionnaire access email
type QuestionnaireEmailContent struct {
	// Subject is the email subject line
	Subject string `json:"subject,omitempty" jsonschema:"description=Email subject line"`
	// Preheader is hidden preview text shown in the inbox list
	Preheader string `json:"preheader,omitempty" jsonschema:"description=Inbox preview text"`
	// Title is the headline shown above the body paragraphs
	Title string `json:"title,omitempty" jsonschema:"description=Email headline"`
	// Intros are body paragraphs rendered before the questionnaire button
	Intros []string `json:"intros,omitempty" jsonschema:"description=Body paragraphs rendered before the questionnaire button"`
	// ButtonText is the questionnaire button label
	ButtonText string `json:"buttonText,omitempty" jsonschema:"description=Questionnaire button label"`
	BrandOverrides
}

// QuestionnaireInvitationRequest is the customer-selectable catalog entry for questionnaire campaign emails
// the access link, recipient and questionnaire fields are set per send and excluded from the authorable schema
type QuestionnaireInvitationRequest struct {
	QuestionnaireAuthEmail `jsonschema:"-"`
	CampaignContext        `jsonschema:"-"`
	QuestionnaireEmailContent
}

// QuestionnaireInvitationOp is the typed operation ref used for catalog dispatch
var QuestionnaireInvitationOp = types.OperationPayloadOf[QuestionnaireInvitationRequest]()

// questionnaireInvitationExample seeds the editor and preview with the default questionnaire wording
var questionnaireInvitationExample = QuestionnaireInvitationRequest{
	QuestionnaireAuthEmail: QuestionnaireAuthEmail{
		RecipientInfo:  RecipientInfo{Email: "jordan.avery@example.com", FirstName: "Jordan", LastName: "Avery"},
		OrgName:        "Acme Security",
		AssessmentName: "Vendor Security Review",
		AuthURL:        "https://example.com/questionnaire",
	},
	CampaignContext: CampaignContext{CampaignName: "Vendor Security Review"},
	Subject:         "Access {{ .questionnaireName }} Questionnaire from {{ .orgName }}",
	Preheader:       "{{ .questionnaireName }} — complete your questionnaire from {{ .orgName }}",
	Title:           "{{ .orgName }} sent you a questionnaire to complete",
	Intros:          []string{"{{ .orgName }} has shared the {{ .questionnaireName }} questionnaire for you to complete."},
	ButtonText:      defaultQuestionnaireButtonText,
}

// questionnaireInvitationUISchema carries only UI hints that JSON Schema cannot express
var questionnaireInvitationUISchema = json.RawMessage(`{
  "intros": {"items": {"ui:widget": "textarea"}}
}`)

var questionnaireInvitation = RegisterEmailOperation(Operation[QuestionnaireInvitationRequest]{
	Op:                 QuestionnaireInvitationOp,
	Theme:              baseTheme,
	Description:        "Questionnaire invitation with customizable text, logos, and colors; the questionnaire access link is always added by the system",
	CustomerSelectable: new(true),
	Example:            questionnaireInvitationExample,
	UISchema:           questionnaireInvitationUISchema,
	Variables: []OperationVariable[QuestionnaireInvitationRequest]{
		{
			Name:        "orgName",
			Description: "Name of the organization sending the questionnaire",
			Value:       func(req QuestionnaireInvitationRequest) string { return req.OrgName },
		},
		{
			Name:        "questionnaireName",
			Description: "Name of the questionnaire being sent",
			Value:       func(req QuestionnaireInvitationRequest) string { return req.AssessmentName },
		},
	},
	Subject: func(_ RuntimeEmailConfig, req QuestionnaireInvitationRequest) string {
		return questionnaireSubject(req.QuestionnaireAuthEmail, req.QuestionnaireEmailContent)
	},
	Build: func(cfg RuntimeEmailConfig, req QuestionnaireInvitationRequest) render.ContentBody {
		return questionnaireBody(cfg, req.QuestionnaireAuthEmail, req.QuestionnaireEmailContent)
	},
	Config: func(cfg RuntimeEmailConfig, req QuestionnaireInvitationRequest) RuntimeEmailConfig {
		return req.apply(cfg)
	},
	MessageOptions: func(cfg RuntimeEmailConfig, _ QuestionnaireInvitationRequest) []newman.MessageOption {
		return questionnaireFromOption(cfg)
	},
})

// questionnaireSubject returns the custom subject when set, otherwise the default questionnaire subject
func questionnaireSubject(req QuestionnaireAuthEmail, content QuestionnaireEmailContent) string {
	return lo.CoalesceOrEmpty(content.Subject, "Access "+req.AssessmentName+" Questionnaire from "+req.OrgName)
}

// questionnaireBody builds the questionnaire access email body, using custom text where set and always linking the button to the auth URL
func questionnaireBody(cfg RuntimeEmailConfig, req QuestionnaireAuthEmail, content QuestionnaireEmailContent) render.ContentBody {
	cfg.ButtonColor = tcButtonColor
	cfg.ButtonTextColor = tcButtonTextColor
	cfg = content.apply(cfg)

	intros := content.Intros
	if len(intros) == 0 {
		intros = []string{req.OrgName + " has shared the " + req.AssessmentName + " questionnaire for you to complete."}
	}

	return render.ContentBody{
		Preheader: lo.CoalesceOrEmpty(content.Preheader, req.AssessmentName+" — complete your questionnaire from "+req.OrgName),
		Header:    defaultHeader(cfg),
		Title:     lo.CoalesceOrEmpty(content.Title, req.OrgName+" sent you a questionnaire to complete"),
		Intros:    render.IntrosBlock{Paragraphs: intros},
		Actions: []render.Action{{
			Button: render.Button{
				Text:      lo.CoalesceOrEmpty(content.ButtonText, defaultQuestionnaireButtonText),
				Link:      req.AuthURL,
				Color:     cfg.ButtonColor,
				TextColor: cfg.ButtonTextColor,
			},
		}},
	}
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

// ValidateCampaignTemplateKey reports whether an email template key may be linked to a campaign of the given type
// questionnaire campaigns only accept the questionnaire invitation template, and other campaigns never accept it
func ValidateCampaignTemplateKey(campaignType enums.CampaignType, key string) error {
	isQuestionnaireKey := key == QuestionnaireInvitationOp.Name()

	if (campaignType == enums.CampaignTypeQuestionnaire) != isQuestionnaireKey {
		return fmt.Errorf("%w: %s template cannot be used for a %s campaign", ErrCampaignTemplateKeyMismatch, key, campaignType)
	}

	return nil
}
