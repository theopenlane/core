package email

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestTemplateVariables_NotEmpty(t *testing.T) {
	vars := TemplateVariables()

	assert.Assert(t, len(vars) > 0)
}

func TestTemplateVariables_ContainsBaseVars(t *testing.T) {
	vars := TemplateVariables()

	names := make(map[string]struct{}, len(vars))
	for _, v := range vars {
		names[v.Name] = struct{}{}
	}

	for _, expected := range []string{
		"companyName", "corporation", "supportemail", "year",
		"email", "firstName", "lastName",
	} {
		assert.Check(t, is.Contains(names, expected), "missing base variable: %s", expected)
	}
}

// TestTemplateVariables_ExcludesOperationalAndSecretFields verifies operational, secret, and
// presentational config fields are never advertised as customer-usable template variables
func TestTemplateVariables_ExcludesOperationalAndSecretFields(t *testing.T) {
	vars := TemplateVariables()

	names := make(map[string]struct{}, len(vars))
	for _, v := range vars {
		names[v.Name] = struct{}{}
	}

	for _, excluded := range []string{
		"apikey", "resendsecret", "testdir", "provider",
		"rooturl", "producturl", "docsurl", "fromemail",
		"logoURL", "buttonColor", "cardStyle",
	} {
		_, found := names[excluded]
		assert.Check(t, !found, "operational/secret field leaked into template variables: %s", excluded)
	}
}

func TestTemplateVariables_ContainsCampaignVars(t *testing.T) {
	vars := TemplateVariables()

	names := make(map[string]struct{}, len(vars))
	for _, v := range vars {
		names[v.Name] = struct{}{}
	}

	assert.Check(t, is.Contains(names, "campaignName"))
	assert.Check(t, is.Contains(names, "campaignDescription"))
}

func TestTemplateVariables_AllHaveDescriptions(t *testing.T) {
	for _, v := range TemplateVariables() {
		assert.Check(t, len(v.Description) > 0, "variable %s missing description", v.Name)
	}
}

func TestTemplateVariables_NoDuplicates(t *testing.T) {
	vars := TemplateVariables()

	seen := make(map[string]struct{}, len(vars))
	for _, v := range vars {
		_, dup := seen[v.Name]
		assert.Check(t, !dup, "duplicate variable name: %s", v.Name)
		seen[v.Name] = struct{}{}
	}
}
