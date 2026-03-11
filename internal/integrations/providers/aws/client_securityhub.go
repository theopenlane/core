package aws

import (
	"context"
	"encoding/json"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers/awskit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// ClientAWSSecurityHub identifies the AWS Security Hub client descriptor
	ClientAWSSecurityHub types.ClientName = "securityhub"
)

// awsSecurityHubClientDescriptors returns the AWS Security Hub client descriptors for pooling
func awsSecurityHubClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeAWS, ClientAWSSecurityHub, "AWS Security Hub client", pooledSecurityHubClient)
}

// pooledSecurityHubClient builds the AWS Security Hub client for pooling, discarding metadata
func pooledSecurityHubClient(ctx context.Context, credential types.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
	client, _, err := buildSecurityHubClient(ctx, credential)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return types.NewClientInstance(client), nil
}

// newSecurityHubSDKClient wraps securityhub.NewFromConfig for use with generic helpers
func newSecurityHubSDKClient(cfg awssdk.Config) *securityhub.Client {
	return securityhub.NewFromConfig(cfg)
}

// buildSecurityHubClient builds a Security Hub client from stored credentials
func buildSecurityHubClient(ctx context.Context, credential types.CredentialSet) (*securityhub.Client, awskit.Metadata, error) {
	return buildAWSClient(ctx, credential, newSecurityHubSDKClient)
}

// resolveSecurityHubClient returns a pooled client when supplied or builds one on demand
func resolveSecurityHubClient(ctx context.Context, input types.OperationInput) (*securityhub.Client, awskit.Metadata, error) {
	return resolveAWSClient(ctx, input, newSecurityHubSDKClient)
}
