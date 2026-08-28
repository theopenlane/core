package hooks

import (
	"context"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	entgen "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/campaign"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// CampaignRecurringListeners keeps recurring campaign scheduling in sync when activation or recurrence fields change
func CampaignRecurringListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaCampaign,
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Fields: []string{
				campaign.FieldIsActive,
				campaign.FieldIsRecurring,
				campaign.FieldRecurrenceFrequency,
				campaign.FieldRecurrenceInterval,
			},
			Caller: func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
				return restored.WithCapabilities(auth.CapInternalOperation)
			},
			Handle: handleCampaignRecurringMutation,
		},
	}
}

// handleCampaignRecurringMutation reacts to scheduling-relevant field changes
// on campaigns to keep next_run_at consistent with the desired state
func handleCampaignRecurringMutation(inv entityops.Invocation, payload entityops.MutationPayload) error {
	activationChanged := payload.FieldChanged(campaign.FieldIsActive) ||
		payload.FieldChanged(campaign.FieldIsRecurring)
	scheduleShapeChanged := payload.FieldChanged(campaign.FieldRecurrenceFrequency) ||
		payload.FieldChanged(campaign.FieldRecurrenceInterval)

	camp, err := inv.Client.Campaign.Query().
		Where(
			campaign.ID(inv.EntityID),
			campaign.OwnerIDEQ(inv.Caller.OrganizationID),
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
		if entgen.IsNotFound(err) {
			logx.FromContext(inv.Context).Info().Msg("failed to find campaign, campaign may have been deleted before running")

			// nothing to do if the campaign can no longer be found
			return nil
		}

		logx.FromContext(inv.Context).Error().Err(err).Msg("failed loading campaign for recurring schedule sync")

		return err
	}

	shouldSchedule := camp.IsRecurring && camp.IsActive && !isTerminalStatus(camp.Status) && camp.RecurrenceFrequency != enums.FrequencyNone

	switch {
	case shouldSchedule && (activationChanged || scheduleShapeChanged || camp.NextRunAt == nil):
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
