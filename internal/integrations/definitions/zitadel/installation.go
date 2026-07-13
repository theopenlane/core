package zitadel

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// resolveInstallationMetadata derives Zitadel instance metadata from the persisted credential
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	domain, ok := resolveDomain(req.Credentials)
	if !ok {
		logx.FromContext(ctx).Error().Msg("missing domain in credentials")
		return InstallationMetadata{}, false, nil
	}

	return InstallationMetadata{
		Domain: domain,
	}, true, nil
}