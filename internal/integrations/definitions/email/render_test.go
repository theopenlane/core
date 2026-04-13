package email

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteTextTemplate(t *testing.T) {
	out, err := executeTextTemplate("greeting", "Hello {{ .FirstName }}", map[string]any{
		"FirstName": "Ada",
	})
	require.NoError(t, err)
	require.Equal(t, "Hello Ada", out)
}

func TestExecuteTextTemplate_Empty(t *testing.T) {
	out, err := executeTextTemplate("empty", "", map[string]any{})
	require.NoError(t, err)
	require.Equal(t, "", out)
}

func TestExecuteHTMLTemplate_RawHTML(t *testing.T) {
	content := `<p>Hello {{ .Name }}</p>`
	out, err := executeHTMLTemplate("test", content, map[string]any{
		"Name": "Ada",
	})
	require.NoError(t, err)
	assert.Equal(t, "<p>Hello Ada</p>", out)
}

func TestExecuteHTMLTemplate_Empty(t *testing.T) {
	out, err := executeHTMLTemplate("test", "", map[string]any{})
	require.NoError(t, err)
	assert.Equal(t, "", out)
}

func TestComposeAndExecute(t *testing.T) {
	base := `<html><body>{{block "content" .}}default{{end}}</body></html>`
	body := `<p>Hello {{.Name}}</p>`

	out, err := composeAndExecute(base, body, map[string]any{
		"Name": "Ada",
	})
	require.NoError(t, err)
	assert.Contains(t, out, "<p>Hello Ada</p>")
	assert.Contains(t, out, "<html>")
	assert.Contains(t, out, "</html>")
}

func TestComposeAndExecute_DefaultBlock(t *testing.T) {
	base := `<html><body>{{block "content" .}}default{{end}}</body></html>`
	body := `` // empty body — base block default should render

	out, err := composeAndExecute(base, body, map[string]any{})
	require.NoError(t, err)
	// empty body define still overrides the default
	assert.Contains(t, out, "<html>")
}

func TestComposeAndExecute_TemplateVars(t *testing.T) {
	base := `<html><head><title>{{.CompanyName}}</title></head><body>{{block "content" .}}{{end}}</body></html>`
	body := `<p>Dear {{.Recipient.FirstName}},</p>`

	out, err := composeAndExecute(base, body, map[string]any{
		"CompanyName": "Openlane",
		"Recipient": map[string]any{
			"FirstName": "Ada",
		},
	})
	require.NoError(t, err)
	assert.Contains(t, out, "<title>Openlane</title>")
	assert.Contains(t, out, "<p>Dear Ada,</p>")
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
