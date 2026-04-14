package email

import (
	"github.com/theopenlane/core/common/models"
)

// baseVariables are the template variables available to template authors in all email contexts
var baseVariables = []models.TemplateVariable{
	{Name: "companyName", Description: "Company display name"},
	{Name: "companyAddress", Description: "Company mailing address"},
	{Name: "corporation", Description: "Legal corporation name"},
	{Name: "fromEmail", Description: "Sender email address"},
	{Name: "supportEmail", Description: "Support contact email address"},
	{Name: "logoURL", Description: "Company logo URL"},
	{Name: "rootURL", Description: "Root application URL"},
	{Name: "productURL", Description: "Product home URL"},
	{Name: "docsURL", Description: "Documentation URL"},
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
	names := make(map[string]struct{}, len(allVariables)+4) //nolint:mnd

	for _, v := range allVariables {
		names[v.Name] = struct{}{}
	}

	// Keys injected at render time but not shown in the UI variable picker
	names["branding"] = struct{}{}
	names["apiKey"] = struct{}{}
	names["provider"] = struct{}{}
	names["questionnaireEmail"] = struct{}{}

	return names
}()

// ReservedFieldNames returns the set of top-level template variable names that are
// injected by the system at render time and should be excluded from jsonconfig
func ReservedFieldNames() map[string]struct{} {
	return reservedFieldNames
}
