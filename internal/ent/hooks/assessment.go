package hooks

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/core/pkg/logx"
)

// HookQuestionnaireAssessment is a hook that checks if the templatate associated with the assessment is a questionnaire
func HookQuestionnaireAssessment() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.AssessmentFunc(func(ctx context.Context, m *generated.AssessmentMutation) (generated.Value, error) {

			id, ok := m.TemplateID()
			if !ok {
				// user provided jsonconfig and uischema directly then
				// and not trying to be created/cloned from a template
				//
				// but at least the jsonconfig needs to be provided
				_, ok := m.Jsonconfig()
				if !ok {
					return nil, fmt.Errorf("jsonconfig is required if you do not create an assessment from a template") //nolint:err113
				}
				return next.Mutate(ctx, m)
			}

			// if a template was provided, validate it is a questionnaire type
			// and inherit the jsonconfig and uischema
			template, err := m.Client().Template.Query().
				Where(template.ID(id)).
				Only(ctx)
			if err != nil {
				if generated.IsNotFound(err) {
					logx.FromContext(ctx).Warn().Str("template_id", id).
						Msg("template not found")

					return nil, ErrTemplateNotFound
				}

				return nil, err
			}

			if template.Kind != enums.TemplateKindQuestionnaire {
				logx.FromContext(ctx).
					Err(errors.New("template is not of type questionnaire")). //nolint:err113
					Str("template_id", id).Str("kind", template.Kind.String()).
					Msg("template is not a questionnaire type")

				return nil, ErrTemplateNotQuestionnaire
			}

			m.SetUischema(template.Uischema)
			m.SetJsonconfig(template.Jsonconfig)

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}
