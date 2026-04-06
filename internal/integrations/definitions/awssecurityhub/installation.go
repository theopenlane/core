package awssecurityhub

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
)

// resolveInstallationMetadata derives AWS connection metadata from whichever credential is bound.
// Either an assume-role credential or a service account credential may be configured independently
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	_, hasAssumeRole := req.Credentials.Resolve(awsAssumeRoleCredential.ID())
	if !hasAssumeRole {
		return InstallationMetadata{}, false, nil
	}

	assumeRole, _, err := awsAssumeRoleCredential.Resolve(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, ErrCredentialMetadataInvalid
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
