package interceptors

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"entgo.io/ent"
	"github.com/theopenlane/gqlgen-plugins/graphutils"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/intercept"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/internal/httpserve/authmanager"
	"github.com/theopenlane/core/v2/pkg/shortlinks"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

// InterceptorAssessmentAccessURL populates the anonymous questionnaire access URL for system-owned assessments.
func InterceptorAssessmentAccessURL() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return intercept.AssessmentFunc(func(ctx context.Context, q *generated.AssessmentQuery) (generated.Value, error) {
			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			if !graphutils.CheckForRequestedField(ctx, "accessURL") {
				return v, nil
			}

			if assessments, ok := v.([]*generated.Assessment); ok {
				for _, assessment := range assessments {
					if err := setAssessmentAccessURL(ctx, assessment, q); err != nil {
						logx.FromContext(ctx).Warn().Err(err).Str("assessment_id", assessment.ID).Msg("failed to set assessment access URL")
					}
				}

				return v, nil
			}

			if assessment, ok := v.(*generated.Assessment); ok {
				if err := setAssessmentAccessURL(ctx, assessment, q); err != nil {
					logx.FromContext(ctx).Warn().Err(err).Str("assessment_id", assessment.ID).Msg("failed to set assessment access URL")
				}
			}

			return v, nil
		})
	})
}

func setAssessmentAccessURL(ctx context.Context, assessment *generated.Assessment, q *generated.AssessmentQuery) error {
	if assessment == nil || !assessment.SystemOwned {
		return nil
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	if q.TokenManager == nil {
		return fmt.Errorf("token manager is required") //nolint:err113
	}

	productURL := strings.TrimSpace(q.EntConfig.QuestionnaireProductURL)
	if productURL == "" {
		return fmt.Errorf("questionnaire product URL is required") //nolint:err113
	}

	baseURL, err := url.Parse(productURL + "/questionnaire")
	if err != nil {
		return fmt.Errorf("parse questionnaire URL: %w", err)
	}

	result, err := urlx.GenerateAnonTokenURL(ctx, q.TokenManager, q.Shortlinks, *baseURL, urlx.AnonTokenRequest{
		Prefix:    authmanager.AnonQuestionnaireJWTPrefix,
		SubjectID: ulids.New().String(),
		OrgID:     orgID,
		Purpose:   shortlinks.PurposeAssessmentAccess,
		Duration:  q.TokenManager.Config().AssessmentAccessDuration,
		ExtraClaims: func(c *tokens.Claims) {
			c.AssessmentID = assessment.ID
		},
	})
	if err != nil {
		return err
	}

	assessment.AccessURL = result.URL

	return nil
}

// InterceptorAssessmentBranding populates the branding field of an assessment based off the available modules
func InterceptorAssessmentBranding() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return intercept.AssessmentFunc(func(ctx context.Context, q *generated.AssessmentQuery) (generated.Value, error) {
			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			if !auth.IsAnonymousFromContext(ctx) {
				return v, nil
			}

			assessments, ok := v.([]*generated.Assessment)
			if !ok {
				return v, nil
			}

			ok, err = rule.HasFeature(ctx, models.CatalogBrandingAssessmentAddon.String())
			if err != nil {
				logx.FromContext(ctx).Warn().Err(err).
					Msg("failed to check if org has assessment branding addon")
			}

			for _, a := range assessments {
				a.RemoveBranding = ok
			}

			return v, nil
		})
	})
}
