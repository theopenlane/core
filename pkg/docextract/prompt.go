package docextract

import (
	"fmt"
	"strings"
	"text/template"
)

// RenderPrompt executes a configured prompt template against its values, prefixing a newline so
// the result can be appended to another prompt
func RenderPrompt(name, text string, vars any) (string, error) {
	tmpl, err := template.New(name).Parse(text)
	if err != nil {
		return "", fmt.Errorf("%w: %s: %w", ErrPromptTemplateInvalid, name, err)
	}

	var out strings.Builder
	if err := tmpl.Execute(&out, vars); err != nil {
		return "", fmt.Errorf("%w: %s: %w", ErrPromptTemplateInvalid, name, err)
	}

	return "\n" + out.String(), nil
}
