package email

import (
	"context"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// ResolveCampaignEmailIntegration returns a connected email integration for
// campaign dispatch, preferring an explicitly linked campaign integration, then
// the owner-level campaign_email provider, then the sole connected email install.
func ResolveCampaignEmailIntegration(ctx context.Context, client *generated.Client, ownerID string, preferredIntegrationID string) string {
	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	query := client.Integration.Query().
		Where(
			integration.OwnerIDEQ(ownerID),
			integration.DefinitionIDEQ(DefinitionID.ID()),
			integration.StatusEQ(enums.IntegrationStatusConnected),
		)

	if preferredIntegrationID != "" {
		preferred, err := query.Clone().Where(integration.IDEQ(preferredIntegrationID)).Only(systemCtx)
		if err == nil {
			return preferred.ID
		}
	}

	integrations, err := query.All(systemCtx)
	if err != nil || len(integrations) == 0 {
		return ""
	}

	for _, inst := range integrations {
		if inst.CampaignEmail {
			return inst.ID
		}
	}

	if len(integrations) == 1 {
		return integrations[0].ID
	}

	return ""
}
