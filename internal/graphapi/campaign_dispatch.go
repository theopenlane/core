package graphapi

import (
	"context"
	"encoding/json"
	"time"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/campaign"
	"github.com/theopenlane/core/internal/ent/generated/campaigntarget"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/integrations/operations"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/utils/rout"
)

// Campaign dispatch action aliases keep gql layer aligned with jobspec values.
const (
	campaignDispatchActionLaunch           = jobspec.CampaignDispatchActionLaunch
	campaignDispatchActionResend           = jobspec.CampaignDispatchActionResend
	campaignDispatchActionResendIncomplete = jobspec.CampaignDispatchActionResendIncomplete
)

// campaignDispatchOptions captures dispatch intent and scheduling input.
type campaignDispatchOptions struct {
	Action      jobspec.CampaignDispatchAction
	ScheduledAt *models.DateTime
}

// campaignDispatchState holds state accumulated during dispatch processing.
type campaignDispatchState struct {
	campaignObj    *generated.Campaign
	opts           campaignDispatchOptions
	resend         bool
	includeOverdue bool
	scheduleAt     *time.Time
	shouldSchedule bool
	now            time.Time
	queuedCount    int
	skippedCount   int
}

// dispatchCampaign coordinates immediate or scheduled dispatch of a campaign.
func (r *mutationResolver) dispatchCampaign(ctx context.Context, campaignID string, opts campaignDispatchOptions) (*model.CampaignLaunchPayload, error) {
	state, err := r.initCampaignDispatch(ctx, campaignID, opts)
	if err != nil {
		return nil, err
	}

	if err := r.processDispatchTargets(ctx, state); err != nil {
		return nil, err
	}

	if err := r.updateCampaignAfterDispatch(ctx, state); err != nil {
		return nil, err
	}

	return r.buildDispatchPayload(ctx, state)
}

// initCampaignDispatch validates and initializes the dispatch state.
func (r *mutationResolver) initCampaignDispatch(ctx context.Context, campaignID string, opts campaignDispatchOptions) (*campaignDispatchState, error) {
	if campaignID == "" {
		return nil, rout.NewMissingRequiredFieldError("campaignID")
	}

	campaignObj, err := withTransactionalMutation(ctx).Campaign.Query().Where(campaign.ID(campaignID)).Only(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "campaign"})
	}

	if err := r.validateCampaignDispatch(ctx, campaignObj); err != nil {
		return nil, err
	}

	resend, includeOverdue, err := campaignDispatchBehavior(opts.Action)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	scheduleAt, err := resolveCampaignScheduleAt(now, campaignObj, opts.Action, opts.ScheduledAt)
	if err != nil {
		return nil, err
	}

	return &campaignDispatchState{
		campaignObj:    campaignObj,
		opts:           opts,
		resend:         resend,
		includeOverdue: includeOverdue,
		scheduleAt:     scheduleAt,
		shouldSchedule: scheduleAt != nil && scheduleAt.After(now),
		now:            now,
	}, nil
}

// validateCampaignDispatch checks permissions and campaign state for dispatch eligibility.
func (r *mutationResolver) validateCampaignDispatch(ctx context.Context, campaignObj *generated.Campaign) error {
	ctx, err := common.SetOrganizationInAuthContext(ctx, &campaignObj.OwnerID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to set organization in auth context")
		return rout.ErrPermissionDenied
	}

	if err := r.ensureCampaignEditAccess(ctx, campaignObj.ID); err != nil {
		return err
	}

	if campaignObj.Status == enums.CampaignStatusCompleted || campaignObj.Status == enums.CampaignStatusCanceled {
		return ErrCampaignDispatchInactive
	}

	return nil
}

// processDispatchTargets counts dispatchable targets and, for immediate dispatch,
// delegates to the appropriate email operation via the integration runtime
func (r *mutationResolver) processDispatchTargets(ctx context.Context, state *campaignDispatchState) error {
	targets, err := withTransactionalMutation(ctx).CampaignTarget.Query().
		Where(campaigntarget.CampaignIDEQ(state.campaignObj.ID)).
		All(ctx)
	if err != nil {
		return parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "campaigntarget"})
	}

	for _, target := range targets {
		if !emaildef.TargetDispatchable(target.Status, target.SentAt, state.resend, state.includeOverdue) {
			state.skippedCount++
			continue
		}

		state.queuedCount++
	}

	// for scheduled dispatch, targets are counted only — the job worker dispatches later
	if state.shouldSchedule || state.queuedCount == 0 {
		return nil
	}

	return r.dispatchCampaignOperation(ctx, state)
}

