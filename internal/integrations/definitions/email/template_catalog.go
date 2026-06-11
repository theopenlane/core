package email

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/pkg/jsonx"
)

// customerTemplateConfigFields is the allowlist of RuntimeEmailConfig keys (by JSON name)
// appropriate to expose as customer-usable template variables. Config fields are operational,
// secret, or presentational by default, so they must be explicitly opted in here. It is an
// allowlist precisely so a newly added config field — or a secret like apikey/resendsecret —
// never auto-leaks into a customer-facing template or its rendered output
var customerTemplateConfigFields = map[string]struct{}{
	"companyName":  {},
	"corporation":  {},
	"supportemail": {},
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
			if _, ok := customerTemplateConfigFields[p.Name]; !ok {
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
