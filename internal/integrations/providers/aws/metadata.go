package aws

import (
	awskit "github.com/theopenlane/core/internal/integrations/providers/awskit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const awsDefaultSession = "openlane-aws"

// awsMetadataFromPayload extracts and validates AWS metadata from a credential payload
func awsMetadataFromPayload(payload types.CredentialPayload, defaultSessionName string) (awskit.AWSMetadata, error) {
	meta := payload.Data.ProviderData
	if len(meta) == 0 {
		return awskit.AWSMetadata{}, ErrMetadataMissing
	}

	parsed, err := awskit.AWSMetadataFromProviderData(meta, defaultSessionName)
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
