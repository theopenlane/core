package awssecurityhub

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
)

// resolveInstallationMetadata derives AWS connection metadata from the persisted assume-role configuration
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	assumeRole, ok, err := awsAssumeRoleCredential.Resolve(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, ErrCredentialMetadataInvalid
	}

	if !ok {
		return InstallationMetadata{}, false, nil
	}

	if assumeRole.RoleARN == "" && assumeRole.HomeRegion == "" && assumeRole.AccountID == "" {
		return InstallationMetadata{}, false, nil
	}

	_, hasSourceCredential, _ := awsServiceAccountCredential.Resolve(req.Credentials)

	return InstallationMetadata{
		RoleARN:              assumeRole.RoleARN,
		HomeRegion:           assumeRole.HomeRegion,
		AccountID:            assumeRole.AccountID,
		AccountScope:         assumeRole.AccountScope,
		AccountIDs:           assumeRole.AccountIDs,
		LinkedRegions:        assumeRole.LinkedRegions,
		UsesSourceCredential: hasSourceCredential,
	}, true, nil
}
