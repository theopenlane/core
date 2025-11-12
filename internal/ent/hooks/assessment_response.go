package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessment"
	"github.com/theopenlane/core/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/riverboat/pkg/jobs"
)

// HookCreateAssessmentResponse sends the email to the user to fill in and input their data.
// It also makes sure to bump up the send attempts if needed.
func HookCreateAssessmentResponse() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.AssessmentResponseFunc(func(ctx context.Context, m *generated.AssessmentResponseMutation) (generated.Value, error) {
			id, ok := m.AssessmentID()
			email, emailExists := m.Email()

			if !ok || !emailExists {
				m.ClearDocumentDataID()
				return next.Mutate(ctx, m)
			}

			existingResponse, err := m.Client().AssessmentResponse.Query().
				Where(
					assessmentresponse.AssessmentID(id),
					assessmentresponse.Email(email),
				).
				Only(ctx)

			if err != nil {
				if !generated.IsNotFound(err) {
					return nil, err
				}

				// not found so this is a new user
				m.ClearDocumentDataID()

				value, err := next.Mutate(ctx, m)
				if err != nil {
					return nil, err
				}

				if err := createResponseEmail(ctx, m); err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to send assessment response email")
				}

				return value, nil
			}

			// if started, we cannot send an invitation again
			if existingResponse.Status != enums.AssessmentResponseStatusSent {
				return nil, ErrAssessmentInProgress
			}

			newAttempts := existingResponse.SendAttempts + 1
			if newAttempts > maxAttempts {
				return nil, gqlerrors.NewCustomError(
					gqlerrors.MaxAttemptsErrorCode,
					"max attempts reached for this email, please delete the invite and try again",
					ErrMaxAttempts)
			}

			updatedResponse, err := m.Client().AssessmentResponse.UpdateOneID(existingResponse.ID).
				SetSendAttempts(newAttempts).
				Save(ctx)
			if err != nil {
				return nil, err
			}

			if err := createResponseEmail(ctx, m); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to resend assessment response email")
			}

			return updatedResponse, nil
		})
	}, ent.OpCreate)
}

func createResponseEmail(ctx context.Context, m *generated.AssessmentResponseMutation) error {
	orgID, _ := m.OwnerID()
	assessmentID, _ := m.AssessmentID()
	emailAddress, _ := m.Email()

	org, err := m.Client().Organization.Query().
		Where(organization.ID(orgID)).
		Select(organization.FieldDisplayName).
		Only(ctx)
	if err != nil {
		return err
	}

	assessmentData, err := m.Client().Assessment.Query().
		Where(assessment.ID(assessmentID)).
		Select(assessment.FieldName).
		Only(ctx)
	if err != nil {
		return err
	}

	anonUserID := fmt.Sprintf("%s%s", authmanager.AnonQuestionnaireJWTPrefix, uuid.New().String())

	newClaims := &tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: anonUserID,
		},
		UserID:       anonUserID,
		OrgID:        orgID,
		AssessmentID: assessmentID,
		Email:        emailAddress,
	}

	accessToken, _, err := m.TokenManager.CreateTokenPair(newClaims)
	if err != nil {
		return err
	}

	email, err := m.Emailer.NewQuestionnaireAuthEmail(emailtemplates.Recipient{
		Email: emailAddress,
	}, accessToken, emailtemplates.QuestionnaireAuthData{
		CompanyName:    org.DisplayName,
		AssessmentName: assessmentData.Name,
	})
	if err != nil {
		return err
	}

	_, err = m.Job.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil)

	return err
}
