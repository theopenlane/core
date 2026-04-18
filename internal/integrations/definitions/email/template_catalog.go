package email

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/providerkit"
)

// nonTemplateConfigFields are RuntimeEmailConfig json keys that are excluded from
// the template variable picker. They are credentials or internal routing knobs
// that are not meaningful template content, but remain reserved so user-supplied
// jsonconfig cannot redefine them
var nonTemplateConfigFields = map[string]struct{}{
	"apiKey":             {},
	"provider":           {},
	"questionnaireEmail": {},
}

// computedVariables are template variables injected at render time that are not
// backed by RuntimeEmailConfig fields, such as time-derived or per-recipient values
var computedVariables = []models.TemplateVariable{
	{Name: "year", Description: "Current year for copyright notices"},
	{Name: "recipientEmail", Description: "Recipient email address"},
	{Name: "recipientFirstName", Description: "Recipient first name"},
	{Name: "recipientLastName", Description: "Recipient last name"},
}

// campaignVariables are additional template variables available in campaign email contexts
var campaignVariables = []models.TemplateVariable{
	{Name: "campaignName", Description: "Campaign name"},
	{Name: "campaignDescription", Description: "Campaign description"},
}

// baseVariables are the template variables available to template authors in all
// email contexts. Config-backed entries are reflected from RuntimeEmailConfig so
// adding a new field with a jsonschema description automatically exposes it to
// templates without a second source of truth
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

// allVariables is baseVariables + campaignVariables
var allVariables = append(append([]models.TemplateVariable{}, baseVariables...), campaignVariables...)

// TemplateVariables returns all system-provided template variables available for
// use in email templates, including campaign-specific variables
func TemplateVariables() []models.TemplateVariable {
	return allVariables
}

// reservedFieldNames is the set of top-level template data keys injected by the
// system at render time. Used by the emailtemplate hook to exclude these from
// jsonconfig so it only describes user-supplied inputs
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

// ReservedFieldNames returns the set of top-level template variable names that are
// injected by the system at render time and should be excluded from jsonconfig
func ReservedFieldNames() map[string]struct{} {
	return reservedFieldNames
}
