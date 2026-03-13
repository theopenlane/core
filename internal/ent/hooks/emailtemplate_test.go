package hooks_test

import (
	"testing"

	"github.com/microcosm-cc/bluemonday"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/ent/hooks"
)

func TestEmailTemplateSanitizePolicy_ScriptTagStripped(t *testing.T) {
	p := hooks.EmailTemplateSanitizePolicy()

	input := `<p>Hello</p><script>alert('xss')</script>`
	got := p.Sanitize(input)

	assert.NotContains(t, got, "<script>")
	assert.NotContains(t, got, "alert")
	assert.Contains(t, got, "<p>Hello</p>")
}

func TestEmailTemplateSanitizePolicy_EventHandlerStripped(t *testing.T) {
	p := hooks.EmailTemplateSanitizePolicy()

	input := `<img src="logo.png" onerror="evil()">`
	got := p.Sanitize(input)

	assert.NotContains(t, got, "onerror")
	assert.NotContains(t, got, "evil()")
}

func TestEmailTemplateSanitizePolicy_JavascriptHrefStripped(t *testing.T) {
	p := hooks.EmailTemplateSanitizePolicy()

	input := `<a href="javascript:alert(1)">click</a>`
	got := p.Sanitize(input)

	assert.NotContains(t, got, "javascript:")
}

func TestEmailTemplateSanitizePolicy_HttpsHrefPreserved(t *testing.T) {
	p := hooks.EmailTemplateSanitizePolicy()

	input := `<a href="https://example.com">link</a>`
	got := p.Sanitize(input)

	assert.Contains(t, got, `href="https://example.com"`)
}

func TestSanitizeBodyHTML_TemplateURLsHrefPreserved(t *testing.T) {
	p := hooks.EmailTemplateSanitizePolicy()

	// {{ .URLS.Verify }} must survive storage-time sanitization via the preprocess/restore path.
	// The raw bluemonday policy rejects non-URL strings in href; SanitizeBodyHTML handles this
	// by replacing template expressions with placeholder https URLs before sanitizing.
	input := `<a href="{{ .URLS.Verify }}">Verify Email</a>`
	got := hooks.SanitizeBodyHTML(p, input)

	assert.Contains(t, got, `{{ .URLS.Verify }}`)
}

func TestSanitizeBodyHTML_NonURLTemplateHrefStripped(t *testing.T) {
	p := hooks.EmailTemplateSanitizePolicy()

	// Only .URLS.* expressions are preprocessed; arbitrary template expressions in href
	// are left for bluemonday's URL scheme validation to strip.
	input := `<a href="{{ .SomeOtherVar }}">link</a>`
	got := hooks.SanitizeBodyHTML(p, input)

	assert.NotContains(t, got, "{{ .SomeOtherVar }}")
}

func TestSanitizeBodyHTML_MultipleTemplateURLsPreserved(t *testing.T) {
	p := hooks.EmailTemplateSanitizePolicy()

	input := `<a href="{{ .URLS.Verify }}">Verify</a> <a href="{{ .URLS.Invite }}">Invite</a>`
	got := hooks.SanitizeBodyHTML(p, input)

	assert.Contains(t, got, `{{ .URLS.Verify }}`)
	assert.Contains(t, got, `{{ .URLS.Invite }}`)
}

func TestEmailTemplateSanitizePolicy_StyleAttributePreserved(t *testing.T) {
	p := hooks.EmailTemplateSanitizePolicy()

	input := `<td style="color:#333;padding:16px">content</td>`
	got := p.Sanitize(input)

	assert.Contains(t, got, "style=")
	assert.Contains(t, got, "content")
}

func TestEmailTemplateSanitizePolicy_TableStructurePreserved(t *testing.T) {
	p := hooks.EmailTemplateSanitizePolicy()

	input := `<table width="600" cellpadding="0" cellspacing="0"><tr><td align="center">content</td></tr></table>`
	got := p.Sanitize(input)

	assert.Contains(t, got, "<table")
	assert.Contains(t, got, "<tr>")
	assert.Contains(t, got, "<td")
}

func TestEmailTemplateSanitizePolicy_IframeStripped(t *testing.T) {
	p := hooks.EmailTemplateSanitizePolicy()

	input := `<p>text</p><iframe src="https://evil.com"></iframe>`
	got := p.Sanitize(input)

	assert.NotContains(t, got, "<iframe")
}

func TestStrictPolicySanitize_HTMLStripped(t *testing.T) {
	// Subject and preheader fields use StrictPolicy — all HTML must be stripped
	strict := bluemonday.StrictPolicy()

	input := `<b>Hello</b> <script>alert(1)</script> world`
	got := strict.Sanitize(input)

	assert.Equal(t, "Hello  world", got)
}
