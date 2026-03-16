package awsassets

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/theopenlane/core/internal/integrations/definitions/awskit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const awsDefaultSessionName = "openlane-awsassets"

// Client builds AWS STS clients for one AWS Assets installation
type Client struct{}

// Build constructs the AWS STS client using STS AssumeRole
func (Client) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	meta, err := awskit.MetadataFromProviderData(req.Credential.ProviderData, awsDefaultSessionName)
	if err != nil {
		return nil, ErrCredentialMetadataInvalid
	}

	if meta.RoleARN == "" {
		return nil, ErrRoleARNMissing
	}

	cfg, err := awskit.BuildAWSConfig(ctx, meta.Region, awskit.CredentialsFromMetadata(meta), awskit.AssumeRole{
		RoleARN:         meta.RoleARN,
		ExternalID:      meta.ExternalID,
		SessionName:     meta.SessionName,
		SessionDuration: meta.SessionDuration,
	})
	if err != nil {
		return nil, ErrAWSConfigBuildFailed
	}

	return sts.NewFromConfig(cfg), nil
}

// FromAny casts a registered client instance to the STS client type
func (Client) FromAny(value any) (*sts.Client, error) {
	c, ok := value.(*sts.Client)
	if !ok {
		return nil, ErrClientType
	}

	return c, nil
}
