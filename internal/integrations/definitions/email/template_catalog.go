package email

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/pkg/jsonx"
)

// nonTemplateConfigFields are RuntimeEmailConfig keys excluded from template variables
var nonTemplateConfigFields = map[string]struct{}{
	"apiKey":             {},
	"provider":           {},
	"questionnaireEmail": {},
	"social":             {},
}

// payloadVariables are per-send fields from RecipientInfo and CampaignContext
// injected into the template variable map at render time
var payloadVariables = lo.Map(
	append(jsonx.PropertyDescriptors[RecipientInfo](), jsonx.PropertyDescriptors[CampaignContext]()...),
	func(p jsonx.PropertyDescriptor, _ int) models.TemplateVariable {
		return models.TemplateVariable{Name: p.Name, Description: p.Description}
	},
)

// TemplateVariables returns all system-provided template variables available to
// email template authors: config-backed fields, payload fields, and computed values
func TemplateVariables() []models.TemplateVariable {
	configVars := lo.FilterMap(
		jsonx.PropertyDescriptors[RuntimeEmailConfig](),
		func(p jsonx.PropertyDescriptor, _ int) (models.TemplateVariable, bool) {
			if _, skip := nonTemplateConfigFields[p.Name]; skip {
				return models.TemplateVariable{}, false
			}

			return models.TemplateVariable{Name: p.Name, Description: p.Description}, true
		},
	)

	configVars = append(configVars,
		models.TemplateVariable{Name: "year", Description: "Current year for copyright notices"},
	)

	return append(configVars, payloadVariables...)
}
