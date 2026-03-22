package awssecurityhub

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// resolveInstallationMetadata derives AWS connection metadata from the persisted assume-role configuration
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	assumeRoleCredential, ok := req.Credentials.Resolve(awsAssumeRoleCredential)
	if !ok {
		return InstallationMetadata{}, false, nil
	}

	var assumeRole AssumeRoleCredentialSchema
	if err := jsonx.UnmarshalIfPresent(assumeRoleCredential.Data, &assumeRole); err != nil {
		return InstallationMetadata{}, false, ErrCredentialMetadataInvalid
	}

	if assumeRole.RoleARN == "" && assumeRole.HomeRegion == "" && assumeRole.AccountID == "" {
		return InstallationMetadata{}, false, nil
	}

	_, hasSourceCredential := req.Credentials.Resolve(awsSourceCredential)

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
