package graphapi

import (
	"context"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/campaign"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/utils/rout"
)

const (
	// maxTestEmailRecipients caps how many recipients a single test email request may target
	maxTestEmailRecipients = 5
	// maxTestEmailsPerWindow caps how many test emails a campaign may send per rate limit window
	maxTestEmailsPerWindow = 10
	// testEmailRateLimitWindow is the rolling window for the per-campaign test email cap
	testEmailRateLimitWindow = time.Hour
	// testEmailRateLimitKeyPrefix scopes per-campaign test email allowances in redis
	testEmailRateLimitKeyPrefix = "campaign-test-email:"
)

// testEmailState holds state for processing test email requests
type testEmailState struct {
	client      *generated.Client
	rt          *intruntime.Runtime
	campaignObj *generated.Campaign
	emails      []string
	queued      int
	skipped     int
}

// initTestEmailState validates input, dedupes recipients, and enforces the test email rate limits
func (r *mutationResolver) initTestEmailState(ctx context.Context, input model.SendCampaignTestEmailInput) (*testEmailState, error) {
	if input.CampaignID == "" {
		return nil, rout.NewMissingRequiredFieldError("campaignID")
	}

	if len(input.Emails) == 0 {
		return nil, rout.NewMissingRequiredFieldError("emails")
	}

	emails, skipped := dedupeTestEmails(input.Emails)
	if skipped > 0 {
		logx.FromContext(ctx).Debug().Str("campaign_id", input.CampaignID).Int("skipped", skipped).Msg("dropped blank or duplicate test email recipients")
	}

	if len(emails) > maxTestEmailRecipients {
		logx.FromContext(ctx).Info().Str("campaign_id", input.CampaignID).Int("requested", len(emails)).Int("limit", maxTestEmailRecipients).Msg("test email request exceeds recipient cap, rejecting")

		return nil, ErrCampaignTestEmailTooManyRecipients
	}

	client := withTransactionalMutation(ctx)

	campaignObj, err := client.Campaign.Get(ctx, input.CampaignID)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "campaign"})
	}

	if err := r.validateTestEmailCampaign(ctx, campaignObj); err != nil {
		return nil, err
	}

	rt := intruntime.FromClient(ctx, client)
	if rt == nil {
		return nil, ErrCampaignDispatchRuntimeRequired
	}

	allowed, err := rt.AllowN(ctx, testEmailRateLimitKeyPrefix+campaignObj.ID, len(emails), maxTestEmailsPerWindow, testEmailRateLimitWindow)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", campaignObj.ID).Msg("failed checking test email rate limit")

		return nil, err
	}

	if !allowed {
		logx.FromContext(ctx).Info().Str("campaign_id", campaignObj.ID).Int("requested", len(emails)).Int("limit", maxTestEmailsPerWindow).Msg("campaign test email hourly limit reached, rejecting")

		return nil, ErrCampaignTestEmailRateLimited
	}

	return &testEmailState{
		client:      client,
		rt:          rt,
		campaignObj: campaignObj,
		emails:      emails,
		skipped:     skipped,
	}, nil
}

// dedupeTestEmails trims whitespace, drops blanks, and case-insensitively dedupes recipients,
// returning the send list and the count removed
func dedupeTestEmails(emails []string) ([]string, int) {
	trimmed := lo.FilterMap(emails, func(email string, _ int) (string, bool) {
		t := strings.TrimSpace(email)

		return t, t != ""
	})

	unique := lo.UniqBy(trimmed, strings.ToLower)

	return unique, len(emails) - len(unique)
}

// validateTestEmailCampaign checks permissions and campaign eligibility for test emails
func (r *mutationResolver) validateTestEmailCampaign(ctx context.Context, campaignObj *generated.Campaign) error {
	ctx, err := common.SetOrganizationInAuthContext(ctx, &campaignObj.OwnerID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to set organization in auth context")
		return rout.ErrPermissionDenied
	}

	if err := r.ensureCampaignEditAccess(ctx, campaignObj.ID); err != nil {
		return err
	}

	if err := ensureCampaignAssessment(ctx, campaignObj); err != nil {
		return err
	}

	return validateCampaignContentSource(campaignObj)
}

// processTestEmails dispatches a test email for each deduped recipient
func (r *mutationResolver) processTestEmails(ctx context.Context, state *testEmailState) error {
	for _, email := range state.emails {
		req, err := r.buildCampaignEmailDispatchRequest(ctx, state.rt, state.campaignObj, false, false, email, nil)
		if err != nil {
			return err
		}

		if _, err := state.rt.Dispatch(ctx, req); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("campaign_id", state.campaignObj.ID).Msg("failed dispatching campaign test email")

			return err
		}

		state.queued++
	}

	return nil
}

// buildTestEmailPayload fetches the final campaign state and constructs the response
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
