package engine

import (
	"bytes"
	"strings"
	"text/template"
)

// renderTemplateText executes Go text/template expressions in the input string
// against the provided vars map
func renderTemplateText(input string, vars map[string]any) (string, error) {
	if strings.TrimSpace(input) == "" || !strings.Contains(input, "{{") {
		return input, nil
	}

	tpl, err := template.New("").Option("missingkey=zero").Parse(input)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, vars); err != nil {
		return "", err
	}

	return strings.ReplaceAll(buf.String(), "<no value>", ""), nil
}
