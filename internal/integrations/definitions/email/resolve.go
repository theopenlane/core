package email

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
)

// ResolveCampaignEmailIntegration returns a connected email integration for
// campaign dispatch, returning the one flagged CampaignEmail, then falling back
// to the sole connected email install
func ResolveCampaignEmailIntegration(ctx context.Context, client *generated.Client, ownerID string) (string, error) {
	integrations, err := client.Integration.Query().
		Where(
			integration.OwnerIDEQ(ownerID),
			integration.DefinitionIDEQ(DefinitionID.ID()),
			integration.StatusEQ(enums.IntegrationStatusConnected),
		).All(ctx)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrIntegrationNotFound, err)
	}

	for _, inst := range integrations {
		if inst.CampaignEmail {
			return inst.ID, nil
		}
	}

	if len(integrations) == 1 {
		return integrations[0].ID, nil
	}

	return "", ErrIntegrationNotFound
}
