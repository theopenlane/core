package graphapi

import (
	"context"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/integrations/definitions/googledrive"
	"github.com/theopenlane/core/internal/integrations/definitions/onedrive"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// driveIntegration pairs a connected integration record with its definition ID
type driveIntegration struct {
	Integration  *generated.Integration
	DefinitionID string
}

// findPrimaryDriveIntegration returns the primary Google Drive or OneDrive integration for the given org.
// An installation with Primary=true is preferred; otherwise the first connected installation is returned.
func (r *internalPolicyResolver) findPrimaryDriveIntegration(ctx context.Context, ownerID string) (*driveIntegration, error) {
	integrations, err := withTransactionalMutation(ctx).Integration.Query().
		Where(
			integration.OwnerID(ownerID),
			integration.DefinitionIDIn(googledrive.DefinitionID(), onedrive.DefinitionID()),
			integration.StatusEQ(enums.IntegrationStatusConnected),
		).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to query drive integrations")
		return nil, err
	}

	for _, integ := range integrations {
		primary := isPrimaryDriveInstallation(integ)
		if primary {
			return &driveIntegration{Integration: integ, DefinitionID: integ.DefinitionID}, nil
		}
	}

	if len(integrations) > 0 {
		return &driveIntegration{Integration: integrations[0], DefinitionID: integrations[0].DefinitionID}, nil
	}

	return nil, nil
}

// isPrimaryDriveInstallation reports whether the installation's client config has Primary set to true
func isPrimaryDriveInstallation(integ *generated.Integration) bool {
	switch integ.DefinitionID {
	case googledrive.DefinitionID():
		var input googledrive.UserInput
		if err := jsonx.UnmarshalIfPresent(integ.Config.ClientConfig, &input); err != nil {
			return false
		}

		return input.Primary
	case onedrive.DefinitionID():
		var input onedrive.UserInput
		if err := jsonx.UnmarshalIfPresent(integ.Config.ClientConfig, &input); err != nil {
			return false
		}

		return input.Primary
	}

	return false
}
