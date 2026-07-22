package taskrules

import (
	"embed"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed templates/*.yaml
var templatesFS embed.FS

// DocsBaseURL is prepended to any metadata.docsLink value that starts with "/", so task
// definitions can use relative paths (e.g. "/compliance/controls") instead of full URLs
const DocsBaseURL = "https://docs.theopenlane.io"

// Template is the rendered content for one task rule
type Template struct {
	Title        string         `yaml:"title"`
	Details      string         `yaml:"details"`
	Priority     int            `yaml:"priority"`
	TaskKindName string         `yaml:"taskKindName,omitempty"`
	Source       Source         `yaml:"source,omitempty"`
	Metadata     map[string]any `yaml:"metadata,omitempty"`
}

// validSources are the Source values allowed in a template YAML file; empty defaults to
// SourceRecommendations at lookup time, so it's valid here too
var validSources = map[Source]bool{
	"":                    true,
	SourceRecommendations: true,
	SourceOnboarding:      true,
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

			if !validSources[tmpl.Source] {
				panic(fmt.Sprintf("taskrules: rule %q has unrecognized source %q", ruleID, tmpl.Source))
			}

			resolveDocsLink(tmpl.Metadata)

			out[ruleID] = tmpl
		}
	}

	return out
}

// resolveDocsLink rewrites a relative metadata.docsLink value (starting with "/") to a full URL
// under DocsBaseURL, in place. Absolute URLs and missing/non-string values are left as-is
func resolveDocsLink(metadata map[string]any) {
	link, ok := metadata["docsLink"].(string)
	if !ok || !strings.HasPrefix(link, "/") {
		return
	}

	metadata["docsLink"] = DocsBaseURL + link
}

// Lookup returns the rendered template registered for ruleID
func Lookup(ruleID string) (Template, bool) {
	t, ok := templates[ruleID]

	return t, ok
}
