package hooks

import (
	"context"
	"encoding/json"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/integrations/operations"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/pkg/logx"
)

// HookCampaignResolveEmailIntegration resolves the email integration for the
// campaign's owner organization and sets integration_id on create. It queries
// integrations matching the email definition and prefers one flagged with
// campaign_email=true. If no integration is found the field is left empty and
// dispatch will fall back to the runtime provider
func HookCampaignResolveEmailIntegration() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.CampaignFunc(func(ctx context.Context, m *generated.CampaignMutation) (generated.Value, error) {
			ownerID, _ := m.OwnerID()
			if ownerID == "" {
				return next.Mutate(ctx, m)
			}

			integrationID := resolveEmailIntegrationForOwner(ctx, m.Client(), ownerID)
			if integrationID != "" {
				m.SetIntegrationID(integrationID)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

// resolveEmailIntegrationForOwner queries for email integrations belonging to
// the given owner and returns the best match: the one flagged with
// campaign_email=true, or the sole match when only one exists
func resolveEmailIntegrationForOwner(ctx context.Context, client *generated.Client, ownerID string) string {
	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	integrations, err := client.Integration.Query().
		Where(
			integration.OwnerIDEQ(ownerID),
			integration.DefinitionIDEQ(emaildef.DefinitionID.ID()),
		).
		All(systemCtx)
	if err != nil || len(integrations) == 0 {
		return ""
	}

	if len(integrations) == 1 {
		return integrations[0].ID
	}

	// multiple email integrations — prefer the one flagged for campaigns
	for _, inst := range integrations {
		if inst.CampaignEmail {
			return inst.ID
		}
	}

	return ""
}

// HookCampaignDispatchOnActive dispatches the send-campaign operation via the
// integration runtime when a campaign's status transitions to ACTIVE.
// It reads the campaign's integration_id edge; when set, the dispatch targets
// that specific installation. When empty, the runtime provider handles it
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

			config, err := json.Marshal(emaildef.SendCampaignRequest{CampaignID: campaignID})
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("campaign_id", campaignID).Msg("failed marshaling campaign dispatch config")
				return v, nil
			}

			integrationID, _ := m.IntegrationID()

			dispatchCampaignSend(ctx, rt, campaignID, integrationID, config)

			return v, nil
		})
	}, ent.OpUpdateOne|ent.OpUpdate)
}

// dispatchCampaignSend dispatches the send-campaign operation. When an
// integration ID is provided, it dispatches through that installation.
// Otherwise it falls back to runtime execution
func dispatchCampaignSend(ctx context.Context, rt *intruntime.Runtime, campaignID string, integrationID string, config json.RawMessage) {
	if integrationID != "" {
		if _, err := rt.Dispatch(ctx, operations.DispatchRequest{
			IntegrationID: integrationID,
			Operation:     emaildef.SendCampaignOp.Name(),
			Config:        config,
			RunType:       enums.IntegrationRunTypeEvent,
		}); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("campaign_id", campaignID).Str("integration_id", integrationID).Msg("failed dispatching campaign send operation")
		}

		return
	}

	// no integration linked — fall back to runtime execution
	if _, err := rt.ExecuteRuntimeOperation(ctx, emaildef.DefinitionID.ID(), emaildef.SendCampaignOp.Name(), config); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", campaignID).Msg("failed executing campaign send via runtime")
	}
}
