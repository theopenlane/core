package awssecurityhub

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// resolveInstallationMetadata derives AWS connection metadata from whichever credential is bound.
// It uses the assume-role credential when present, otherwise falls back to the service account credential.
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	_, hasAssumeRole := req.Credentials.Resolve(awsAssumeRoleCredential.ID())
	if !hasAssumeRole {
		return InstallationMetadata{}, true, nil
	}

	assumeRole, ok, err := awsAssumeRoleCredential.Resolve(req.Credentials)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("awssecurityhub: error resolving assume role credentials")
		return InstallationMetadata{}, false, ErrCredentialMetadataInvalid
	}

	if !ok {
		return InstallationMetadata{}, ok, nil
	}

	return InstallationMetadata{
		RoleARN:       assumeRole.RoleARN,
		HomeRegion:    assumeRole.HomeRegion,
		AccountID:     assumeRole.AccountID,
		AccountScope:  assumeRole.AccountScope,
		AccountIDs:    assumeRole.AccountIDs,
		LinkedRegions: assumeRole.LinkedRegions,
	}, true, nil
}
