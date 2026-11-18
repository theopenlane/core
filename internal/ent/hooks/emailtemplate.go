package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	emaildef "github.com/theopenlane/core/v2/internal/integrations/definitions/email"
)

// HookEmailTemplateSanitize sanitizes customer-supplied fields on email template
// create and update mutations. String values in the defaults map are scrubbed of
// HTML tags to prevent stored XSS; Go template expressions like {{ .companyName }}
// pass through unmodified since they are not HTML
func HookEmailTemplateSanitize() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.EmailTemplateFunc(func(ctx context.Context, m *generated.EmailTemplateMutation) (generated.Value, error) {
			defaults, hasDefaults := m.Defaults()
			if hasDefaults || m.Op().Is(ent.OpCreate) {
				key, err := emailTemplateKey(ctx, m)
				if err != nil {
					return nil, err
				}

				if err := emaildef.ValidateTemplateDefaults(key, defaults); err != nil {
					return nil, err
				}
			}

			if len(defaults) > 0 {
				m.SetDefaults(scrubMapStrings(defaults))
			}

			if name, ok := m.Name(); ok {
				m.SetName(emaildef.Scrubber().Scrub(name))
			}

			if desc, ok := m.Description(); ok {
				m.SetDescription(emaildef.Scrubber().Scrub(desc))
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// emailTemplateKey returns the catalog key being set, or the stored key for a single-row update
func emailTemplateKey(ctx context.Context, m *generated.EmailTemplateMutation) (string, error) {
	if key, ok := m.Key(); ok {
		return key, nil
	}

	if !m.Op().Is(ent.OpUpdateOne) {
		return "", nil
	}

	return m.OldKey(ctx)
}

// scrubMapStrings recursively scrubs all string values in a map using the email
// HTML scrubber. Nested maps are traversed; non-string values are left untouched
func scrubMapStrings(m map[string]any) map[string]any {
	scrubber := emaildef.Scrubber()

	for k, v := range m {
		switch val := v.(type) {
		case string:
			m[k] = scrubber.Scrub(val)
		case map[string]any:
			m[k] = scrubMapStrings(val)
		}
	}

	return m
}
