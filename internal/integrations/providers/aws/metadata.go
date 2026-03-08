package aws

import (
	"github.com/theopenlane/core/common/models"
	awskit "github.com/theopenlane/core/internal/integrations/providers/awskit"
)

const awsDefaultSession = "openlane-aws"

// awsMetadataFromPayload extracts and validates AWS metadata from a credential payload
func awsMetadataFromPayload(payload models.CredentialSet, defaultSessionName string) (awskit.AWSMetadata, error) {
	if len(payload.ProviderData) == 0 {
		return awskit.AWSMetadata{}, ErrMetadataMissing
	}

	parsed, err := awskit.AWSMetadataFromProviderData(payload.ProviderData, defaultSessionName)
	if err != nil {
		return awskit.AWSMetadata{}, err
	}
	if parsed.RoleARN == "" {
		return awskit.AWSMetadata{}, ErrRoleARNMissing
	}
	if parsed.Region == "" {
		return awskit.AWSMetadata{}, ErrRegionMissing
	}

	return parsed, nil
}
