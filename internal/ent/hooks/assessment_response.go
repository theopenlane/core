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
	"github.com/theopenlane/newman"
	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessment"
	"github.com/theopenlane/core/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/internal/ent/generated/campaigntarget"
	"github.com/theopenlane/core/internal/ent/generated/contact"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/pkg/logx"
)

// HookCreateAssessmentResponse sends the email to the user to fill in and input their data.
// It also makes sure to bump up the send attempts if needed. The hook is idempotent: multiple
// create calls for the same assessment/email/campaign combination will update the existing
// record rather than creating duplicates.
func HookCreateAssessmentResponse() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.AssessmentResponseFunc(func(ctx context.Context, m *generated.AssessmentResponseMutation) (generated.Value, error) {
			assessmentID, ok := m.AssessmentID()
			email, emailExists := m.Email()
			campaignID, _ := m.CampaignID()
			ownerID, _ := m.OwnerID()

			isTest := false
			if value, exists := m.IsTest(); exists {
				isTest = value
			}

			if !ok || !emailExists {
				m.ClearDocumentDataID()
				return next.Mutate(ctx, m)
			}

			if err := validateAndSetDueDate(ctx, m); err != nil {
				return nil, err
			}

			existingResponse, err := findExistingAssessmentResponse(ctx, m.Client(), assessmentID, email, campaignID, ownerID, isTest)
			if err != nil {
				return nil, err
			}

			if existingResponse != nil {
				return handleExistingAssessmentResponse(ctx, m, existingResponse, isTest)
			}

			return createNewAssessmentResponse(ctx, m, next, email)
		})
	}, ent.OpCreate)
}

// findExistingAssessmentResponse queries for an existing assessment response matching the given criteria.
func findExistingAssessmentResponse(ctx context.Context, client *generated.Client, assessmentID, email, campaignID, ownerID string, isTest bool) (*generated.AssessmentResponse, error) {
	query := client.AssessmentResponse.Query().
		Where(
			assessmentresponse.AssessmentID(assessmentID),
			assessmentresponse.EmailEqualFold(email),
			assessmentresponse.IsTestEQ(isTest),
		)

	if ownerID != "" {
		query = query.Where(assessmentresponse.OwnerID(ownerID))
	}

	if campaignID != "" {
		query = query.Where(assessmentresponse.CampaignIDEQ(campaignID))
	} else {
		query = query.Where(assessmentresponse.CampaignIDIsNil())
	}

	existingResponse, err := query.Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return existingResponse, nil
}

// handleExistingAssessmentResponse processes a resend request for an existing assessment response.
// It validates the response can be resent, increments send attempts, updates due date if provided,
// and sends a new email invitation.
func handleExistingAssessmentResponse(ctx context.Context, m *generated.AssessmentResponseMutation, existingResponse *generated.AssessmentResponse, isTest bool) (*generated.AssessmentResponse, error) {
	if existingResponse.Status != enums.AssessmentResponseStatusSent &&
		existingResponse.Status != enums.AssessmentResponseStatusOverdue &&
		!isTest {
		return nil, ErrAssessmentInProgress
	}

	newAttempts := existingResponse.SendAttempts + 1
	if newAttempts > maxAttempts {
		return nil, gqlerrors.NewCustomError(
			gqlerrors.MaxAttemptsErrorCode,
			"max attempts reached for this email, please delete the questionnaire request and try again",
			ErrMaxAttemptsAssessments)
	}

	update := m.Client().AssessmentResponse.UpdateOneID(existingResponse.ID).
		SetSendAttempts(newAttempts)

	if dueDate, ok := m.DueDate(); ok {
		update = update.SetDueDate(dueDate)
	}

	if existingResponse.Status == enums.AssessmentResponseStatusOverdue {
		update = update.SetStatus(enums.AssessmentResponseStatusSent)
	}

	updatedResponse, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	if err := createResponseEmail(ctx, m, updatedResponse.ID); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to resend assessment response email")
	}

	return updatedResponse, nil
}

// createNewAssessmentResponse creates a new assessment response record, handles contact creation,
// and sends the initial email invitation. It handles race conditions by detecting unique constraint
// violations and re-querying to update the existing record.
func createNewAssessmentResponse(ctx context.Context, m *generated.AssessmentResponseMutation, next ent.Mutator, email string) (generated.Value, error) {
	m.ClearDocumentDataID()

	count, err := m.Client().Contact.Query().Select(contact.FieldEmail).
		Where(contact.EmailEqualFold(email)).
		Count(ctx)
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("could not fetch existing contacts")
		return nil, ErrUnableToCreateContact
	}

	if count == 0 {
		if err := m.Client().Contact.Create().SetEmail(email).Exec(ctx); err != nil {
			logx.FromContext(ctx).Err(err).Msg("could not create contact for assessment response")
		}
	}

	value, err := next.Mutate(ctx, m)
	if err != nil {
		if isUniqueConstraintError(err) {
			return handleConstraintViolation(ctx, m)
		}

		return nil, err
	}

	var responseID string
	if resp, ok := value.(*generated.AssessmentResponse); ok && resp != nil {
		responseID = resp.ID
	}

	if err := createResponseEmail(ctx, m, responseID); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to send assessment response email")
	}

	return value, nil
}

