package cloudflare

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// resolveInstallationMetadata derives Cloudflare account metadata from installation user input
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	var input UserInput
	if err := jsonx.UnmarshalIfPresent(req.Config.ClientConfig, &input); err != nil {
		return InstallationMetadata{}, false, err
	}

	if input.AccountID == "" {
		return InstallationMetadata{}, false, nil
	}

	return InstallationMetadata{
		AccountID: input.AccountID,
	}, true, nil
}