// dispatchCampaignOperation dispatches the campaign email operation through the
// integration runtime. It performs an active lookup for the email integration at
// dispatch time so integrations created after campaign creation are picked up
func (r *mutationResolver) dispatchCampaignOperation(ctx context.Context, state *campaignDispatchState) error {
	rt := intruntime.FromClient(ctx, withTransactionalMutation(ctx))
	if rt == nil {
		return ErrCampaignDispatchRuntimeRequired
	}

	req, err := r.buildCampaignEmailDispatchRequest(ctx, state.campaignObj, state.resend, state.includeOverdue, "", nil)
	if err != nil {
		return err
	}

	if _, err := rt.Dispatch(ctx, req); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", state.campaignObj.ID).Msg("failed dispatching campaign operation")

		return err
	}

	return nil
}

// updateCampaignAfterDispatch updates campaign status based on scheduled vs immediate dispatch.
func (r *mutationResolver) updateCampaignAfterDispatch(ctx context.Context, state *campaignDispatchState) error {
	if state.shouldSchedule {
		return r.updateCampaignForScheduledDispatch(ctx, state)
	}
	return r.updateCampaignForImmediateDispatch(ctx, state)
}

// updateCampaignForScheduledDispatch enqueues the job and sets scheduled status.
func (r *mutationResolver) updateCampaignForScheduledDispatch(ctx context.Context, state *campaignDispatchState) error {
	if err := r.enqueueCampaignDispatchJob(ctx, state); err != nil {
		return parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "campaign"})
	}

	if state.opts.Action != campaignDispatchActionLaunch {
		return nil
	}

	update := withTransactionalMutation(ctx).Campaign.UpdateOneID(state.campaignObj.ID).
		SetStatus(enums.CampaignStatusScheduled).
		SetIsActive(false)

	if state.scheduleAt != nil {
		update.SetScheduledAt(models.DateTime(*state.scheduleAt))
	}

	if err := update.Exec(ctx); err != nil {
		return parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "campaign"})
	}

	return nil
}

// updateCampaignForImmediateDispatch sets the campaign to active and updates timestamps.
func (r *mutationResolver) updateCampaignForImmediateDispatch(ctx context.Context, state *campaignDispatchState) error {
	update := withTransactionalMutation(ctx).Campaign.UpdateOneID(state.campaignObj.ID).
		SetStatus(enums.CampaignStatusActive).
		SetIsActive(true)

	if state.campaignObj.LaunchedAt == nil || state.campaignObj.LaunchedAt.IsZero() {
		update.SetLaunchedAt(models.DateTime(state.now))
	}

	if state.resend && state.queuedCount > 0 {
		update.SetLastResentAt(models.DateTime(state.now)).
			AddResendCount(1)
	}

	if err := update.Exec(ctx); err != nil {
		return parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "campaign"})
	}

	return nil
}

// buildDispatchPayload fetches the final campaign state and constructs the response.
func (r *mutationResolver) buildDispatchPayload(ctx context.Context, state *campaignDispatchState) (*model.CampaignLaunchPayload, error) {
	query, err := withTransactionalMutation(ctx).Campaign.Query().
		Where(campaign.ID(state.campaignObj.ID)).
		CollectFields(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "campaign"})
	}

	finalResult, err := query.Only(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "campaign"})
	}

	return &model.CampaignLaunchPayload{
		Campaign:     finalResult,
		QueuedCount:  state.queuedCount,
		SkippedCount: state.skippedCount,
	}, nil
}

// campaignDispatchBehavior resolves resend and overdue handling based on the action.
func campaignDispatchBehavior(action jobspec.CampaignDispatchAction) (bool, bool, error) {
	switch action {
	case campaignDispatchActionLaunch:
		return false, false, nil
	case campaignDispatchActionResend:
		return true, false, nil
	case campaignDispatchActionResendIncomplete:
		return true, true, nil
	default:
		return false, false, ErrCampaignDispatchActionUnsupported
	}
}

