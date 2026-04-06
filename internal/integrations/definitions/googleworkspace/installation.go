package googleworkspace

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// resolveInstallationMetadata derives Google Workspace installation metadata from installation user input
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	var input UserInput
	if err := jsonx.UnmarshalIfPresent(req.Config.ClientConfig, &input); err != nil {
		return InstallationMetadata{}, false, err
	}

	if input.AdminEmail == "" && input.CustomerID == "" && input.Domain == "" {
		return InstallationMetadata{}, false, nil
	}

	return InstallationMetadata{
		AdminEmail: input.AdminEmail,
		CustomerID: input.CustomerID,
		Domain:     input.Domain,
	}, true, nil
}
