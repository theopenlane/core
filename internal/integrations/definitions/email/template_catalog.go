package email

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/providerkit"
)

// nonTemplateConfigFields are RuntimeEmailConfig keys excluded from template
// variables (secrets, provider internals, non-scalar types)
var nonTemplateConfigFields = map[string]struct{}{
	"apiKey":             {},
	"provider":           {},
	"questionnaireEmail": {},
	"social":             {},
}

// payloadVariables are per-send fields from RecipientInfo and CampaignContext
// injected into the template variable map at render time
var payloadVariables = lo.Map(
	append(providerkit.PropertyDescriptors[RecipientInfo](), providerkit.PropertyDescriptors[CampaignContext]()...),
	func(p providerkit.PropertyDescriptor, _ int) models.TemplateVariable {
		return models.TemplateVariable{Name: p.Name, Description: p.Description}
	},
)

// TemplateVariables returns all system-provided template variables available to
// email template authors: config-backed fields, payload fields, and computed values
func TemplateVariables() []models.TemplateVariable {
	configVars := lo.FilterMap(
		providerkit.PropertyDescriptors[RuntimeEmailConfig](),
		func(p providerkit.PropertyDescriptor, _ int) (models.TemplateVariable, bool) {
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
