package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
)

// HookEmailTemplateSanitize sanitizes customer-supplied fields on email template
// create and update mutations. String values in the defaults map are scrubbed of
// HTML tags to prevent stored XSS; Go template expressions like {{ .companyName }}
// pass through unmodified since they are not HTML
func HookEmailTemplateSanitize() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.EmailTemplateFunc(func(ctx context.Context, m *generated.EmailTemplateMutation) (generated.Value, error) {
			if defaults, ok := m.Defaults(); ok && len(defaults) > 0 {
				m.SetDefaults(scrubMapStrings(defaults))
			}

			if name, ok := m.Name(); ok {
				m.SetName(emaildef.EmailScrubber().Scrub(name))
			}

			if desc, ok := m.Description(); ok {
				m.SetDescription(emaildef.EmailScrubber().Scrub(desc))
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// scrubMapStrings recursively scrubs all string values in a map using the email
// HTML scrubber. Nested maps are traversed; non-string values are left untouched
func scrubMapStrings(m map[string]any) map[string]any {
	scrubber := emaildef.EmailScrubber()

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
