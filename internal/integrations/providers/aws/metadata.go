package aws

import (
	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const awsDefaultSession = "openlane-aws"

func awsMetadataFromPayload(payload types.CredentialPayload, defaultSessionName string) (auth.AWSMetadata, error) {
	meta := payload.Data.ProviderData
	if len(meta) == 0 {
		return auth.AWSMetadata{}, ErrMetadataMissing
	}

	parsed, err := auth.AWSMetadataFromProviderData(meta, defaultSessionName)
	if err != nil {
		return auth.AWSMetadata{}, err
	}
	if parsed.RoleARN == "" {
		return auth.AWSMetadata{}, ErrRoleARNMissing
	}
	if parsed.Region == "" {
		return auth.AWSMetadata{}, ErrRegionMissing
	}

	return parsed, nil
}
