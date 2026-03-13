package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/microcosm-cc/bluemonday"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
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
