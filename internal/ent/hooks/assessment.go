package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/logx"
)

// HookQuestionnaireAssessment is a hook that checks if the templatate associated with the assessment is a questionnaire
func HookQuestionnaireAssessment() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.AssessmentFunc(func(ctx context.Context, m *generated.AssessmentMutation) (generated.Value, error) {
			templateID, ok := m.TemplateID()
			if !ok {
				return nil, ErrTemplateIDRequired
			}

			tmpl, err := m.Client().Template.Query().
				Where(template.ID(templateID)).
				Only(ctx)
			if err != nil {
				if generated.IsNotFound(err) {
					logx.FromContext(ctx).Warn().Str("template_id", templateID).
						Msg("HookQuestionnaireAssessment: template not found")

					return nil, ErrTemplateNotFound
				}

				return nil, err
			}

			if tmpl.Kind != enums.TemplateKindQuestionnaire {
				logx.FromContext(ctx).Debug().
					Str("template_id", templateID).
					Str("kind", tmpl.Kind.String()).
					Msg("HookQuestionnaireAssessment: template is not a questionnaire, skipping")

				return nil, ErrTemplateNotQuestionnaire
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}
