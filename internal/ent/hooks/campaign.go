package hooks

import (
	"context"
	"encoding/json"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/pkg/logx"
)

// HookCampaignDispatchOnActive dispatches the send-campaign operation via the
// integration runtime when a campaign's status transitions to ACTIVE
func HookCampaignDispatchOnActive() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.CampaignFunc(func(ctx context.Context, m *generated.CampaignMutation) (generated.Value, error) {
			status, exists := m.Status()
			if !exists || status != enums.CampaignStatusActive {
				return next.Mutate(ctx, m)
			}

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			rt := intruntime.FromClient(ctx, m.Client())
			if rt == nil {
				return v, nil
			}

			campaignID, _ := m.ID()

			ownerID, _ := m.OwnerID()

			config, err := json.Marshal(emaildef.SendCampaignRequest{CampaignID: campaignID})
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("campaign_id", campaignID).Msg("failed marshaling campaign dispatch config")
				return v, nil
			}

			if err := rt.DispatchForOwner(ctx, emaildef.DefinitionID(), emaildef.SendCampaignOpName(), ownerID, config); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("campaign_id", campaignID).Msg("failed dispatching campaign send operation")
			}

			return v, nil
		})
	}, ent.OpUpdateOne|ent.OpUpdate)
}
