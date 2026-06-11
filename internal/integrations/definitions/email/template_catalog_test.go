package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateVariables_NotEmpty(t *testing.T) {
	vars := TemplateVariables()

	require.NotEmpty(t, vars)
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
		assert.Contains(t, names, expected, "missing base variable: %s", expected)
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
		assert.NotContains(t, names, excluded, "operational/secret field leaked into template variables: %s", excluded)
	}
}

func TestTemplateVariables_ContainsCampaignVars(t *testing.T) {
	vars := TemplateVariables()

	names := make(map[string]struct{}, len(vars))
	for _, v := range vars {
		names[v.Name] = struct{}{}
	}

	assert.Contains(t, names, "campaignName")
	assert.Contains(t, names, "campaignDescription")
}

func TestTemplateVariables_AllHaveDescriptions(t *testing.T) {
	for _, v := range TemplateVariables() {
		assert.NotEmpty(t, v.Description, "variable %s missing description", v.Name)
	}
}

func TestTemplateVariables_NoDuplicates(t *testing.T) {
	vars := TemplateVariables()

	seen := make(map[string]struct{}, len(vars))
	for _, v := range vars {
		_, dup := seen[v.Name]
		assert.False(t, dup, "duplicate variable name: %s", v.Name)
		seen[v.Name] = struct{}{}
	}
}