// handleConstraintViolation handles the case where a unique constraint violation occurs during
// creation due to a race condition. It re-queries for the existing record and updates it.
func handleConstraintViolation(ctx context.Context, m *generated.AssessmentResponseMutation) (*generated.AssessmentResponse, error) {
	assessmentID, _ := m.AssessmentID()
	email, _ := m.Email()
	campaignID, _ := m.CampaignID()
	ownerID, _ := m.OwnerID()

	isTest := false
	if value, exists := m.IsTest(); exists {
		isTest = value
	}

	existingResponse, err := findExistingAssessmentResponse(ctx, m.Client(), assessmentID, email, campaignID, ownerID, isTest)
	if err != nil {
		return nil, err
	}

	if existingResponse == nil {
		return nil, ErrUnableToCreateAssessmentResponse
	}

	return handleExistingAssessmentResponse(ctx, m, existingResponse, isTest)
}

// isUniqueConstraintError checks if the error is a unique constraint violation.
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	return generated.IsConstraintError(err)
}

// createResponseEmail builds and queues the assessment response email for a recipient.
func createResponseEmail(ctx context.Context, m *generated.AssessmentResponseMutation, responseID string) error {
	orgID, _ := m.OwnerID()
	assessmentID, _ := m.AssessmentID()
	emailAddress, _ := m.Email()
	campaignID, _ := m.CampaignID()

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

	duration := m.Client().TokenManager.Config().AssessmentAccessDuration

	accessToken, _, err := m.TokenManager.CreateTokenPair(newClaims, tokens.WithAccessDuration(duration))
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

	tags := []newman.Tag{}
	if responseID != "" {
		tags = append(tags, newman.Tag{Name: "assessment_response_id", Value: responseID})
	}

	if campaignID != "" {
		tags = append(tags, newman.Tag{Name: "campaign_id", Value: campaignID})
	}

	if isTest, ok := m.IsTest(); ok && isTest {
		tags = append(tags, newman.Tag{Name: "is_test", Value: "true"})
	}

	if ctxData, ok := CampaignEmailContextFrom(ctx); ok {
		if ctxData.CampaignID != "" && campaignID == "" {
			tags = append(tags, newman.Tag{Name: "campaign_id", Value: ctxData.CampaignID})
		}
		if ctxData.CampaignTargetID != "" {
			tags = append(tags, newman.Tag{Name: "campaign_target_id", Value: ctxData.CampaignTargetID})
		}
	}

	if len(tags) > 0 {
		email.Tags = append(email.Tags, tags...)
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

			newStatus, statusBeingSet := m.Status()
			value, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			updatedResp, ok := value.(*generated.AssessmentResponse)
			if !ok || updatedResp == nil {
				return value, nil
			}

			if updatedResp.CampaignID != "" && !updatedResp.IsTest && statusBeingSet {
				if err := updateCampaignTargetFromAssessmentResponse(ctx, m.Client(), updatedResp); err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to update campaign target status from assessment response")
				}

				if newStatus == enums.AssessmentResponseStatusCompleted {
					if err := updateCampaignCompletionFromTargets(ctx, m.Client(), updatedResp.CampaignID); err != nil {
						logx.FromContext(ctx).Error().Err(err).Msg("failed to update campaign completion status")
					}
				}
			}

			return value, nil
		})
	}, ent.OpUpdateOne)
}

// updateCampaignTargetFromAssessmentResponse mirrors response status into the campaign target record.
func updateCampaignTargetFromAssessmentResponse(ctx context.Context, client *generated.Client, resp *generated.AssessmentResponse) error {
	if resp == nil || resp.CampaignID == "" {
		return nil
	}

	update := client.CampaignTarget.Update().
		Where(
			campaigntarget.CampaignIDEQ(resp.CampaignID),
			campaigntarget.EmailEqualFold(resp.Email),
		)

	switch resp.Status {
	case enums.AssessmentResponseStatusCompleted:
		update.SetStatus(enums.AssessmentResponseStatusCompleted)
		if !resp.CompletedAt.IsZero() {
			update.SetCompletedAt(models.DateTime(resp.CompletedAt))
		} else {
			update.SetCompletedAt(models.DateTime(time.Now()))
		}
	case enums.AssessmentResponseStatusOverdue:
		update.SetStatus(enums.AssessmentResponseStatusOverdue)
	default:
		return nil
	}

	return update.Exec(ctx)
}

// updateCampaignCompletionFromTargets marks campaigns complete when all targets are completed.
func updateCampaignCompletionFromTargets(ctx context.Context, client *generated.Client, campaignID string) error {
	if campaignID == "" {
		return nil
	}

	campaignObj, err := client.Campaign.Get(ctx, campaignID)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil
		}
		return err
	}

	if campaignObj.Status == enums.CampaignStatusCompleted || campaignObj.Status == enums.CampaignStatusCanceled {
		return nil
	}

	total, err := client.CampaignTarget.Query().
		Where(campaigntarget.CampaignIDEQ(campaignID)).
		Count(ctx)
	if err != nil {
		return err
	}

	if total == 0 {
		return nil
	}

	completed, err := client.CampaignTarget.Query().
		Where(
			campaigntarget.CampaignIDEQ(campaignID),
			campaigntarget.StatusEQ(enums.AssessmentResponseStatusCompleted),
		).
		Count(ctx)
	if err != nil {
		return err
	}

	if completed != total {
		return nil
	}

	update := client.Campaign.UpdateOneID(campaignID).
		SetStatus(enums.CampaignStatusCompleted).
		SetIsActive(false)

	if campaignObj.CompletedAt == nil || campaignObj.CompletedAt.IsZero() {
		update.SetCompletedAt(models.DateTime(time.Now()))
	}

	return update.Exec(ctx)
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
