package graphapi

import (
	"context"
	"strings"
	"time"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/campaign"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/utils/rout"
)

// testEmailState holds state for processing test email requests.
type testEmailState struct {
	client      *generated.Client
	campaignObj *generated.Campaign
	seenEmails  map[string]struct{}
	queued      int
	skipped     int
}

// initTestEmailState validates input and initializes the test email state.
func (r *mutationResolver) initTestEmailState(ctx context.Context, input model.SendCampaignTestEmailInput) (*testEmailState, error) {
	if input.CampaignID == "" {
		return nil, rout.NewMissingRequiredFieldError("campaignID")
	}

	if len(input.Emails) == 0 {
		return nil, rout.NewMissingRequiredFieldError("emails")
	}

	client := withTransactionalMutation(ctx)
	campaignObj, err := client.Campaign.Get(ctx, input.CampaignID)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "campaign"})
	}

	if err := r.validateTestEmailCampaign(ctx, campaignObj); err != nil {
		return nil, err
	}

	return &testEmailState{
		client:      client,
		campaignObj: campaignObj,
		seenEmails:  make(map[string]struct{}),
	}, nil
}

// validateTestEmailCampaign checks permissions and campaign eligibility for test emails.
func (r *mutationResolver) validateTestEmailCampaign(ctx context.Context, campaignObj *generated.Campaign) error {
	if err := common.SetOrganizationInAuthContext(ctx, &campaignObj.OwnerID); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to set organization in auth context")
		return rout.ErrPermissionDenied
	}

	if err := r.ensureCampaignEditAccess(ctx, campaignObj.ID); err != nil {
		return err
	}

	if campaignObj.CampaignType != enums.CampaignTypeQuestionnaire {
		return ErrCampaignTestEmailNotQuestionnaire
	}

	if campaignObj.AssessmentID == "" {
		return ErrCampaignMissingAssessmentID
	}

	return nil
}

// processTestEmails iterates through emails and creates test assessment responses.
func (r *mutationResolver) processTestEmails(ctx context.Context, state *testEmailState, emails []string) error {
	for _, email := range emails {
		if err := r.processTestEmail(ctx, state, email); err != nil {
			return err
		}
	}
	return nil
}

// processTestEmail handles a single test email, deduplicating and creating the response.
func (r *mutationResolver) processTestEmail(ctx context.Context, state *testEmailState, email string) error {
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		state.skipped++
		return nil
	}

	key := strings.ToLower(trimmed)
	if _, exists := state.seenEmails[key]; exists {
		state.skipped++
		return nil
	}
	state.seenEmails[key] = struct{}{}

	create := state.client.AssessmentResponse.Create().
		SetAssessmentID(state.campaignObj.AssessmentID).
		SetCampaignID(state.campaignObj.ID).
		SetEmail(trimmed).
		SetIsTest(true)

	if state.campaignObj.EntityID != "" {
		create.SetEntityID(state.campaignObj.EntityID)
	}

	if state.campaignObj.DueDate != nil && !state.campaignObj.DueDate.IsZero() {
		create.SetDueDate(time.Time(*state.campaignObj.DueDate))
	}

	if err := create.Exec(ctx); err != nil {
		return parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "assessmentresponse"})
	}

	state.queued++
	return nil
}

// buildTestEmailPayload fetches the final campaign state and constructs the response.
func (r *mutationResolver) buildTestEmailPayload(ctx context.Context, state *testEmailState) (*model.CampaignTestEmailPayload, error) {
	query, err := state.client.Campaign.Query().
		Where(campaign.ID(state.campaignObj.ID)).
		CollectFields(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "campaign"})
	}

	finalResult, err := query.Only(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "campaign"})
	}

	return &model.CampaignTestEmailPayload{
		Campaign:     finalResult,
		QueuedCount:  state.queued,
		SkippedCount: state.skipped,
	}, nil
}
