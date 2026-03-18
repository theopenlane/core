package aws

import (
	"context"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientAWSSecurityHub identifies the AWS Security Hub client descriptor.
	ClientAWSSecurityHub types.ClientName = "securityhub"
)

// awsSecurityHubClientDescriptors returns the AWS Security Hub client descriptors for pooling.
func awsSecurityHubClientDescriptors() []types.ClientDescriptor {
	return auth.DefaultClientDescriptors(TypeAWS, ClientAWSSecurityHub, "AWS Security Hub client", pooledSecurityHubClient)
}

// pooledSecurityHubClient builds the AWS Security Hub client for pooling, discarding metadata.
func pooledSecurityHubClient(ctx context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	client, _, err := buildSecurityHubClient(ctx, payload)
	return client, err
}
