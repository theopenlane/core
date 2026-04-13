//go:build test

package hooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractTemplateVarNames_Simple(t *testing.T) {
	vars := extractTemplateVarNames("Hello {{ .FirstName }}, welcome to {{ .CompanyName }}")
	assert.Equal(t, map[string]string{"FirstName": "string", "CompanyName": "string"}, vars)
}

func TestExtractTemplateVarNames_DottedPathTopLevelOnly(t *testing.T) {
	vars := extractTemplateVarNames("{{ .User.FirstName }} from {{ .Org.Name }}")
	assert.Equal(t, map[string]string{"User": "object", "Org": "object"}, vars)
}

func TestExtractTemplateVarNames_Deduplicated(t *testing.T) {
	vars := extractTemplateVarNames("{{ .Name }} and {{ .Name }} again")
	assert.Equal(t, map[string]string{"Name": "string"}, vars)
}

func TestExtractTemplateVarNames_AcrossMultipleTemplates(t *testing.T) {
	vars := extractTemplateVarNames("Subject: {{ .CompanyName }}", "<p>Hello {{ .FirstName }}</p>")
	assert.Equal(t, map[string]string{"CompanyName": "string", "FirstName": "string"}, vars)
}

func TestExtractTemplateVarNames_NoVars(t *testing.T) {
	vars := extractTemplateVarNames("No template expressions here")
	assert.Empty(t, vars)
}

func TestExtractTemplateVarNames_URLsExtractedAsObject(t *testing.T) {
	vars := extractTemplateVarNames(`<a href="{{ .URLS.Verify }}">click</a>`)
	assert.Equal(t, map[string]string{"URLS": "object"}, vars)
}

func TestMergeTemplateVarsIntoSchema_NilSchema(t *testing.T) {
	result := mergeTemplateVarsIntoSchema(nil, map[string]string{"SupportEmail": "string", "CompanyName": "string"})

	require.Equal(t, "object", result["type"])

	props, ok := result["properties"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, props, "SupportEmail")
	assert.Contains(t, props, "CompanyName")
}

func TestMergeTemplateVarsIntoSchema_PreservesExistingProperties(t *testing.T) {
	existing := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"SupportEmail": map[string]any{"type": "string", "description": "Support contact"},
		},
	}

	result := mergeTemplateVarsIntoSchema(existing, map[string]string{"SupportEmail": "string", "FirstName": "string"})

	props, ok := result["properties"].(map[string]any)
	require.True(t, ok)

	// existing property with description must be untouched
	support, ok := props["SupportEmail"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Support contact", support["description"])

	// new variable added with default string type
	assert.Contains(t, props, "FirstName")
}

func TestMergeTemplateVarsIntoSchema_SetsObjectType(t *testing.T) {
	result := mergeTemplateVarsIntoSchema(map[string]any{}, map[string]string{"Foo": "string"})
	assert.Equal(t, "object", result["type"])
}

func TestMergeTemplateVarsIntoSchema_PreservesExistingType(t *testing.T) {
	existing := map[string]any{"type": "object", "title": "My Template"}
	result := mergeTemplateVarsIntoSchema(existing, map[string]string{"Foo": "string"})
	assert.Equal(t, "My Template", result["title"])
}

func TestMergeTemplateVarsIntoSchema_ObjectTypeProperty(t *testing.T) {
	// variables accessed via dotted paths (e.g. .URLS.Verify) must be stored as type: object
	result := mergeTemplateVarsIntoSchema(nil, map[string]string{"URLS": "object"})

	props, ok := result["properties"].(map[string]any)
	require.True(t, ok)

	urlsProp, ok := props["URLS"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "object", urlsProp["type"])
}
