package aws

import (
	"github.com/theopenlane/core/internal/integrations/providers/awskit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const awsDefaultSession = "openlane-aws"

// awsMetadataFromCredential extracts and validates AWS metadata from a credential set
func awsMetadataFromCredential(credential types.CredentialSet, defaultSessionName string) (awskit.Metadata, error) {
	if len(credential.ProviderData) == 0 {
		return awskit.Metadata{}, ErrMetadataMissing
	}

	parsed, err := awskit.MetadataFromProviderData(credential.ProviderData, defaultSessionName)
	if err != nil {
		return awskit.Metadata{}, err
	}

	if parsed.RoleARN == "" {
		return awskit.Metadata{}, ErrRoleARNMissing
	}

	if parsed.Region == "" {
		return awskit.Metadata{}, ErrRegionMissing
	}

	return parsed, nil
}
