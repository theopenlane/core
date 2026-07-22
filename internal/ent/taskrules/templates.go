package taskrules

import (
	"embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed templates/*.yaml
var templatesFS embed.FS

// Template is the rendered content for one task rule
type Template struct {
	// Title is the title of the generated task
	Title string `yaml:"title"`
	// Details is the body content of the generated task
	Details string `yaml:"details"`
	// Priority is the priority of the generated task
	Priority int `yaml:"priority"`
	// TaskKindName is the kind of task to generate
	TaskKindName string `yaml:"taskKindName,omitempty"`
	// Source is where the task rule originates
	Source Source `yaml:"source,omitempty"`
	// Metadata is arbitrary extra data attached to the task
	Metadata map[string]any `yaml:"metadata,omitempty"`
}

type templateFile struct {
	Rules map[string]Template `yaml:"rules"`
}

var templates = mustLoadTemplates()

const (
	dirName = "templates"
)

// mustLoadTemplates reads every *.yaml file in templates
func mustLoadTemplates() map[string]Template {
	entries, err := templatesFS.ReadDir(dirName)
	if err != nil {
		panic(fmt.Sprintf("taskrules: reading %s dir: %v", dirName, err))
	}

	out := map[string]Template{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		raw, err := templatesFS.ReadFile(dirName + "/" + entry.Name())
		if err != nil {
			panic(fmt.Sprintf("taskrules: reading %s: %v", entry.Name(), err))
		}

		var f templateFile
		if err := yaml.Unmarshal(raw, &f); err != nil {
			panic(fmt.Sprintf("taskrules: invalid template yaml %s: %v", entry.Name(), err))
		}

		for ruleID, tmpl := range f.Rules {
			if _, exists := out[ruleID]; exists {
				panic(fmt.Sprintf("taskrules: duplicate template rule id %q", ruleID))
			}

			out[ruleID] = tmpl
		}
	}

	return out
}

// Lookup returns the rendered template registered for ruleID
func Lookup(ruleID string) (Template, bool) {
	t, ok := templates[ruleID]

	return t, ok
}
