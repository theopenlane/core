package emailruntime

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeGoTemplateShorthand(t *testing.T) {
	input := "Hi {{ FirstName }} {{ .LastName }}"
	out := normalizeGoTemplateShorthand(input)

	require.Equal(t, "Hi {{ .FirstName }} {{ .LastName }}", out)
}

func TestNormalizeGoTemplateShorthandSkipsKeywords(t *testing.T) {
	input := "{{ end }} {{ range .Items }}{{ end }}"
	out := normalizeGoTemplateShorthand(input)

	require.Equal(t, input, out)
}

func TestRenderTextTemplateWithBareIdentifier(t *testing.T) {
	out, err := renderTextTemplate("greeting", "Hello {{ FirstName }}", "", map[string]any{
		"FirstName": "Ada",
	})
	require.NoError(t, err)
	require.Equal(t, "Hello Ada", out)
}

func TestRenderHTMLTemplate_RawHTMLNoBase(t *testing.T) {
	content := `<p>Hello {{ .Name }}</p>`
	out, err := renderHTMLTemplate(RenderModeRawHTML, "test", content, "", "", map[string]any{
		"Name": "Ada",
	})
	require.NoError(t, err)
	assert.Equal(t, "<p>Hello Ada</p>", out)
}

func TestRenderHTMLTemplate_EmptyContentReturnsEmpty(t *testing.T) {
	out, err := renderHTMLTemplate(RenderModeRawHTML, "test", "", "", "", map[string]any{})
	require.NoError(t, err)
	assert.Equal(t, "", out)
}

func TestRenderHTMLTemplate_BundleMode_UsesEntrypoint(t *testing.T) {
	content := `{{define "main"}}<p>{{.Name}}</p>{{end}}{{define "other"}}other{{end}}`
	out, err := renderHTMLTemplate(RenderModeGoTemplateBundle, "bundle", content, "", "main", map[string]any{
		"Name": "World",
	})
	require.NoError(t, err)
	assert.Equal(t, "<p>World</p>", out)
}

func TestRenderHTMLWithBase_InjectsContentBlock(t *testing.T) {
	base := `{{define "base"}}<html><body>{{block "content" .}}{{end}}</body></html>{{end}}`
	customer := `{{define "content"}}<p>Hello {{.Name}}</p>{{end}}`

	out, err := renderHTMLWithBase("customer", base, customer, map[string]any{
		"Name": "Ada",
	})
	require.NoError(t, err)
	assert.Contains(t, out, "<p>Hello Ada</p>")
	assert.Contains(t, out, "<html>")
	assert.Contains(t, out, "</html>")
}

func TestRenderHTMLWithBase_DefaultBlockWhenNoOverride(t *testing.T) {
	base := `{{define "base"}}<html><body>{{block "content" .}}default{{end}}</body></html>{{end}}`
	// customer defines a different block — content gets the default
	customer := `{{define "other"}}irrelevant{{end}}`

	out, err := renderHTMLWithBase("customer", base, customer, map[string]any{})
	require.NoError(t, err)
	assert.Contains(t, out, "default")
}

func TestRenderHTMLWithBase_TemplateVariablesSubstituted(t *testing.T) {
	base := `{{define "base"}}<html><head><title>{{.CompanyName}}</title></head><body>{{block "content" .}}{{end}}</body></html>{{end}}`
	customer := `{{define "content"}}<p>Dear {{.Recipient.FirstName}},</p>{{end}}`

	out, err := renderHTMLWithBase("customer", base, customer, map[string]any{
		"CompanyName": "Openlane",
		"Recipient": map[string]any{
			"FirstName": "Ada",
		},
	})
	require.NoError(t, err)
	assert.Contains(t, out, "<title>Openlane</title>")
	assert.Contains(t, out, "<p>Dear Ada,</p>")
}

func TestRenderHTMLWithBase_BareIdentifiersNormalized(t *testing.T) {
	// normalizeGoTemplateShorthand runs on base content before parsing
	base := `{{define "base"}}<html><body>{{ CompanyName }}: {{block "content" .}}{{end}}</body></html>{{end}}`
	customer := `{{define "content"}}<p>hello</p>{{end}}`

	out, err := renderHTMLWithBase("customer", base, customer, map[string]any{
		"CompanyName": "Openlane",
	})
	require.NoError(t, err)
	assert.Contains(t, out, "Openlane:")
}

func TestRenderTimeHTMLSanitize_ScriptStripped(t *testing.T) {
	html := `<p>Safe content</p><script>alert('xss')</script>`
	out := renderTimeHTMLSanitize(html)

	assert.NotContains(t, out, "<script>")
	assert.NotContains(t, out, "alert")
	assert.Contains(t, out, "Safe content")
}

func TestRenderTimeHTMLSanitize_TablePreserved(t *testing.T) {
	html := `<table width="600"><tr><td style="padding:16px">content</td></tr></table>`
	out := renderTimeHTMLSanitize(html)

	assert.Contains(t, out, "<table")
	assert.Contains(t, out, "content")
}

func TestRenderTimeHTMLSanitize_StyleAttributePreserved(t *testing.T) {
	html := `<p style="color:#333;font-size:14px">styled text</p>`
	out := renderTimeHTMLSanitize(html)

	assert.True(t, strings.Contains(out, "style="), "style attribute should be preserved")
	assert.Contains(t, out, "styled text")
}
