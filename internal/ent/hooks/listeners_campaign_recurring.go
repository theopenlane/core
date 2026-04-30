package hooks

import (
	"context"
	"time"

	"entgo.io/ent"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/campaign"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaCampaignRecurringListeners registers mutation listeners that
// manage recurring campaign scheduling when is_active or is_recurring changes
func RegisterGalaCampaignRecurringListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry,
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic: eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeCampaign),
			Name:  "campaign.recurring.schedule_sync",
			Operations: []string{
				ent.OpUpdate.String(),
				ent.OpUpdateOne.String(),
			},
			Handle: handleCampaignRecurringMutation,
		},
	)
}

// handleCampaignRecurringMutation reacts to is_active and is_recurring field
// changes on campaigns to keep next_run_at consistent with the desired state
func handleCampaignRecurringMutation(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	activeChanged := eventqueue.MutationFieldChanged(payload, campaign.FieldIsActive)
	recurringChanged := eventqueue.MutationFieldChanged(payload, campaign.FieldIsRecurring)

	if !activeChanged && !recurringChanged {
		return nil
	}

	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	campaignID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || campaignID == "" {
		return nil
	}

	caller, ok := auth.CallerFromContext(ctx.Context)
	if !ok || caller == nil {
		return nil
	}

	camp, err := client.Campaign.Query().
		Where(
			campaign.ID(campaignID),
			campaign.OwnerIDEQ(caller.OrganizationID),
		).
		Select(
			campaign.FieldIsActive,
			campaign.FieldIsRecurring,
			campaign.FieldRecurrenceFrequency,
			campaign.FieldRecurrenceInterval,
			campaign.FieldRecurrenceTimezone,
			campaign.FieldNextRunAt,
			campaign.FieldStatus,
		).
		Only(ctx.Context)
	if err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Str("campaign_id", campaignID).Msg("failed loading campaign for recurring schedule sync")

		return err
	}

	shouldSchedule := camp.IsRecurring && camp.IsActive && !isTerminalStatus(camp.Status) && camp.RecurrenceFrequency != enums.FrequencyNone

	switch {
	case shouldSchedule && camp.NextRunAt == nil:
		return recomputeNextRunAt(ctx.Context, client, camp)
	case !shouldSchedule && camp.NextRunAt != nil:
		return clearNextRunAt(ctx.Context, client, camp.ID)
	default:
		return nil
	}
}

// recomputeNextRunAt computes next_run_at from now and persists it
func recomputeNextRunAt(ctx context.Context, client *entgen.Client, camp *entgen.Campaign) error {
	now := time.Now()
	nextRun := operations.NextCampaignRunAt(now, camp.RecurrenceFrequency, camp.RecurrenceInterval, camp.RecurrenceTimezone)

	if err := client.Campaign.UpdateOneID(camp.ID).
		SetNextRunAt(models.DateTime(nextRun)).
		Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", camp.ID).Msg("failed setting next_run_at on reactivation")
		return err
	}

	logx.FromContext(ctx).Debug().Str("campaign_id", camp.ID).Time("next_run_at", nextRun).Msg("recurring campaign schedule recomputed")

	return nil
}

// clearNextRunAt removes next_run_at when a campaign is deactivated or no longer recurring
func clearNextRunAt(ctx context.Context, client *entgen.Client, campaignID string) error {
	if err := client.Campaign.UpdateOneID(campaignID).
		ClearNextRunAt().
		Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", campaignID).Msg("failed clearing next_run_at on deactivation")
		return err
	}

	logx.FromContext(ctx).Debug().Str("campaign_id", campaignID).Msg("recurring campaign schedule cleared")

	return nil
}

// isTerminalStatus reports whether a campaign status prevents future dispatch
func isTerminalStatus(status enums.CampaignStatus) bool {
	return lo.Contains([]enums.CampaignStatus{
		enums.CampaignStatusCompleted,
		enums.CampaignStatusCanceled,
	}, status)
}
