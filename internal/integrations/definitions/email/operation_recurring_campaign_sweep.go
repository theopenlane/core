package email

import (
	"context"
	"encoding/json"
	"time"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/campaign"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// RecurringCampaignSweep configures one recurring campaign sweep cycle
type RecurringCampaignSweep struct{}

var recurringCampaignSweepSchema, RecurringCampaignOp = providerkit.OperationSchema[RecurringCampaignSweep]() //nolint:revive

// Handle adapts the recurring campaign sweep to the generic operation registration boundary
func (r RecurringCampaignSweep) Handle() types.OperationHandler {
	return func(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
		processed, err := r.Run(ctx, req)
		if err != nil {
			return nil, err
		}

		return providerkit.EncodeResult(types.ScheduledCycleResult{Processed: processed}, ErrResultEncode)
	}
}

// Run executes one recurring campaign sweep and returns the number of campaigns dispatched
func (RecurringCampaignSweep) Run(ctx context.Context, req types.OperationRequest) (int, error) {
	db := req.DB
	now := time.Now()
	systemCtx := systemSweepContext(ctx)

	campaigns, err := db.Campaign.Query().
		Where(dueCampaignPredicates(now)...).
		All(systemCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed querying due recurring campaigns")
		return 0, err
	}

	if len(campaigns) == 0 {
		return 0, nil
	}

	logx.FromContext(ctx).Info().Int("count", len(campaigns)).Msg("recurring campaigns due for dispatch")

	dispatched := 0

	for _, camp := range campaigns {
		if err := dispatchRecurringCampaign(systemCtx, req, camp, now); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("campaign_id", camp.ID).Str("owner_id", camp.OwnerID).Msg("failed dispatching recurring campaign")

			continue
		}

		dispatched++
	}

	return dispatched, nil
}

// dueCampaignPredicates returns the where clauses for campaigns eligible for recurring dispatch
func dueCampaignPredicates(now time.Time) []predicate.Campaign {
	return []predicate.Campaign{
		campaign.IsRecurring(true),
		campaign.IsActive(true),
		campaign.StatusNotIn(enums.CampaignStatusCompleted, enums.CampaignStatusCanceled, enums.CampaignStatusDraft),
		campaign.NextRunAtNotNil(),
		campaign.NextRunAtLTE(models.DateTime(now)),
		campaign.Or(
			campaign.RecurrenceEndAtIsNil(),
			campaign.RecurrenceEndAtGT(models.DateTime(now)),
		),
	}
}

// dispatchRecurringCampaign dispatches a single recurring campaign and advances its scheduling fields
func dispatchRecurringCampaign(ctx context.Context, req types.OperationRequest, camp *ent.Campaign, now time.Time) error {
	input := CampaignDispatchInput{
		CampaignID: camp.ID,
		Resend:     true,
	}

	var (
		operation string
		config    []byte
		err       error
	)

	switch camp.CampaignType {
	case enums.CampaignTypeQuestionnaire:
		operation = SendQuestionnaireCampaignOp.Name()
		config, err = json.Marshal(SendQuestionnaireCampaignRequest{
			CampaignDispatchInput: input,
		})
	default:
		operation = SendCampaignOp.Name()
		config, err = json.Marshal(SendBrandedCampaignRequest{
			CampaignDispatchInput: input,
		})
	}

	if err != nil {
		return err
	}

	integrationID, err := operations.ResolveOwnerIntegration(ctx, req.DB, DefinitionID.ID(), camp.OwnerID, func(inst *ent.Integration) bool {
		return inst.CampaignEmail
	})
	if err != nil {
		return err
	}

	if _, err := req.Dispatch(ctx, types.DispatchRequest{
		IntegrationID: integrationID,
		DefinitionID:  DefinitionID.ID(),
		OwnerID:       camp.OwnerID,
		Operation:     operation,
		Config:        config,
		RunType:       enums.IntegrationRunTypeScheduled,
		Runtime:       integrationID == "",
	}); err != nil {
		return err
	}

	nextRun := operations.NextCampaignRunAt(now, camp.RecurrenceFrequency, camp.RecurrenceInterval, camp.RecurrenceTimezone)
	nowDT := models.DateTime(now)
	exhausted := camp.RecurrenceEndAt != nil && !camp.RecurrenceEndAt.IsZero() && !nextRun.Before(time.Time(*camp.RecurrenceEndAt))

	update := req.DB.Campaign.UpdateOneID(camp.ID).
		SetLastRunAt(nowDT)

	switch {
	case exhausted:
		update.ClearNextRunAt().
			SetStatus(enums.CampaignStatusCompleted).
			SetIsActive(false).
			SetCompletedAt(nowDT)
	default:
		update.SetNextRunAt(models.DateTime(nextRun))
	}

	if err := update.Exec(privacy.DecisionContext(ctx, privacy.Allow)); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", camp.ID).Msg("failed updating recurring campaign schedule")

		return err
	}

	return nil
}
