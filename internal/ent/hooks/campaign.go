package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
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
