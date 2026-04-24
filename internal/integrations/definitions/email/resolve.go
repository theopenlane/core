package email

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// ResolveEmailIntegration queries for email integrations belonging to the given
// owner and returns the best match: the one flagged with campaign_email=true, or
// the sole match when only one exists. Returns empty string when no integration
// is found, allowing dispatch to fall back to the runtime provider
func ResolveEmailIntegration(ctx context.Context, client *generated.Client, ownerID string) string {
	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	integrations, err := client.Integration.Query().
		Where(
			integration.OwnerIDEQ(ownerID),
			integration.DefinitionIDEQ(DefinitionID.ID()),
		).
		All(systemCtx)
	if err != nil || len(integrations) == 0 {
		return ""
	}

	if len(integrations) == 1 {
		return integrations[0].ID
	}

	for _, inst := range integrations {
		if inst.CampaignEmail {
			return inst.ID
		}
	}

	return ""
}
