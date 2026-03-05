package aws

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// ClientAWSSecurityHub identifies the AWS Security Hub client descriptor.
	ClientAWSSecurityHub types.ClientName = "securityhub"
)

// awsSecurityHubClientDescriptors returns the AWS Security Hub client descriptors for pooling.
func awsSecurityHubClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeAWS, ClientAWSSecurityHub, "AWS Security Hub client", pooledSecurityHubClient)
}

// pooledSecurityHubClient builds the AWS Security Hub client for pooling, discarding metadata.
func pooledSecurityHubClient(ctx context.Context, payload types.CredentialPayload, _ json.RawMessage) (types.ClientInstance, error) {
	client, _, err := buildSecurityHubClient(ctx, payload)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return types.NewClientInstance(client), nil
}
