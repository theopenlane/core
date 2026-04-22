package hooks

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/pkg/jsonx"
)

// HookNotificationTemplateSanitize sanitizes template content fields on create and update
// for non-system-owned notification templates. System-owned templates are loaded via
// harmonize and are trusted. Body content is sanitized with the email-aware bluemonday
// policy; title and subject fields are stripped of all HTML tags.
func HookNotificationTemplateSanitize() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.NotificationTemplateFunc(func(ctx context.Context, m *generated.NotificationTemplateMutation) (generated.Value, error) {
			if v, exists := m.TitleTemplate(); exists {
				m.SetTitleTemplate(emaildef.EmailScrubber().Scrub(v))
			}

			if v, exists := m.SubjectTemplate(); exists {
				m.SetSubjectTemplate(emaildef.EmailScrubber().Scrub(v))
			}

			if v, exists := m.BodyTemplate(); exists {
				m.SetBodyTemplate(emaildef.EmailScrubber().Scrub(v))
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// HookExtractNotificationTemplateVariables parses template content fields on create and update,
// extracts Go template variable references, and merges them as properties into jsonconfig.
// Existing jsonconfig properties are preserved; only newly discovered variables are added.
// System-reserved field names (CompanyName, Recipient, URLS, etc.) are filtered out so
// jsonconfig only describes user-supplied inputs.
// When defaults are also set in the mutation, they are validated against the finalized schema.
func HookExtractNotificationTemplateVariables() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.NotificationTemplateFunc(func(ctx context.Context, m *generated.NotificationTemplateMutation) (generated.Value, error) {
			var templates []string

			if v, exists := m.TitleTemplate(); exists {
				templates = append(templates, v)
			}

			if v, exists := m.SubjectTemplate(); exists {
				templates = append(templates, v)
			}

			if v, exists := m.BodyTemplate(); exists {
				templates = append(templates, v)
			}

			// blocks is structured JSON, but may contain template expressions in string values
			if v, exists := m.Blocks(); exists && len(v) > 0 {
				if raw, err := json.Marshal(v); err == nil {
					templates = append(templates, string(raw))
				}
			}

			vars := extractTemplateVarNames(templates...)

			defaults, defaultsSet := m.Defaults()

			if len(vars) == 0 && !defaultsSet {
				return next.Mutate(ctx, m)
			}

			var jsonconfig map[string]any

			if v, exists := m.Jsonconfig(); exists {
				jsonconfig = v
			} else if !m.Op().Is(ent.OpCreate) {
				if old, err := m.OldJsonconfig(ctx); err == nil {
					jsonconfig = old
				}
			}

			finalSchema := mergeTemplateVarsIntoSchema(jsonconfig, vars)
			m.SetJsonconfig(finalSchema)

			if defaultsSet && len(defaults) > 0 {
				result, err := jsonx.ValidateSchema(finalSchema, defaults)
				if err != nil || !result.Valid() {
					return nil, ErrInvalidTemplateDefaults
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// templateVarPattern matches dot-prefixed Go template variable references, e.g. {{ .Foo }} or {{ .User.Name }}.
var templateVarPattern = regexp.MustCompile(`{{\s*\.([A-Za-z][A-Za-z0-9_.]*)\s*}}`)

// extractTemplateVarNames scans the provided template strings for {{ .VarName }} references
// and returns a map of top-level variable name to JSON Schema type. Dotted paths like
// {{ .User.Name }} yield "User" with type "object". Direct references like {{ .Name }}
// yield type "string"
func extractTemplateVarNames(templates ...string) map[string]string {
	vars := map[string]string{}

	for _, tmpl := range templates {
		for _, match := range templateVarPattern.FindAllStringSubmatch(tmpl, -1) {
			if len(match) < 2 {
				continue
			}

			top, rest, found := strings.Cut(match[1], ".")
			if found && rest != "" {
				vars[top] = "object"
			} else if _, exists := vars[top]; !exists {
				vars[top] = "string"
			}
		}
	}

	return vars
}

// mergeTemplateVarsIntoSchema adds discovered variable names as typed properties
// into a JSON Schema map. Existing properties are preserved; only absent keys are added.
// Variables whose names match system-reserved fields (injected at render time by the
// email client config) are excluded so that jsonconfig only describes user-supplied inputs
func mergeTemplateVarsIntoSchema(schema map[string]any, vars map[string]string) map[string]any {
	if schema == nil {
		schema = map[string]any{}
	}

	if _, ok := schema["type"]; !ok {
		schema["type"] = "object"
	}

	props, _ := schema["properties"].(map[string]any)
	if props == nil {
		props = map[string]any{}
	}

	reserved := emaildef.ReservedFieldNames()

	for name, typ := range vars {
		if _, isReserved := reserved[name]; isReserved {
			continue
		}

		if _, exists := props[name]; !exists {
			props[name] = map[string]any{"type": typ}
		}
	}

	schema["properties"] = props

	return schema
}
