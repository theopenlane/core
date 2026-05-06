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
		"companyName", "fromEmail", "productURL", "year",
		"email", "firstName", "lastName",
	} {
		assert.Contains(t, names, expected, "missing base variable: %s", expected)
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
