package awssecurityhub

import (
	"context"

	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientAWSSecurityHub identifies the AWS Security Hub client descriptor.
	ClientAWSSecurityHub types.ClientName = "securityhub"
)

// awsSecurityHubClientDescriptors returns the AWS Security Hub client descriptors for pooling.
func awsSecurityHubClientDescriptors() []types.ClientDescriptor {
	return []types.ClientDescriptor{
		{
			Provider:     TypeAWSSecurityHub,
			Name:         ClientAWSSecurityHub,
			Description:  "AWS Security Hub client",
			Build:        buildAWSSecurityHubClient,
			ConfigSchema: map[string]any{"type": "object"},
		},
	}
}

// buildAWSSecurityHubClient builds the AWS Security Hub client for pooling.
func buildAWSSecurityHubClient(ctx context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	client, _, err := buildSecurityHubClient(ctx, payload)
	if err != nil {
		return nil, err
	}

	return client, nil
}
