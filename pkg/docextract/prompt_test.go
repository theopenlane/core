package docextract

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestRenderPrompt(t *testing.T) {
	rendered, err := RenderPrompt("scope", "Only these: {{ .RefCodes }}.", struct{ RefCodes string }{"A, B"})

	assert.NilError(t, err)
	assert.Check(t, rendered == "\nOnly these: A, B.")
}

func TestRenderPromptInvalidTemplate(t *testing.T) {
	_, err := RenderPrompt("scope", "{{ .Unclosed", nil)

	assert.ErrorIs(t, err, ErrPromptTemplateInvalid)
}

func TestRenderPromptMissingField(t *testing.T) {
	_, err := RenderPrompt("scope", "{{ .Missing }}", struct{ RefCodes string }{})

	assert.ErrorIs(t, err, ErrPromptTemplateInvalid)
}
