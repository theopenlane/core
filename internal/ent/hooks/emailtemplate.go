package hooks

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"entgo.io/ent"
	"github.com/microcosm-cc/bluemonday"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/templatecontext"
	"github.com/theopenlane/core/pkg/jsonx"
)

// templateVarPattern matches dot-prefixed Go template variable references, e.g. {{ .Foo }} or {{ .User.Name }}.
var templateVarPattern = regexp.MustCompile(`{{\s*\.([A-Za-z][A-Za-z0-9_.]*)\s*}}`)

// tmplURLExpr matches Go template expressions of the form {{ .URLS.<Name> }}.
// Only .URLS.* references are matched; arbitrary template expressions are left
// for the sanitizer to handle normally (which strips them from href attributes).
var tmplURLExpr = regexp.MustCompile(`\{\{\s*\.URLS\.[A-Za-z][A-Za-z0-9_]*\s*\}\}`)

// tmplURLPlaceholderBase is the https base used to temporarily stand in for template
// URL expressions during sanitization. The .invalid TLD is reserved by RFC 2606
// and is not resolvable, making collisions with real content extremely unlikely.
const tmplURLPlaceholderBase = "https://templ-url-placeholder.invalid/"

// EmailTemplateSanitizePolicy returns a bluemonday policy suitable for storing
// HTML email template source. It extends the UGC policy (which already covers tables,
// images, links, and standard elements) with inline style support required for HTML
// email layout. Go template URL expressions ({{ .URLS.* }}) in href and src attributes
// are handled by SanitizeBodyHTML, which preprocesses them before passing content to this policy.
func EmailTemplateSanitizePolicy() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("style").OnElements("table", "tr", "td", "th", "div", "span", "p", "a", "img")

	return p
}

// SanitizeBodyHTML sanitizes HTML email template body content using the provided policy
// while preserving Go template URL expressions ({{ .URLS.<Name> }}) in attribute values.
// Template expressions matching {{ .URLS.<Name> }} are replaced with indexed https placeholder
// URLs before sanitization, then restored afterward. This is necessary because bluemonday's
// URL scheme validation rejects non-URL strings regardless of any custom regex policy.
// All other template expressions in href/src attributes are subject to normal policy enforcement
// and will be stripped by bluemonday's URL scheme check.
func SanitizeBodyHTML(p *bluemonday.Policy, content string) string {
	var originals []string

	processed := tmplURLExpr.ReplaceAllStringFunc(content, func(match string) string {
		placeholder := fmt.Sprintf("%s%d", tmplURLPlaceholderBase, len(originals))
		originals = append(originals, match)

		return placeholder
	})

	sanitized := p.Sanitize(processed)

	for i, original := range originals {
		placeholder := fmt.Sprintf("%s%d", tmplURLPlaceholderBase, i)
		sanitized = strings.Replace(sanitized, placeholder, original, 1)
	}

	return sanitized
}

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

	reserved := templatecontext.ReservedFieldNames()

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

// HookEmailTemplateSanitize sanitizes HTML template content fields on create and update
// for non-system-owned email templates. System-owned templates are loaded via harmonize
// and are trusted; user-composed templates are sanitized to prevent stored XSS.
// Body content preserves {{ .URLS.<Name> }} template expressions; subject, preheader,
// and text fields are stripped of all HTML.
func HookEmailTemplateSanitize() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.EmailTemplateFunc(func(ctx context.Context, m *generated.EmailTemplateMutation) (generated.Value, error) {
			// system-owned templates are loaded via harmonize and are pre-trusted
			if owned, exists := m.SystemOwned(); exists && owned {
				return next.Mutate(ctx, m)
			}

			// on update operations also check the persisted value; the mutation may not re-set system_owned
			if !m.Op().Is(ent.OpCreate) {
				if old, err := m.OldSystemOwned(ctx); err == nil && old {
					return next.Mutate(ctx, m)
				}
			}

			p := EmailTemplateSanitizePolicy()

			if v, exists := m.SubjectTemplate(); exists {
				m.SetSubjectTemplate(bluemonday.StrictPolicy().Sanitize(v))
			}

			if v, exists := m.PreheaderTemplate(); exists {
				m.SetPreheaderTemplate(bluemonday.StrictPolicy().Sanitize(v))
			}

			if v, exists := m.BodyTemplate(); exists {
				m.SetBodyTemplate(SanitizeBodyHTML(p, v))
			}

			if v, exists := m.TextTemplate(); exists {
				m.SetTextTemplate(bluemonday.StrictPolicy().Sanitize(v))
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// HookExtractEmailTemplateVariables parses template content fields on create and update,
// extracts Go template variable references, and merges them as properties into jsonconfig.
// Existing jsonconfig properties are preserved; only newly discovered variables are added.
// System-reserved field names (CompanyName, Recipient, URLS, etc.) are filtered out so
// jsonconfig only describes user-supplied inputs.
// When defaults are also set in the mutation, they are validated against the finalized schema.
func HookExtractEmailTemplateVariables() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.EmailTemplateFunc(func(ctx context.Context, m *generated.EmailTemplateMutation) (generated.Value, error) {
			var templates []string

			if v, exists := m.SubjectTemplate(); exists {
				templates = append(templates, v)
			}

			if v, exists := m.PreheaderTemplate(); exists {
				templates = append(templates, v)
			}

			if v, exists := m.BodyTemplate(); exists {
				templates = append(templates, v)
			}

			if v, exists := m.TextTemplate(); exists {
				templates = append(templates, v)
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