// resolveCampaignScheduleAt chooses the effective schedule time for a dispatch action.
func resolveCampaignScheduleAt(now time.Time, campaignObj *generated.Campaign, action jobspec.CampaignDispatchAction, input *models.DateTime) (*time.Time, error) {
	if input != nil && !input.IsZero() {
		scheduled := time.Time(*input)
		if scheduled.Before(now) {
			return nil, ErrCampaignDispatchScheduledAtInPast
		}

		return &scheduled, nil
	}

	if action == campaignDispatchActionLaunch && campaignObj != nil && campaignObj.ScheduledAt != nil && !campaignObj.ScheduledAt.IsZero() {
		scheduled := time.Time(*campaignObj.ScheduledAt)
		if scheduled.After(now) {
			return &scheduled, nil
		}
	}

	return nil, nil
}

func (r *mutationResolver) buildCampaignEmailDispatchRequest(ctx context.Context, campaignObj *generated.Campaign, resend bool, includeOverdue bool, testEmail string, scheduledAt *time.Time) (operations.DispatchRequest, error) {
	if campaignObj == nil {
		return operations.DispatchRequest{}, emaildef.ErrCampaignNotFound
	}

	operation := emaildef.SendBrandedCampaignOp.Name()
	config, err := json.Marshal(emaildef.SendBrandedCampaignRequest{
		CampaignID:     campaignObj.ID,
		Resend:         resend,
		IncludeOverdue: includeOverdue,
	})
	if campaignObj.CampaignType == enums.CampaignTypeQuestionnaire {
		operation = emaildef.SendQuestionnaireCampaignOp.Name()
		config, err = json.Marshal(emaildef.SendQuestionnaireCampaignRequest{
			CampaignID:     campaignObj.ID,
			Resend:         resend,
			IncludeOverdue: includeOverdue,
			TestEmail:      testEmail,
		})
	}
	if err != nil {
		return operations.DispatchRequest{}, err
	}

	req := operations.DispatchRequest{
		Operation:   operation,
		Config:      config,
		RunType:     enums.IntegrationRunTypeEvent,
		ScheduledAt: scheduledAt,
	}

	integrationID := emaildef.ResolveCampaignEmailIntegration(ctx, withTransactionalMutation(ctx), campaignObj.OwnerID, campaignObj.IntegrationID)
	if integrationID != "" {
		req.IntegrationID = integrationID
		return req, nil
	}

	req.DefinitionID = emaildef.DefinitionID.ID()
	req.OwnerID = campaignObj.OwnerID
	req.Runtime = true

	return req, nil
}

// enqueueCampaignDispatchJob schedules a campaign dispatch through the integration
// framework with deferred execution at the specified time. Uses the customer
// integration when available, falling back to the runtime provider
func (r *mutationResolver) enqueueCampaignDispatchJob(ctx context.Context, state *campaignDispatchState) error {
	if state == nil || state.campaignObj == nil || state.scheduleAt == nil {
		return ErrCampaignDispatchScheduleRequired
	}

	rt := intruntime.FromClient(ctx, withTransactionalMutation(ctx))
	if rt == nil {
		return ErrCampaignDispatchRuntimeRequired
	}

	req, err := r.buildCampaignEmailDispatchRequest(ctx, state.campaignObj, state.resend, state.includeOverdue, "", state.scheduleAt)
	if err != nil {
		return err
	}

	if _, err := rt.Dispatch(ctx, req); err != nil {
		logx.FromContext(ctx).Error().Err(err).
			Str("campaign_id", state.campaignObj.ID).
			Msg("failed scheduling campaign dispatch")

		return err
	}

	return nil
}


// ensureCampaignEditAccess verifies the caller can edit the campaign.
func (r *mutationResolver) ensureCampaignEditAccess(ctx context.Context, campaignID string) error {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return auth.ErrNoAuthUser
	}

	if caller.Has(auth.CapSystemAdmin) {
		return nil
	}

	if caller.SubjectID == "" {
		return parseRequestError(ctx, auth.ErrNoAuthUser, common.Action{Action: common.ActionGet, Object: "user"})
	}

	allow, err := r.db.Authz.CheckAccess(ctx, fgax.AccessCheck{
		Relation:    fgax.CanEdit,
		ObjectType:  generated.TypeCampaign,
		ObjectID:    campaignID,
		SubjectType: caller.SubjectType(),
		SubjectID:   caller.SubjectID,
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error checking campaign edit access")

		return err
	}
	if !allow {
		logx.FromContext(ctx).Warn().Str("campaign_id", campaignID).Msg("access denied to edit campaign")

		return newPermissionDeniedError()
	}
	return nil
}
