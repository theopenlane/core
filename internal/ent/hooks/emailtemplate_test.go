//go:build test

package hooks_test

import (
	"testing"

	"github.com/microcosm-cc/bluemonday"
	"github.com/stretchr/testify/assert"

	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
)

func TestEmailScrubber_ScriptTagStripped(t *testing.T) {
	input := `<p>Hello</p><script>alert('xss')</script>`
	got := emaildef.EmailScrubber().Scrub(input)

	assert.NotContains(t, got, "<script>")
	assert.NotContains(t, got, "alert")
	assert.Contains(t, got, "<p>Hello</p>")
}

func TestEmailScrubber_EventHandlerStripped(t *testing.T) {
	input := `<img src="logo.png" onerror="evil()">`
	got := emaildef.EmailScrubber().Scrub(input)

	assert.NotContains(t, got, "onerror")
	assert.NotContains(t, got, "evil()")
}

func TestEmailScrubber_JavascriptHrefStripped(t *testing.T) {
	input := `<a href="javascript:alert(1)">click</a>`
	got := emaildef.EmailScrubber().Scrub(input)

	assert.NotContains(t, got, "javascript:")
}

func TestEmailScrubber_HttpsHrefPreserved(t *testing.T) {
	input := `<a href="https://example.com">link</a>`
	got := emaildef.EmailScrubber().Scrub(input)

	assert.Contains(t, got, `href="https://example.com"`)
}

func TestEmailScrubber_StyleAttributePreserved(t *testing.T) {
	input := `<td style="color:#333;padding:16px">content</td>`
	got := emaildef.EmailScrubber().Scrub(input)

	assert.Contains(t, got, "style=")
	assert.Contains(t, got, "content")
}

func TestEmailScrubber_TableStructurePreserved(t *testing.T) {
	input := `<table width="600" cellpadding="0" cellspacing="0"><tr><td align="center">content</td></tr></table>`
	got := emaildef.EmailScrubber().Scrub(input)

	assert.Contains(t, got, "<table")
	assert.Contains(t, got, "<tr>")
	assert.Contains(t, got, "<td")
}

func TestEmailScrubber_IframeStripped(t *testing.T) {
	input := `<p>text</p><iframe src="https://evil.com"></iframe>`
	got := emaildef.EmailScrubber().Scrub(input)

	assert.NotContains(t, got, "<iframe")
}

func TestStrictPolicySanitize_HTMLStripped(t *testing.T) {
	strict := bluemonday.StrictPolicy()

	input := `<b>Hello</b> <script>alert(1)</script> world`
	got := strict.Sanitize(input)

	assert.Equal(t, "Hello  world", got)
}
