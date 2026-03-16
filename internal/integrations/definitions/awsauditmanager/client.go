package awsauditmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/auditmanager"

	"github.com/theopenlane/core/internal/integrations/definitions/awskit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const defaultSessionName = "openlane-awsauditmanager"

// Client builds AWS Audit Manager clients for one installation
type Client struct{}

// Build constructs the AWS Audit Manager client using STS AssumeRole
func (Client) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	if len(req.Credential.ProviderData) == 0 {
		return nil, ErrCredentialMetadataRequired
	}

	meta, err := awskit.MetadataFromProviderData(req.Credential.ProviderData, defaultSessionName)
	if err != nil {
		return nil, ErrCredentialMetadataInvalid
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
		return nil, ErrAWSConfigBuildFailed
	}

	return auditmanager.NewFromConfig(cfg), nil
}

// FromAny casts a registered client instance to the Audit Manager client type
func (Client) FromAny(value any) (*auditmanager.Client, error) {
	c, ok := value.(*auditmanager.Client)
	if !ok {
		return nil, ErrClientType
	}

	return c, nil
}
