package graphapi

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/campaign"
	"github.com/theopenlane/core/internal/ent/generated/campaigntarget"
	"github.com/theopenlane/core/internal/ent/hooks"
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
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "control"})
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

	if campaignObj.CampaignType == enums.CampaignTypeQuestionnaire && campaignObj.AssessmentID == "" {
		return ErrCampaignMissingAssessmentID
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
		if !campaignTargetDispatchable(target.Status, state.resend, state.includeOverdue) {
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

// dispatchCampaignOperation dispatches the appropriate email operation for the
// campaign type through the integration runtime. Questionnaire campaigns use
// SendQuestionnaireCampaignOp; all others use SendBrandedCampaignOp
func (r *mutationResolver) dispatchCampaignOperation(ctx context.Context, state *campaignDispatchState) error {
	rt := intruntime.FromClient(ctx, withTransactionalMutation(ctx))
	if rt == nil {
		return nil
	}

	var operationName string

	switch state.campaignObj.CampaignType {
	case enums.CampaignTypeQuestionnaire:
		operationName = emaildef.SendQuestionnaireCampaignOp.Name()
	default:
		operationName = emaildef.SendBrandedCampaignOp.Name()
	}

	config, err := json.Marshal(struct {
		CampaignID string `json:"campaignId"`
	}{CampaignID: state.campaignObj.ID})
	if err != nil {
		return err
	}

	integrationID := state.campaignObj.IntegrationID
	if integrationID != "" {
		if _, err := rt.Dispatch(ctx, operations.DispatchRequest{
			IntegrationID: integrationID,
			Operation:     operationName,
			Config:        config,
			RunType:       enums.IntegrationRunTypeEvent,
		}); err != nil {
			logx.FromContext(ctx).Error().Err(err).
				Str("campaign_id", state.campaignObj.ID).
				Str("integration_id", integrationID).
				Msg("failed dispatching campaign operation via customer integration")
		}

		return nil
	}

	if _, err := rt.ExecuteRuntimeOperation(ctx, emaildef.DefinitionID.ID(), operationName, config); err != nil {
		logx.FromContext(ctx).Error().Err(err).
			Str("campaign_id", state.campaignObj.ID).
			Msg("failed executing campaign operation via runtime")
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
	if err := r.enqueueCampaignDispatchJob(ctx, state.campaignObj, state.opts.Action, state.scheduleAt); err != nil {
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

	//	r.recordCampaignDispatchEvent(ctx, state.campaignObj, state.opts.Action, state.queuedCount, state.skippedCount, state.scheduleAt)

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

// campaignTargetDispatchable determines whether a target should be dispatched.
func campaignTargetDispatchable(status enums.AssessmentResponseStatus, resend bool, includeOverdue bool) bool {
	switch status {
	case enums.AssessmentResponseStatusCompleted:
		return false
	case enums.AssessmentResponseStatusOverdue:
		return includeOverdue || resend
	case enums.AssessmentResponseStatusSent:
		return resend
	default:
		return true
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

// enqueueCampaignDispatchJob inserts the campaign dispatch job into the queue.
func (r *mutationResolver) enqueueCampaignDispatchJob(ctx context.Context, campaignObj *generated.Campaign, action jobspec.CampaignDispatchAction, scheduledAt *time.Time) error {
	if campaignObj == nil || scheduledAt == nil {
		return ErrCampaignDispatchScheduleRequired
	}

	args := jobspec.CampaignDispatchArgs{
		CampaignID:  campaignObj.ID,
		Action:      action,
		ScheduledAt: scheduledAt,
	}

	if caller, ok := auth.CallerFromContext(ctx); ok && caller != nil && caller.SubjectID != "" {
		args.RequestedBy = caller.SubjectID
	}

	opts := args.InsertOpts()
	opts.ScheduledAt = *scheduledAt

	_, err := r.db.Job.Insert(ctx, args, &opts)
	if err != nil {
		if hooks.IsUniqueConstraintError(err) {
			logx.FromContext(ctx).Info().Err(err).Msg("campaign dispatch job already scheduled")

			return nil
		}

		return err
	}

	return nil
}

// recordCampaignDispatchEvent emits an audit event for the dispatch request.
func (r *mutationResolver) recordCampaignDispatchEvent(ctx context.Context, campaignObj *generated.Campaign, action jobspec.CampaignDispatchAction, queuedCount, skippedCount int, scheduleAt *time.Time) {
	if campaignObj == nil {
		return
	}

	eventType := "campaign." + strings.ToLower(string(action))
	if scheduleAt != nil {
		eventType += ".scheduled"
	}

	metadata := map[string]any{
		"campaign_id":   campaignObj.ID,
		"action":        string(action),
		"queued_count":  queuedCount,
		"skipped_count": skippedCount,
	}
	if scheduleAt != nil {
		metadata["scheduled_at"] = scheduleAt.UTC().Format(time.RFC3339Nano)
	}

	input := generated.CreateEventInput{
		EventType:       eventType,
		Metadata:        metadata,
		OrganizationIDs: []string{campaignObj.OwnerID},
	}

	if caller, ok := auth.CallerFromContext(ctx); ok && caller != nil && caller.SubjectID != "" {
		input.UserIDs = []string{caller.SubjectID}
	}

	if err := withTransactionalMutation(ctx).Event.Create().SetInput(input).Exec(ctx); err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("failed to record campaign dispatch event")
	}
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
