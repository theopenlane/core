package awssecurityhub

import (
	"context"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientAWSSecurityHub identifies the AWS Security Hub client descriptor.
	ClientAWSSecurityHub types.ClientName = "securityhub"
)

// awsSecurityHubClientDescriptors returns the AWS Security Hub client descriptors for pooling.
func awsSecurityHubClientDescriptors() []types.ClientDescriptor {
	return helpers.DefaultClientDescriptors(TypeAWSSecurityHub, ClientAWSSecurityHub, "AWS Security Hub client", buildAWSSecurityHubClient)
}

// buildAWSSecurityHubClient builds the AWS Security Hub client for pooling.
func buildAWSSecurityHubClient(ctx context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	client, _, err := buildSecurityHubClient(ctx, payload)
	if err != nil {
		return nil, err
	}

	return client, nil
}
