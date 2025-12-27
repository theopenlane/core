package hooks

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessment"
	"github.com/theopenlane/core/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/internal/ent/generated/contact"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/pkg/logx"
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

			if err := validateAndSetDueDate(ctx, m); err != nil {
				return nil, err
			}

			existingResponse, err := m.Client().AssessmentResponse.Query().
				Where(
					assessmentresponse.AssessmentID(id),
					assessmentresponse.EmailEqualFold(email),
				).
				Only(ctx)

			if err != nil {
				if !generated.IsNotFound(err) {
					return nil, err
				}

				// not found so this is a new user
				m.ClearDocumentDataID()

				// try to make email unique per org
				count, err := m.Client().Contact.Query().Select(contact.FieldEmail).
					Where(contact.EmailEqualFold(email)).
					Count(ctx)
				if err != nil {
					logx.FromContext(ctx).Err(err).Msg("could not fetch existing contacts")
					return nil, ErrUnableToCreateContact
				}

				if count == 0 {
					err = m.Client().Contact.Create().SetEmail(email).
						Exec(ctx)
					if err != nil {
						logx.FromContext(ctx).Err(err).Msg("could not create contact for assessment response")
					}
				}

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
					"max attempts reached for this email, please delete the questionnaire request and try again",
					ErrMaxAttemptsAssessments)
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

// HookUpdateAssessmentResponse checks if the assessment response is past due and updates the status accordingly
func HookUpdateAssessmentResponse() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.AssessmentResponseFunc(func(ctx context.Context, m *generated.AssessmentResponseMutation) (generated.Value, error) {
			id, exists := m.ID()
			if !exists {
				return next.Mutate(ctx, m)
			}

			assessmentResp, err := m.Client().AssessmentResponse.Get(ctx, id)
			if err != nil {
				return nil, err
			}

			if !assessmentResp.DueDate.IsZero() && time.Now().After(assessmentResp.DueDate) {
				newStatus, statusBeingSet := m.Status()

				shouldSetOverdue := (statusBeingSet && newStatus == enums.AssessmentResponseStatusCompleted) ||
					(!statusBeingSet && assessmentResp.Status != enums.AssessmentResponseStatusOverdue)

				if shouldSetOverdue {
					m.SetStatus(enums.AssessmentResponseStatusOverdue)
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpUpdateOne)
}

// validateAndSetDueDate validates the due_date field and sets it based on the assessment's response_due_duration if not provided
func validateAndSetDueDate(ctx context.Context, m *generated.AssessmentResponseMutation) error {
	dueDate, ok := m.DueDate()

	// if due_date is provided, validate it's not in the past
	if ok {
		if dueDate.Before(time.Now()) {
			return ErrPastTimeNotAllowed
		}
		return nil
	}

	// if due date is not provided, calculate it from the parent assessment
	id, ok := m.AssessmentID()
	if !ok {
		return nil
	}

	assessmentDB, err := m.Client().Assessment.Query().
		Where(assessment.ID(id)).
		Select(assessment.FieldResponseDueDuration).
		Only(ctx)
	if err != nil {
		return err
	}

	if assessmentDB.ResponseDueDuration > 0 {
		duration := time.Duration(assessmentDB.ResponseDueDuration) * time.Second
		calculatedDueDate := time.Now().Add(duration)
		m.SetDueDate(calculatedDueDate)
	}

	return nil
}
