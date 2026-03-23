package awssecurityhub

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck holds the result of an AWS Security Hub health check
type HealthCheck struct {
	// Region is the AWS region used for the session
	Region string `json:"region"`
	// RoleARN is the assumed role ARN when present
	RoleARN string `json:"roleArn,omitempty"`
	// HubARN is the Security Hub ARN
	HubARN string `json:"hubArn,omitempty"`
	// SubscribedAt is the Security Hub subscription timestamp
	SubscribedAt string `json:"subscribedAt,omitempty"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClientRequest(
		SecurityHubClient,
		func(ctx context.Context, request types.OperationRequest, client *securityhub.Client) (json.RawMessage, error) {
			return h.Run(ctx, request.Credentials, client)
		},
	)
}

// Run validates Security Hub access by calling DescribeHub
func (HealthCheck) Run(ctx context.Context, credentials types.CredentialBindings, c *securityhub.Client) (json.RawMessage, error) {
	awsCredential, err := resolveAssumeRoleCredential(credentials)
	if err != nil {
		return nil, err
	}

	resp, err := c.DescribeHub(ctx, &securityhub.DescribeHubInput{})
	if err != nil {
		return nil, ErrDescribeHubFailed
	}

	details := HealthCheck{
		Region:  awsCredential.HomeRegion,
		RoleARN: awsCredential.RoleARN,
	}

	if resp.HubArn != nil {
		details.HubARN = *resp.HubArn
	}

	if resp.SubscribedAt != nil {
		details.SubscribedAt = *resp.SubscribedAt
	}

	return providerkit.EncodeResult(details, ErrResultEncode)
}
