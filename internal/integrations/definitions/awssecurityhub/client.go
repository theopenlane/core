package awssecurityhub

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"

	"github.com/theopenlane/core/internal/integrations/definitions/awskit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const defaultSessionName = "openlane-awssecurityhub"

// Client builds AWS Security Hub clients for one installation
type Client struct{}

// Build constructs the AWS Security Hub client using STS AssumeRole
func (Client) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	if len(req.Credential.ProviderData) == 0 {
		return nil, ErrCredentialMetadataRequired
	}

	meta, err := awskit.MetadataFromProviderData(req.Credential.ProviderData, defaultSessionName)
	if err != nil {
		return nil, fmt.Errorf("awssecurityhub: metadata decode failed: %w", err)
	}

	if meta.RoleARN == "" {
		return nil, ErrRoleARNMissing
	}

	if meta.Region == "" {
		return nil, ErrRegionMissing
	}

	cfg, err := awskit.BuildAWSConfig(ctx, meta.Region, awskit.CredentialsFromMetadata(meta), awskit.AssumeRole{
		RoleARN:         meta.RoleARN,
		ExternalID:      meta.ExternalID,
		SessionName:     meta.SessionName,
		SessionDuration: meta.SessionDuration,
	})
	if err != nil {
		return nil, fmt.Errorf("awssecurityhub: aws config build failed: %w", err)
	}

	return securityhub.NewFromConfig(cfg), nil
}

// FromAny casts a registered client instance to the Security Hub client type
func (Client) FromAny(value any) (*securityhub.Client, error) {
	c, ok := value.(*securityhub.Client)
	if !ok {
		return nil, ErrClientType
	}

	return c, nil
}
