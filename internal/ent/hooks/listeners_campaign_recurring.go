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
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaCampaignRecurringListeners registers mutation listeners that
// manage recurring campaign scheduling when is_active or is_recurring changes
func RegisterGalaCampaignRecurringListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	return eventqueue.RegisterMutationListeners(g,
		eventqueue.MutationListener{
			Schema: entgen.TypeCampaign,
			Name:   "campaign.recurring.schedule_sync",
			Operations: []string{
				ent.OpUpdate.String(),
				ent.OpUpdateOne.String(),
			},
			Fields: []string{
				campaign.FieldIsActive,
				campaign.FieldIsRecurring,
				campaign.FieldRecurrenceFrequency,
				campaign.FieldRecurrenceInterval,
			},
			Handle: handleCampaignRecurringMutation,
		},
	)
}

// handleCampaignRecurringMutation reacts to scheduling-relevant field changes
// on campaigns to keep next_run_at consistent with the desired state
func handleCampaignRecurringMutation(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
	activationChanged := eventqueue.MutationFieldChanged(payload, campaign.FieldIsActive) ||
		eventqueue.MutationFieldChanged(payload, campaign.FieldIsRecurring)
	scheduleShapeChanged := eventqueue.MutationFieldChanged(payload, campaign.FieldRecurrenceFrequency) ||
		eventqueue.MutationFieldChanged(payload, campaign.FieldRecurrenceInterval)

	caller, ok := auth.CallerFromContext(inv.Context)
	if !ok || caller == nil {
		return nil
	}

	camp, err := inv.Client.Campaign.Query().
		Where(
			campaign.ID(inv.EntityID),
			campaign.OwnerIDEQ(caller.OrganizationID),
		).
		Select(
			campaign.FieldIsActive,
			campaign.FieldIsRecurring,
			campaign.FieldRecurrenceFrequency,
			campaign.FieldRecurrenceInterval,
			campaign.FieldRecurrenceTimezone,
			campaign.FieldNextRunAt,
			campaign.FieldLastRunAt,
			campaign.FieldStatus,
		).
		Only(inv.Context)
	if err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Str("campaign_id", inv.EntityID).Msg("failed loading campaign for recurring schedule sync")

		return err
	}

	shouldSchedule := camp.IsRecurring && camp.IsActive && !isTerminalStatus(camp.Status) && camp.RecurrenceFrequency != enums.FrequencyNone

	switch {
	case shouldSchedule && (activationChanged || camp.NextRunAt == nil):
		return recomputeNextRunAt(inv.Context, inv.Client, camp)
	case shouldSchedule && scheduleShapeChanged:
		return recomputeNextRunAt(inv.Context, inv.Client, camp)
	case !shouldSchedule && camp.NextRunAt != nil:
		return clearNextRunAt(inv.Context, inv.Client, camp.ID)
	default:
		return nil
	}
}

// recomputeNextRunAt computes next_run_at from the last run (or now if never run) and persists it
func recomputeNextRunAt(ctx context.Context, client *entgen.Client, camp *entgen.Campaign) error {
	base := time.Now()
	if camp.LastRunAt != nil {
		base = time.Time(*camp.LastRunAt)
	}

	nextRun := camp.RecurrenceFrequency.NextOccurrence(base, camp.RecurrenceInterval, camp.RecurrenceTimezone)

	if !nextRun.After(time.Now()) {
		nextRun = time.Now()
	}

	if err := client.Campaign.UpdateOneID(camp.ID).
		SetNextRunAt(models.DateTime(nextRun)).
		Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed setting next_run_at on reactivation")
		return err
	}

	logx.FromContext(ctx).Debug().Time("next_run_at", nextRun).Msg("recurring campaign schedule recomputed")

	return nil
}

// clearNextRunAt removes next_run_at when a campaign is deactivated or no longer recurring
func clearNextRunAt(ctx context.Context, client *entgen.Client, campaignID string) error {
	if err := client.Campaign.UpdateOneID(campaignID).
		ClearNextRunAt().
		Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed clearing next_run_at on deactivation")
		return err
	}

	logx.FromContext(ctx).Debug().Msg("recurring campaign schedule cleared")

	return nil
}

// isTerminalStatus reports whether a campaign status prevents future dispatch
func isTerminalStatus(status enums.CampaignStatus) bool {
	return lo.Contains([]enums.CampaignStatus{
		enums.CampaignStatusCompleted,
		enums.CampaignStatusCanceled,
	}, status)
}

// campaignScheduleContext grants the internal operation capability to the restored caller so the
// recurrence schedule can be read and written without the caller's own object permissions
func campaignScheduleContext(ctx context.Context) context.Context {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		caller = &auth.Caller{}
	}

	return auth.WithCaller(ctx, caller.WithCapabilities(auth.CapInternalOperation))
}

// withCampaignLogContext sets the campaign and organization ID in the log context
func withCampaignLogContext(ctx context.Context, campaignID, orgID string) context.Context {
	return logx.WithFields(ctx, map[string]any{
		"campaign_id":     campaignID,
		"organization_id": orgID,
	})
}
