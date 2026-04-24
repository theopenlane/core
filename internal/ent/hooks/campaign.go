package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
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

			integrationID := emaildef.ResolveEmailIntegration(ctx, m.Client(), ownerID)
			if integrationID != "" {
				m.SetIntegrationID(integrationID)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}
