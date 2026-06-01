package graphapi

import (
	"context"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/integrations/definitions/googledrive"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// findPrimaryDriveIntegration finds the Drive integration marked as primary for the given org
func (r *internalPolicyResolver) findPrimaryDriveIntegration(ctx context.Context, ownerID string) (*generated.Integration, error) {
	integrations, err := withTransactionalMutation(ctx).Integration.Query().
		Where(
			integration.OwnerID(ownerID),
			integration.DefinitionID(googledrive.DefinitionID()),
			integration.StatusEQ(enums.IntegrationStatusConnected),
		).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to query drive integrations")
		return nil, err
	}

	for _, integ := range integrations {
		var input googledrive.UserInput
		if err := jsonx.UnmarshalIfPresent(integ.Config.ClientConfig, &input); err != nil {
			continue
		}

		if input.Primary {
			return integ, nil
		}
	}

	if len(integrations) > 0 {
		return integrations[0], nil
	}

	return nil, nil
}
