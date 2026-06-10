package interceptors

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"

	"github.com/theopenlane/gqlgen-plugins/graphutils"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integrationrun"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/pkg/logx"
)

// InterceptorAssessmentResponseError populates the latest questionnaire transform error when requested.
func InterceptorAssessmentResponseError() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return intercept.AssessmentResponseFunc(func(ctx context.Context, q *generated.AssessmentResponseQuery) (generated.Value, error) {
			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			if !graphutils.CheckForRequestedField(ctx, "questionnaireTransformError") {
				return v, nil
			}

			var responses []*generated.AssessmentResponse

			if resp, ok := v.([]*generated.AssessmentResponse); ok {
				responses = append(responses, resp...)
			}

			if resp, ok := v.(*generated.AssessmentResponse); ok && resp != nil {
				responses = append(responses, resp)
			}

			if err := setErrorFromIntegrationRuns(ctx, responses); err != nil {
				logx.FromContext(ctx).Warn().Err(err).Msg("failed to set questionnaire transform error")
			}

			return v, nil
		})
	})
}

func setErrorFromIntegrationRuns(ctx context.Context, responses []*generated.AssessmentResponse) error {
	client := generated.FromContext(ctx)
	if client == nil {
		logx.FromContext(ctx).Debug().Msg("ent client not found in context, skipping questionnaire transform error")

		return nil
	}

	responseByID := make(map[string]*generated.AssessmentResponse, len(responses))
	ids := make([]string, 0, len(responses))

	for _, response := range responses {
		if response == nil || response.ID == "" {
			continue
		}

		if _, ok := responseByID[response.ID]; ok {
			continue
		}

		responseByID[response.ID] = response
		ids = append(ids, response.ID)
	}

	if len(ids) == 0 {
		return nil
	}

	integrationRuns, err := client.IntegrationRun.Query().
		Where(
			integrationrun.AssessmentResponseIDIn(ids...),
			integrationrun.OperationNameEQ(operations.QuestionnaireTransformOperationName),
		).
		Order(integrationrun.ByStartedAt(sql.OrderDesc())).
		All(ctx)
	if err != nil {
		return err
	}

	seen := make(map[string]struct{}, len(responseByID))
	for _, run := range integrationRuns {
		if run == nil || run.AssessmentResponseID == "" {
			continue
		}

		if _, ok := seen[run.AssessmentResponseID]; ok {
			continue
		}

		response, ok := responseByID[run.AssessmentResponseID]
		if !ok {
			continue
		}

		response.QuestionnaireTransformError = run.Error
		seen[run.AssessmentResponseID] = struct{}{}
	}

	return nil
}
