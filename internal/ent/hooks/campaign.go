package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/emailruntime"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/logx"
)

// HookCampaignSendEmails dispatches campaign emails when a campaign transitions to ACTIVE status.
// The hook fires after the mutation completes, only when email_template_id is present on the result.
// Send errors are logged and not propagated; a delivery failure must not roll back the status change.
func HookCampaignSendEmails() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.CampaignFunc(func(ctx context.Context, m *generated.CampaignMutation) (generated.Value, error) {
			status, statusSet := m.Status()
			if !statusSet || status != enums.CampaignStatusActive {
				return next.Mutate(ctx, m)
			}

			// only fire on transition to ACTIVE; skip if the record is already ACTIVE
			if !m.Op().Is(ent.OpCreate) {
				if old, err := m.OldStatus(ctx); err == nil && old == enums.CampaignStatusActive {
					return next.Mutate(ctx, m)
				}
			}

			val, err := next.Mutate(ctx, m)
			if err != nil {
				return val, err
			}

			camp, ok := val.(*generated.Campaign)
			if !ok || camp == nil || camp.EmailTemplateID == "" {
				return val, nil
			}

			if sendErr := emailruntime.SendCampaignEmails(ctx, m.Client(), camp.ID); sendErr != nil {
				logx.FromContext(ctx).Error().Err(sendErr).
					Str("campaign_id", camp.ID).
					Msg("failed dispatching campaign emails on activation")
			}

			return val, nil
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}
