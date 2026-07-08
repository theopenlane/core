package zitadel

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// resolveInstallationMetadata derives Zitadel instance metadata from the persisted credential
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error resolving credentials")
		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if cred.Domain == "" || cred.Token == "" {
		logx.FromContext(ctx).Error().Msg("missing domain or token in credentials")
		return InstallationMetadata{}, false, nil
	}

	return InstallationMetadata{
		Domain: cred.Domain,
	}, true, nil
}