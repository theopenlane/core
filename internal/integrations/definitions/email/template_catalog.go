package email

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/providerkit"
)

// nonTemplateConfigFields are json keys that are excluded from the template variable picker
var nonTemplateConfigFields = map[string]struct{}{
	"apiKey":             {},
	"provider":           {},
	"questionnaireEmail": {},
	"social":             {},
}

// computedVariables are template variables injected at render time
var computedVariables = []models.TemplateVariable{
	{Name: "year", Description: "Current year for copyright notices"},
}

// payloadVariables are template variables derived from the operation payload and injected at render time
var payloadVariables = lo.Map(
	append(providerkit.PropertyDescriptors[RecipientInfo](), providerkit.PropertyDescriptors[CampaignContext]()...),
	func(p providerkit.PropertyDescriptor, _ int) models.TemplateVariable {
		return models.TemplateVariable{Name: p.Name, Description: p.Description}
	},
)

// baseVariables are the template variables available to template authors in all email contexts
var baseVariables = buildBaseVariables()

// buildBaseVariables reflects RuntimeEmailConfig for config-backed template
// variables and appends the computed variables injected at render time
func buildBaseVariables() []models.TemplateVariable {
	props := providerkit.PropertyDescriptors[RuntimeEmailConfig]()

	configVars := lo.FilterMap(props, func(p providerkit.PropertyDescriptor, _ int) (models.TemplateVariable, bool) {
		if _, skip := nonTemplateConfigFields[p.Name]; skip {
			return models.TemplateVariable{}, false
		}

		return models.TemplateVariable{Name: p.Name, Description: p.Description}, true
	})

	return append(configVars, computedVariables...)
}

// allVariables is the complete appended set
var allVariables = append(append([]models.TemplateVariable{}, baseVariables...), payloadVariables...)

// TemplateVariables returns all system-provided template variables available for use in email templates
func TemplateVariables() []models.TemplateVariable {
	return allVariables
}

// reservedFieldNames is the set of top-level template data keys injected by the system at render time
var reservedFieldNames = func() map[string]struct{} {
	names := lo.SliceToMap(allVariables, func(v models.TemplateVariable) (string, struct{}) {
		return v.Name, struct{}{}
	})

	for name := range nonTemplateConfigFields {
		names[name] = struct{}{}
	}

	names["branding"] = struct{}{}

	return names
}()

// ReservedFieldNames returns the set of top-level template variable names
func ReservedFieldNames() map[string]struct{} {
	return reservedFieldNames
}
