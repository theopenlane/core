package hooks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/pkg/logx"
)

// HookQuestionnaireAssessment is a hook that checks if the templatate associated with the assessment is a questionnaire
func HookQuestionnaireAssessment() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.AssessmentFunc(func(ctx context.Context, m *generated.AssessmentMutation) (generated.Value, error) {

			syncOptions := assessmentSyncOptions{}

			syncOptions.dueDuration, syncOptions.dueDurationSet = m.ResponseDueDuration()
			syncOptions.dueDurationCleared = m.ResponseDueDurationCleared()

			if syncOptions.dueDurationSet || syncOptions.dueDurationCleared {
				switch {
				case m.Op().Is(ent.OpUpdateOne):
					if id, ok := m.ID(); ok {
						syncOptions.assessmentIDs = []string{id}
					}
				case m.Op().Is(ent.OpUpdate):
					var err error
					syncOptions.assessmentIDs, err = m.IDs(ctx)
					if err != nil {
						return nil, err
					}
				}
			}

			id, ok := m.TemplateID()
			if !ok {
				if m.Op().Is(ent.OpCreate) {
					// user provided jsonconfig and uischema directly then
					// and not trying to be created/cloned from a template
					//
					// but at least the jsonconfig needs to be provided
					_, ok := m.Jsonconfig()
					if !ok {
						return nil, fmt.Errorf("jsonconfig is required if you do not create an assessment from a template") //nolint:err113
					}
				}

				value, err := next.Mutate(ctx, m)
				if err != nil {
					return value, err
				}

				if err := syncDueDateChanges(ctx, m.Client(), syncOptions); err != nil {
					return value, err
				}

				return value, nil
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

			if template.Kind == enums.TemplateKindTrustCenterNda {
				logx.FromContext(ctx).
					Err(errors.New("template is not of type questionnaire")). //nolint:err113
					Str("template_id", id).Str("kind", template.Kind.String()).
					Msg("template is not a questionnaire type")

				return nil, ErrTemplateNotQuestionnaire
			}

			m.SetUischema(template.Uischema)
			m.SetJsonconfig(template.Jsonconfig)

			value, err := next.Mutate(ctx, m)
			if err != nil {
				return value, err
			}

			if err := syncDueDateChanges(ctx, m.Client(), syncOptions); err != nil {
				return value, err
			}

			return value, nil
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

type assessmentSyncOptions struct {
	assessmentIDs      []string
	dueDuration        int64
	dueDurationSet     bool
	dueDurationCleared bool
}

// syncDueDateChanges syncs due date changes from the root assessment to all existing responses
func syncDueDateChanges(ctx context.Context, client *generated.Client, opts assessmentSyncOptions) error {
	if len(opts.assessmentIDs) == 0 || (!opts.dueDurationSet && !opts.dueDurationCleared) {
		return nil
	}

	update := client.AssessmentResponse.Update().
		Where(
			assessmentresponse.AssessmentIDIn(opts.assessmentIDs...),
		)

	update.SetDueDate(time.Now().Add(time.Duration(opts.dueDuration) * time.Second))

	if opts.dueDurationCleared || opts.dueDuration <= 0 {
		update.ClearDueDate()
	}

	return update.Exec(ctx)
}
