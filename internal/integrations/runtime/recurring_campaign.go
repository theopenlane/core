package runtime

import (
	"context"
	"encoding/json"
	"time"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/campaign"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// SeedRecurringCampaigns starts the durable recurring campaign polling loop
// after runtime listeners have been registered.
func (r *Runtime) SeedRecurringCampaigns(ctx context.Context) error {
	receipt := r.Gala().EmitWithHeaders(ctx, operations.RecurringCampaignTopic, operations.RecurringCampaignEnvelope{}, gala.Headers{})

	return receipt.Err
}

// HandleRecurringCampaigns queries all due recurring campaigns and dispatches
// each one through the email integration runtime. Returns the number of
// campaigns dispatched as the delta for adaptive scheduling
func (r *Runtime) HandleRecurringCampaigns(ctx context.Context, _ operations.RecurringCampaignEnvelope) (int, error) {
	db := r.DB()
	now := time.Now()
	systemCtx := auth.WithCaller(privacy.DecisionContext(ctx, privacy.Allow), &auth.Caller{
		Capabilities: auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation,
	})

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
		if err := r.dispatchRecurringCampaign(systemCtx, camp, now); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("campaign_id", camp.ID).Str("owner_id", camp.OwnerID).Msg("failed dispatching recurring campaign")

			continue
		}

		dispatched++
	}

	return dispatched, nil
}

// dueCampaignPredicates returns the where clauses for campaigns eligible for
// recurring dispatch: recurring, active status, next_run_at <= now, and
// optionally bounded by recurrence_end_at
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

// dispatchRecurringCampaign dispatches a single recurring campaign and advances
// its scheduling fields
func (r *Runtime) dispatchRecurringCampaign(ctx context.Context, camp *ent.Campaign, now time.Time) error {
	input := emaildef.CampaignDispatchInput{
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
		operation = emaildef.SendQuestionnaireCampaignOp.Name()
		config, err = json.Marshal(emaildef.SendQuestionnaireCampaignRequest{
			CampaignDispatchInput: input,
		})
	default:
		operation = emaildef.SendCampaignOp.Name()
		config, err = json.Marshal(emaildef.SendBrandedCampaignRequest{
			CampaignDispatchInput: input,
		})
	}

	if err != nil {
		return err
	}

	integrationID, err := r.ResolveOwnerIntegration(ctx, emaildef.DefinitionID.ID(), camp.OwnerID, func(inst *ent.Integration) bool {
		return inst.CampaignEmail
	})
	if err != nil {
		return err
	}

	req := operations.DispatchRequest{
		IntegrationID: integrationID,
		DefinitionID:  emaildef.DefinitionID.ID(),
		OwnerID:       camp.OwnerID,
		Operation:     operation,
		Config:        config,
		RunType:       enums.IntegrationRunTypeScheduled,
		Runtime:       integrationID == "",
	}

	if _, err := r.Dispatch(ctx, req); err != nil {
		return err
	}

	nextRun := operations.NextCampaignRunAt(now, camp.RecurrenceFrequency, camp.RecurrenceInterval, camp.RecurrenceTimezone)
	nowDT := models.DateTime(now)
	exhausted := camp.RecurrenceEndAt != nil && !camp.RecurrenceEndAt.IsZero() && !nextRun.Before(time.Time(*camp.RecurrenceEndAt))

	update := r.DB().Campaign.UpdateOneID(camp.ID).
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
