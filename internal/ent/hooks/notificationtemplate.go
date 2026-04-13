package hooks

import (
	"context"
	"encoding/json"

	"entgo.io/ent"
	"github.com/microcosm-cc/bluemonday"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/jsonx"
)

// HookNotificationTemplateSanitize sanitizes template content fields on create and update
// for non-system-owned notification templates. System-owned templates are loaded via
// harmonize and are trusted. Body content is sanitized with the email-aware bluemonday
// policy; title and subject fields are stripped of all HTML tags.
func HookNotificationTemplateSanitize() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.NotificationTemplateFunc(func(ctx context.Context, m *generated.NotificationTemplateMutation) (generated.Value, error) {
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
			strict := bluemonday.StrictPolicy()

			if v, exists := m.TitleTemplate(); exists {
				m.SetTitleTemplate(strict.Sanitize(v))
			}

			if v, exists := m.SubjectTemplate(); exists {
				m.SetSubjectTemplate(strict.Sanitize(v))
			}

			if v, exists := m.BodyTemplate(); exists {
				m.SetBodyTemplate(SanitizeBodyHTML(p, v))
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
